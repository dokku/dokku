"""Sequential browser oracles with neutral readiness and post-case HTTP checks."""
import argparse
import hashlib
import html
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import time
import urllib.error
import urllib.request
import uuid
import xml.etree.ElementTree as ET

from importlib.util import module_from_spec, spec_from_file_location

ROOT = Path(__file__).resolve().parents[1]
REPO = ROOT.parents[1]
GATING = (1, 2, 3, 4, 5, 6, 7, 8)


def read_http(path):
  try:
    with urllib.request.urlopen('http://localhost:19080' + path, timeout=10) as r:
      return r.status, r.read().decode()
  except urllib.error.HTTPError as r:
    return r.code, r.read().decode()


def verdict(report, exit_code):
  tree = ET.parse(report)
  cases = tree.findall('.//testcase')
  if len(cases) != 1 or tree.findall('.//error') or tree.findall('.//skipped'):
    raise RuntimeError('Expected exactly one completed browser verdict')
  run_id = next((e.get('value') for e in tree.iter() if e.get('name') == 'taskId'), None)
  if not run_id:
    raise RuntimeError('Missing cloud run ID')
  failures = tree.findall('.//failure')
  passed = exit_code == 0 and not failures
  if not passed and not failures:
    raise RuntimeError('CLI failed without a browser assertion verdict')
  return passed, run_id, [e.get('message', '') for e in failures]


def agree(browser, observed):
  if browser != observed:
    raise RuntimeError('oracle disagreement: browser and post-case cross-check differ')


class Controller:
  def __init__(self, artifacts):
    self.artifacts = artifacts
    artifacts.mkdir(parents=True, exist_ok=False)
    self.token = os.environ['TESTRIGOR_API_KEY']
    self.suite = os.environ['TESTRIGOR_SUITE_ID']
    if not self.token.strip() or not re.fullmatch(r'[A-Za-z0-9_-]+', self.suite):
      raise ValueError('Nonempty CI token and valid suite ID required')
    self.local_env = {k: v for k, v in os.environ.items() if k != 'TESTRIGOR_API_KEY'}
    self.cli_env = dict(self.local_env, TESTRIGOR_API_KEY=self.token,
                        TESTRIGOR_CLI_HOME=str(artifacts / 'cli-home'))
    self.cli = str(ROOT / 'node_modules/.bin/testrigor')
    self.compose = ['docker', 'compose', '-f', str(ROOT / 'compose.yml')]
    self.results = []
    self.redirect_control_route = False

  def command(self, args):
    return subprocess.check_output(args, cwd=REPO, env=self.local_env,
                                   text=True, encoding='utf-8', stderr=subprocess.PIPE).strip()

  def dokku(self, *args):
    return self.command(self.compose + ['exec', '-T', 'dokku', 'dokku', *args])

  def hostname(self):
    # Inspect the named active container, not the response under test or retiring containers.
    return self.command(self.compose + ['exec', '-T', 'dokku', 'docker', 'inspect',
                        'browser-probe.api.1', '--format', '{{.Config.Hostname}}'])

  def neutral_ready(self):
    deadline = time.monotonic() + 120
    while time.monotonic() < deadline:
      try:
        if read_http('/healthz') == (200, 'web ready'):
          self.command(self.compose + ['exec', '-T', 'dokku', 'nginx', '-t'])
          return
      except OSError:
        pass
      time.sleep(1)
    raise RuntimeError('Neutral web readiness timed out')

  def route(self, present=True, strip=False):
    if present:
      self.dokku('proxy:route:set', 'browser-probe', 'api', '/api/v0', '--port', '5001',
                 *(['--strip-prefix'] if strip else []))
    else:
      self.dokku('proxy:route:remove', 'browser-probe', '/api/v0')

  def cli_process(self, args, log):
    code = None
    with log.open('w', encoding='utf-8') as stream:
      process = subprocess.Popen([self.cli, *args], cwd=REPO, env=self.cli_env,
        stdin=subprocess.DEVNULL, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
        text=True, encoding='utf-8', errors='replace')
      for line in process.stdout:
        safe = line.replace(self.token, '[REDACTED]')
        stream.write(safe); stream.flush(); print(safe, end='', flush=True)
      code = process.wait()
    return code

  def browser(self, number, variables, negative):
    folder = self.artifacts / ('negative' if negative else 'positive') / f'TC-{number:02d}'
    folder.mkdir(parents=True)
    values = folder / 'variables.json'
    values.write_text(json.dumps(variables), encoding='utf-8')
    case, = ROOT.glob(f'test-cases/TC-{number:02d} *.yaml')
    base = ['test-suite', 'validate-syntax', self.suite, '--test-cases-path', str(case), '--explicit-mutations']
    if self.cli_process(base, folder / 'validation.log'):
      raise RuntimeError('Syntax validation failed; no browser verdict')
    args = ['test-suite', 'run', self.suite, '--localhost', '--url', 'http://localhost:19080',
            '--test-cases-path', str(case), '--variables-path', str(values),
            '--settings-path', str(ROOT / 'settings.json'),
            '--labels', f'routing-tc-{number:02d}',
            '--branch', os.environ['DOKKU_HARNESS_BRANCH'], '--commit', os.environ['DOKKU_SOURCE_REVISION'],
            '--min-port', '48196', '--max-port', '48196',
            '--junit-report-save-path', str(folder / 'report.xml')]
    code = self.cli_process(args, folder / 'cli.log')
    passed, run_id, failures = verdict(folder / 'report.xml', code)
    result = {'case': f'TC-{number:02d}', 'negativeControl': negative, 'browserPassed': passed,
              'exitCode': code, 'runID': run_id, 'failureMessages': failures}
    self.results.append(result)
    return passed, result

  def crosscheck(self, number, values):
    if number == 6:
      spec = spec_from_file_location('transport', Path(__file__).with_name('transport-probe.py'))
      module = module_from_spec(spec); spec.loader.exec_module(module)
      try:
        return module.websocket()['marker'] == values['expectedBackendMarker']
      except RuntimeError:
        return False
    status, body = read_http('/api/v0/redirect' if number == 8 else '/api/v0/identity')
    if number == 4:
      return status == 404 and body == 'Wrong upstream or path'
    if number == 7:
      return status == 502 and all(values[key] in html.unescape(body) for key in ('expectedErrorHeading', 'expectedErrorDetail'))
    if number == 8:
      return status == 200 and 'API login page' in body
    if status != 200:
      return False
    try:
      data = json.loads(body)
    except ValueError:
      return False
    return (isinstance(data, dict) and data.get('role') == 'api'
            and data.get('marker') == values['expectedBackendMarker']
            and (number != 2 or (bool(data.get('hostname')) and data['hostname'] != values['previousHostname']))
            and (number != 3 or data['marker'] != values['previousBackendMarker']))

  def stage(self, number, negative):
    values = {'expectedBackendMarker': self.original}
    if number == 1:
      self.route(not negative)
    elif number == 2:
      self.route()
      values['previousHostname'] = self.hostname()
      if not negative:
        self.dokku('ps:restart', 'browser-probe')
    elif number == 3:
      self.route()
      values.update(previousBackendMarker=self.original, expectedBackendMarker='dokku-backend-' + str(uuid.uuid4()))
      if not negative:
        self.dokku('config:set', 'browser-probe', 'E2E_MARKER=' + values['expectedBackendMarker'])
    elif number == 4:
      self.dokku('config:set', 'browser-probe', 'E2E_MARKER=' + self.original)
      self.route(negative)
    elif number == 5:
      self.route(False)
      if not negative:
        self.route()
    elif number == 6:
      if negative:
        self.dokku('proxy:route:remove', 'browser-probe', '/ws')
      else:
        self.dokku('proxy:route:set', 'browser-probe', 'ws', '/ws', '--port', '5002')
    elif number == 7:
      self.route()
      page = (REPO / 'plugins/nginx-vhosts/templates/500-error.html').read_text()
      values.update(expectedErrorHeading=html.unescape(re.search(r'<h1>(.*?)</h1>', page)[1]),
                    expectedErrorDetail=html.unescape(re.search(r'<p>(.*?)</p>', page)[1]))
      self.dokku('ps:scale', 'browser-probe', 'api=1' if negative else 'api=0')
    elif number == 8:
      self.dokku('ps:scale', 'browser-probe', 'api=1')
      self.route(strip=True)
      if negative:
        # Strip the entire endpoint instead of the base prefix: API root is 200
        # with the wrong heading, so the browser must reach its heading assertion.
        self.redirect_control_route = True
        self.dokku('proxy:route:set', 'browser-probe', 'api', '/api/v0/redirect', '--port', '5001', '--strip-prefix')
    self.neutral_ready()
    passed, result = self.browser(number, values, negative)
    observed = self.crosscheck(number, values)
    result['postCaseCrosscheck'] = observed
    agree(passed, observed)
    if negative and passed:
      raise RuntimeError('Negative control unexpectedly passed')
    if negative and not any('check that' in message.lower() for message in result['failureMessages']):
      raise RuntimeError('Negative control did not reach its intended assertion')
    if not negative and not passed:
      raise RuntimeError('Browser assertion failed: ' + result['case'])

  def execute(self, numbers, negative=False):
    revision = os.environ['DOKKU_SOURCE_REVISION']
    if self.command(['git', 'rev-parse', 'HEAD']) != revision:
      raise RuntimeError('Checkout differs from requested source revision')
    if self.command(self.compose + ['exec', '-T', 'dokku', 'cat', '/var/lib/dokku/GIT_REV']) != revision:
      raise RuntimeError('Installed Dokku revision differs from checkout')
    if self.dokku('config:get', 'browser-probe', 'E2E_FIXTURE_OWNER') != 'dokku-browser-e2e':
      raise RuntimeError('Refusing to change an unowned app')
    self.original = self.dokku('config:get', 'browser-probe', 'E2E_MARKER')
    self.original_routes = self.dokku('proxy:route:report', 'browser-probe')
    if not self.original.startswith('dokku-backend-') or '/api/v0 -> api:5001' not in self.original_routes or '/ws -> ws:5002' not in self.original_routes:
      raise RuntimeError('Unexpected fixture marker or original routes')
    if negative and 8 in numbers and 'Route /api/v0/redirect ->' in self.original_routes:
      raise RuntimeError('Refusing to overwrite an existing redirect-control route')
    external_prefix = self.dokku('config:get', 'browser-probe', 'E2E_EXTERNAL_PREFIX') if 8 in numbers else None
    if 8 in numbers and external_prefix != '/api/v0':
      raise RuntimeError('TC-08 requires the prefix-aware fixture; run the installer first')
    expected = json.loads((ROOT / 'package.json').read_text())['dependencies']['@testrigor/testrigor-cli']
    version = self.command([self.cli, '--version'])
    if '/' + expected + ' ' not in version:
      raise RuntimeError('Installed CLI differs from package.json')
    image = json.loads(self.command(['docker', 'image', 'inspect', os.environ['DOKKU_E2E_IMAGE'], '--format', '{{json .}}']))
    provenance = {'sourceRevision': revision, 'harnessRevision': revision, 'suiteID': self.suite,
      'externalPrefix': external_prefix,
      'cliVersion': version, 'runtimeImageID': image['Id'],
      'requestedMaximumRetries': json.loads((ROOT / 'settings.json').read_text())['maximumNumerOfRetries'],
      'hashes': {str(p.relative_to(REPO)): hashlib.sha256(p.read_bytes()).hexdigest()
                 for p in ROOT.rglob('*') if 'node_modules' not in p.parts and '__pycache__' not in p.parts and p.is_file()}}
    (self.artifacts / 'provenance.json').write_text(json.dumps(provenance, indent=2))
    passed = False
    cleanup = []
    try:
      for number in numbers:
        self.stage(number, negative)
      passed = True
    finally:
      for args in [('ps:scale', 'browser-probe', 'api=1', 'ws=1'),
                   ('config:set', 'browser-probe', 'E2E_MARKER=' + self.original),
                   ('proxy:route:set', 'browser-probe', 'api', '/api/v0', '--port', '5001'),
                   ('proxy:route:set', 'browser-probe', 'ws', '/ws', '--port', '5002'),
                   *([('proxy:route:remove', 'browser-probe', '/api/v0/redirect')] if self.redirect_control_route else [])]:
        try:
          self.dokku(*args)
        except Exception as error:
          cleanup.append(type(error).__name__)
      try:
        self.neutral_ready()
        status, body = read_http('/api/v0/identity')
        if status != 200 or json.loads(body)['marker'] != self.original or self.dokku('proxy:route:report', 'browser-probe') != self.original_routes:
          cleanup.append('Original state was not restored')
        if 8 in numbers and self.dokku('config:get', 'browser-probe', 'E2E_EXTERNAL_PREFIX') != external_prefix:
          cleanup.append('Fixture external prefix changed')
      except Exception as error:
        cleanup.append(type(error).__name__)
      (self.artifacts / 'result.json').write_text(json.dumps({'passed': passed and not cleanup,
        'stages': self.results, 'cleanupFailures': cleanup}, indent=2))
      if cleanup:
        raise RuntimeError('Fixture cleanup failed')


def execute(artifacts, numbers=GATING, negative=False):
  Controller(artifacts).execute(numbers, negative)


if __name__ == '__main__':
  parser = argparse.ArgumentParser()
  parser.add_argument('artifacts', type=Path)
  parser.add_argument('--cases', type=int, nargs='+', choices=GATING, default=GATING)
  parser.add_argument('--negative-controls', action='store_true')
  args = parser.parse_args()
  sys.stdout.reconfigure(encoding='utf-8', errors='replace')
  try:
    execute(args.artifacts, args.cases, args.negative_controls)
  except Exception as error:
    # Only controlled messages; never print subprocess/HTTP objects or env values.
    print('Routing run failed: ' + (str(error) if isinstance(error, RuntimeError) else type(error).__name__), file=sys.stderr)
    sys.exit(1)
