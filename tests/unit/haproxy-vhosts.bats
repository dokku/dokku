#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
  dokku haproxy:set --global log-level >/dev/null 2>&1 || true
  dokku haproxy:set --global refresh-conf >/dev/null 2>&1 || true
}

@test "(haproxy:report) info-flag works before deploy" {
  run /bin/bash -c "dokku haproxy:set --global refresh-conf"
  assert_success

  run /bin/bash -c "dokku haproxy:report $TEST_APP --haproxy-computed-refresh-conf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "10"

  run /bin/bash -c "dokku haproxy:report $TEST_APP --haproxy-invalid-flag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid flag passed"
}

@test "(haproxy:report) --format json" {
  run /bin/bash -c "dokku haproxy:set --global log-level"
  assert_success

  run /bin/bash -c "dokku haproxy:report --global --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:report --global --format json | jq -r '.\"computed-log-level\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ERROR"

  run /bin/bash -c "dokku haproxy:report --global"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "global haproxy information"
}

@test "(haproxy:report) refresh-conf global computed" {
  run /bin/bash -c "dokku haproxy:set --global refresh-conf"
  assert_success

  run /bin/bash -c "dokku haproxy:report --global --haproxy-global-refresh-conf"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku haproxy:report --global --haproxy-computed-refresh-conf"
  assert_success
  assert_output "10"

  run /bin/bash -c "dokku haproxy:set --global refresh-conf 30"
  assert_success

  run /bin/bash -c "dokku haproxy:report --global --haproxy-global-refresh-conf"
  assert_success
  assert_output "30"

  run /bin/bash -c "dokku haproxy:report --global --haproxy-computed-refresh-conf"
  assert_success
  assert_output "30"
}
