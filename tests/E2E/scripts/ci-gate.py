"""Evaluate CI opt-in without starting Docker, contacting testRigor or logging secrets."""
import os
from pathlib import Path


def evaluate(env):
  manual = env.get('GITHUB_EVENT_NAME') == 'workflow_dispatch'
  scheduled = (env.get('GITHUB_EVENT_NAME') == 'schedule'
               and bool(env.get('DOKKU_DEFAULT_BRANCH'))
               and env.get('GITHUB_REF') == 'refs/heads/' + env['DOKKU_DEFAULT_BRANCH'])
  pull_request = (env.get('GITHUB_EVENT_NAME') == 'pull_request'
                  and bool(env.get('GITHUB_REPOSITORY'))
                  and env.get('DOKKU_PR_HEAD_REPOSITORY') == env['GITHUB_REPOSITORY'])
  trusted = manual or scheduled or pull_request
  configured = bool(env.get('TESTRIGOR_SUITE_ID', '').strip()) and env.get('TESTRIGOR_TOKEN_PRESENT') == 'true'
  return trusted and configured


def reason(env):
  if not env.get('TESTRIGOR_SUITE_ID', '').strip():
    return 'suite ID is not configured'
  if env.get('TESTRIGOR_TOKEN_PRESENT') != 'true':
    return 'CI token is not configured'
  return 'event is not a same-repository PR, manual dispatch or default-branch schedule'


if __name__ == '__main__':
  enabled = evaluate(os.environ)
  value = 'true' if enabled else 'false'
  with Path(os.environ['GITHUB_OUTPUT']).open('a', encoding='utf-8') as output:
    output.write('enabled=' + value + '\n')
  message = 'Browser E2E enabled.' if enabled else 'Browser E2E skipped: ' + reason(os.environ)
  print(message if enabled else '::notice::' + message)
  if os.environ.get('GITHUB_STEP_SUMMARY'):
    with Path(os.environ['GITHUB_STEP_SUMMARY']).open('a', encoding='utf-8') as summary:
      summary.write(message + '\n')
