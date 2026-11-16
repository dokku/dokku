"""Publish direct execution links from completed JUnit reports, without credentials."""
import argparse
import json
import os
from pathlib import Path
import re
import xml.etree.ElementTree as ET


def execution_reports(artifacts):
  reports = []
  for mode in ('positive', 'negative'):
    for path in sorted(artifacts.glob(mode + '/TC-*/report.xml')):
      tree = ET.parse(path)
      if len(tree.findall('.//testcase')) != 1:
        raise ValueError('Expected one completed case per execution report')
      properties = {e.get('name'): e.get('value') for e in tree.iter() if e.get('name') and e.get('value')}
      suite, run = properties.get('testSuiteId', ''), properties.get('taskId', '')
      case = path.parent.name
      if (not re.fullmatch(r'TC-\d{2}', case)
          or not all(re.fullmatch(r'[A-Za-z0-9_-]+', value) for value in (suite, run))):
        raise ValueError('A completed report is missing a valid suite or run identifier')
      if tree.findall('.//error'):
        status = 'error'
      elif tree.findall('.//skipped'):
        status = 'skipped'
      elif tree.findall('.//failure'):
        status = 'failed'
      else:
        status = 'passed'
      reports.append({'case': case, 'mode': mode, 'status': status, 'runID': run,
        'url': f'https://app.testrigor.com/public/{run}'})
  return reports


def markdown(reports):
  lines = ['## testRigor execution reports', '']
  if not reports:
    lines.append('No browser execution report was produced. Check the build and setup logs.')
  else:
    lines.extend(['| Case | Mode | Browser result | Report |', '| --- | --- | --- | --- |'])
    lines.extend(f"| {r['case']} | {r['mode']} | {r['status']} | [View execution]({r['url']}) |" for r in reports)
    lines.extend(['', 'Links open the public testRigor execution viewer.'])
  return '\n'.join(lines) + '\n'


if __name__ == '__main__':
  parser = argparse.ArgumentParser()
  parser.add_argument('artifacts', type=Path)
  args = parser.parse_args()
  reports = execution_reports(args.artifacts)
  content = markdown(reports)
  args.artifacts.mkdir(parents=True, exist_ok=True)
  (args.artifacts / 'execution-links.md').write_text(content, encoding='utf-8')
  (args.artifacts / 'execution-links.json').write_text(json.dumps(reports, indent=2), encoding='utf-8')
  if os.environ.get('GITHUB_STEP_SUMMARY'):
    with Path(os.environ['GITHUB_STEP_SUMMARY']).open('a', encoding='utf-8') as stream:
      stream.write(content)
  print(content, end='')
