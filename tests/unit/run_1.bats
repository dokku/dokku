#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(run) run:help" {
  run /bin/bash -c "dokku run:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Run a one-off process inside a container"
}

@test "(run) run (with --options)" {
  deploy_app
  run /bin/bash -c "dokku --force --quiet run $TEST_APP python -V"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(run) run (with --env / -e)" {
  deploy_app
  run /bin/bash -c "dokku run --env TEST=testvalue -e TEST2=testvalue2 $TEST_APP env | grep -E '^TEST=testvalue'"
  echo "output: $output"
  echo "status: $status"

  run /bin/bash -c "dokku run --env TEST=testvalue -e TEST2=testvalue2 $TEST_APP env | grep -E '^TEST2=testvalue2'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
