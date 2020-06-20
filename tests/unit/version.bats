#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  global_teardown
}

@test "(version) version, -v, --version" {
  run /bin/bash -c "dokku version"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku -v"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --version"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
