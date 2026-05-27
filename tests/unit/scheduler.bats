#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  dokku scheduler:set --global selected >/dev/null 2>&1 || true
  dokku scheduler:set --global shell >/dev/null 2>&1 || true
  destroy_app
  global_teardown
}

@test "(scheduler) scheduler:help" {
  run /bin/bash -c "dokku scheduler"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage scheduler settings"
  help_output="$output"

  run /bin/bash -c "dokku scheduler:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage scheduler settings"
  assert_output "$help_output"
}

@test "(scheduler:report) selected raw vs computed vs global" {
  run /bin/bash -c "dokku scheduler:set --global selected"
  assert_success

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-selected\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-global-selected\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-computed-selected\"'"
  assert_success
  assert_output "docker-local"

  run /bin/bash -c "dokku scheduler:set --global selected null"
  assert_success

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-global-selected\"'"
  assert_success
  assert_output "null"

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-computed-selected\"'"
  assert_success
  assert_output "null"

  run /bin/bash -c "dokku scheduler:set $TEST_APP selected docker-local"
  assert_success

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-selected\"'"
  assert_success
  assert_output "docker-local"

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-global-selected\"'"
  assert_success
  assert_output "null"

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-computed-selected\"'"
  assert_success
  assert_output "docker-local"

  run /bin/bash -c "dokku scheduler:set $TEST_APP selected"
  assert_success

  run /bin/bash -c "dokku scheduler:set --global selected"
  assert_success
}

@test "(scheduler:report) shell raw vs computed vs global" {
  run /bin/bash -c "dokku scheduler:set --global shell"
  assert_success

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-shell\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-global-shell\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-computed-shell\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku scheduler:set --global shell sh"
  assert_success

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-global-shell\"'"
  assert_success
  assert_output "sh"

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-computed-shell\"'"
  assert_success
  assert_output "sh"

  run /bin/bash -c "dokku scheduler:set $TEST_APP shell bash"
  assert_success

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-shell\"'"
  assert_success
  assert_output "bash"

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-global-shell\"'"
  assert_success
  assert_output "sh"

  run /bin/bash -c "dokku scheduler:report $TEST_APP --format json | jq -r '.\"scheduler-computed-shell\"'"
  assert_success
  assert_output "bash"

  run /bin/bash -c "dokku scheduler:set $TEST_APP shell"
  assert_success

  run /bin/bash -c "dokku scheduler:set --global shell"
  assert_success
}

@test "(scheduler:report) --global --format json" {
  run /bin/bash -c "dokku scheduler:set --global selected"
  assert_success

  run /bin/bash -c "dokku scheduler:report --global --format json | jq -e ."
  assert_success

  run /bin/bash -c "dokku scheduler:report --global --format json | jq -r '.\"scheduler-global-selected\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku scheduler:report --global --format json | jq -r '.\"scheduler-computed-selected\"'"
  assert_success
  assert_output "docker-local"

  run /bin/bash -c "dokku scheduler:report --global --format json | jq -r 'has(\"scheduler-selected\")'"
  assert_success
  assert_output "false"
}
