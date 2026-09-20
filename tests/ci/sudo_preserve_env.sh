#!/usr/bin/env bash
# Checks that the environment dokku forwards across its privilege drop survives the
# local sudo implementation. sudo-rs, the default sudo on ubuntu 25.10 and later,
# ignores -E and bare --preserve-env, so dokku names every variable it needs.
set -eo pipefail

PLUGIN_CORE_AVAILABLE_PATH="${PLUGIN_CORE_AVAILABLE_PATH:-$(dirname "$0")/../../plugins}"
export PLUGIN_CORE_AVAILABLE_PATH
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"

FAILURES=0

fail() {
  declare desc="report a failed check"
  echo "not ok - $1" 1>&2
  FAILURES=$((FAILURES + 1))
}

pass() {
  declare desc="report a passing check"
  echo "ok - $1"
}

report-whole-environment-support() {
  declare desc="record whether this sudo honours -E, for the log"
  local warning

  warning="$(DOKKU_TEST_CROSSING=1 sudo -E -u nobody true 2>&1 >/dev/null || true)"
  if [[ -n "$warning" ]]; then
    echo "# sudo -E is not supported here: $warning"
  else
    echo "# sudo -E is supported here"
  fi
}

check-named-list-crosses() {
  declare desc="the named variables reach the target user"
  local flag output

  flag="$(DOKKU_QUIET_OUTPUT=1 DOKKU_TRACE=1 SSH_USER=ci fn-sudo-preserve-env-flag DOKKU_QUIET_OUTPUT DOKKU_TRACE SSH_USER)"
  output="$(DOKKU_QUIET_OUTPUT=1 DOKKU_TRACE=1 SSH_USER=ci sudo "$flag" -u nobody /usr/bin/env)"

  local name
  for name in DOKKU_QUIET_OUTPUT=1 DOKKU_TRACE=1 SSH_USER=ci; do
    if grep -qxF "$name" <<<"$output"; then
      pass "$name crossed the privilege drop"
    else
      fail "$name did not cross the privilege drop"
    fi
  done
}

check-named-list-is-silent() {
  declare desc="no warning is written to stderr"
  local flag stderr

  flag="$(SSH_USER=ci fn-sudo-preserve-env-flag SSH_USER)"
  stderr="$(SSH_USER=ci sudo "$flag" -u nobody true 2>&1 >/dev/null || true)"
  if [[ -z "$stderr" ]]; then
    pass "no warning on stderr"
  else
    fail "unexpected stderr: $stderr"
  fi
}

check-absent-names-are-tolerated() {
  declare desc="naming a variable that is not set is not an error"
  local flag

  flag="$(fn-sudo-preserve-env-flag DOKKU_DEFINITELY_NOT_SET_HERE SSH_USER)"
  if sudo "$flag" -u nobody true; then
    pass "unset names are ignored"
  else
    fail "sudo rejected a list naming an unset variable"
  fi
}

check-home-still-wins() {
  declare desc="-H is not defeated by the preserved list"
  local flag output
  local names=()

  read -ra names <<<"$(fn-cli-preserved-env-names)"
  flag="$(fn-sudo-preserve-env-flag "${names[@]}")"
  output="$(sudo "$flag" -H -u nobody /usr/bin/env)"
  if grep -qxF "HOME=$(getent passwd nobody | cut -d: -f6)" <<<"$output"; then
    pass "-H still sets HOME for the target user"
  else
    fail "HOME was not reset for the target user"
  fi
}

main() {
  declare desc="run every check"

  echo "# sudo version:"
  sudo --version | head -1 | sed -e 's/^/# /'

  report-whole-environment-support
  check-named-list-crosses
  check-named-list-is-silent
  check-absent-names-are-tolerated
  check-home-still-wins

  if [[ "$FAILURES" -gt 0 ]]; then
    echo "$FAILURES check(s) failed" 1>&2
    exit 1
  fi
  echo "all checks passed"
}

main "$@"
