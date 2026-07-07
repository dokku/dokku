#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
  dokku caddy:set --global log-level >/dev/null 2>&1 || true
  dokku caddy:set --global tls-internal >/dev/null 2>&1 || true
}

@test "(caddy:report) info-flag works before deploy" {
  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-computed-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-invalid-flag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid flag passed"
}

@test "(caddy:report) --format json" {
  run /bin/bash -c "dokku caddy:set --global log-level"
  assert_success

  run /bin/bash -c "dokku caddy:report --global --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:report --global --format json | jq -r '.\"computed-log-level\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ERROR"

  run /bin/bash -c "dokku caddy:report --global"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "global caddy information"
}

@test "(caddy:report) tls-internal raw global computed" {
  run /bin/bash -c "dokku caddy:set --global tls-internal"
  assert_success

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-tls-internal"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-computed-tls-internal"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku caddy:set --global tls-internal true"
  assert_success

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-global-tls-internal"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-computed-tls-internal"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku caddy:set $TEST_APP tls-internal false"
  assert_success

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-tls-internal"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-computed-tls-internal"
  assert_success
  assert_output "false"
}
