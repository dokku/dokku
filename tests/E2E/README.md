# Browser routing tests

This optional testRigor suite checks nginx path routing to non-web processes
through a source-built Dokku instance. Eight cases cover backend identity,
restart recovery, configuration rollout, route removal and re-addition,
WebSocket traffic, the scaled-zero error page and prefix-aware redirects.
See [COVERAGE.md](COVERAGE.md) for assertions and overlap with native tests.

## Run locally

Requires an x86-64 Linux host, Bash, Docker Engine with Compose, Python 3,
Git and the Node version in `.node-version`. Docker must support privileged
containers. Run these commands from the repository root.

Dokku runs locally; browser execution runs on testRigor's service through
the CLI's localhost tunnel. An account, a dedicated Desktop Web Chrome suite
and network access to testRigor and build dependencies are required.
Create a suite at [testRigor](https://app.testrigor.com), copy its ID from
`/test-suites/<ID>/test-cases`, and obtain an authentication token from your
account's API Tokens page. Reserve the suite for this harness and run one
controller at a time: a new cloud run can cancel an existing suite run.

The builder uses committed files from `git archive`, not working-tree edits.
Commit your changes locally before building. Keep the default app name
`browser-probe` and published ports; the controller expects HTTP port 19080.

```bash
npm ci --prefix tests/E2E
export TESTRIGOR_SUITE_ID="<your-suite-id>"
read -rsp "testRigor API token: " TESTRIGOR_API_KEY; echo
export TESTRIGOR_API_KEY
export DOKKU_SOURCE_REVISION=$(git rev-parse HEAD)
export DOKKU_E2E_IMAGE=dokku-e2e/source:${DOKKU_SOURCE_REVISION:0:9}
export DOKKU_HARNESS_BRANCH=$(git branch --show-current)
export E2E_ARTIFACTS=$(mktemp -d)

bash tests/E2E/scripts/build-dokku.sh "$E2E_ARTIFACTS"
bash tests/E2E/scripts/install-dokku.sh install
python3 tests/E2E/scripts/run-ci.py "$E2E_ARTIFACTS/positive"
python3 tests/E2E/scripts/report-results.py "$E2E_ARTIFACTS/positive"
```

The controller supplies per-case variables and requests zero failed-test retries
through `settings.json`. No Test Data entries need to be entered manually.
Run commands omit `--explicit-mutations`; syntax validation uses it to validate
only the supplied case. Reserve a dedicated suite for this harness.

The installer deploys web, API and WebSocket processes into a separate Docker
daemon without mounting the host Docker socket. Ports bind only to loopback.
The controller checks source revision and app ownership before making changes,
then restores and verifies fixture state in a `finally` block.
Compose teardown preserves named volumes for later runs.

Optional development checks:

```bash
# A selected case; the controller prepares its required state.
python3 tests/E2E/scripts/run-ci.py "$E2E_ARTIFACTS/redirect" --cases 8
# Each wrong-state control must fail its browser assertion.
python3 tests/E2E/scripts/run-ci.py "$E2E_ARTIFACTS/negative" --negative-controls
# Offline controller and reporting checks; no account or Docker required.
python3 tests/E2E/scripts/test_ci.py
```

Stop the isolated deployment when finished, including after a failed run:

```bash
docker compose -f tests/E2E/compose.yml down --remove-orphans
unset TESTRIGOR_API_KEY
```

Installer actions are `install`, `status`, `probe`, `logs`, `restart` and
`reset`. Reset recreates the owned fixture app and preserves daemon volumes.
To test a different commit, rebuild its image and run the installer again.

## CI configuration

Set repository variable `TESTRIGOR_SUITE_ID` and secret `TESTRIGOR_CI_TOKEN`.
The secret supplies the same authentication token used locally.
Eligible same-repository pull requests run when routing or harness files change;
manual dispatch and a weekly default-branch schedule are also supported.
Missing configuration and fork pull requests skip with a notice.
Concurrency prevents overlapping workflow runs against the same suite.
Each case has a distinct `routing-tc-01` through `routing-tc-08` label. The
controller selects that label so each execution contains one case even when
all eight definitions are saved in the online suite.

The workflow builds the checked-out source, installs the fixture, runs all eight
cases sequentially and stops the isolated deployment even on failure.
Checkout credentials are not persisted, actions are pinned and permissions
are read-only. The token is passed only to CLI subprocesses and redacted from
their logs; it is not passed into Docker.

JUnit reports, CLI logs, source provenance and diagnostics are retained for
seven days. The job summary and uploaded Markdown/JSON indexes link completed
executions through testRigor's public viewer. Report visibility depends on the
suite's sharing configuration. Runs that fail during setup have no execution link.
