#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  destroy_app
  global_teardown
}

@test "(network) logs:help" {
  run /bin/bash -c "dokku logs:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage log integration for an app"
}
