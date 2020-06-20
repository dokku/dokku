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

@test "(git) git:help" {
  run /bin/bash -c "dokku git"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage app deploys via git"
  help_output="$output"

  run /bin/bash -c "dokku git:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage app deploys via git"
  assert_output "$help_output"
}

@test "(git) ensure GIT_REV env var is set" {
  deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP GIT_REV"
  echo "output: $output"
  echo "status: $status"
  assert_output_exists
}

@test "(git) disable GIT_REV" {
  run /bin/bash -c "dokku git:set $TEST_APP rev-env-var"
  echo "output: $output"
  echo "status: $status"
  assert_success

  deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP GIT_REV"
  echo "output: $output"
  echo "status: $status"
  assert_output ""
}

@test "(git) customize the GIT_REV environment variable" {
  run /bin/bash -c "dokku git:set $TEST_APP rev-env-var GIT_REV_ALT"
  echo "output: $output"
  echo "status: $status"
  assert_success

  deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP GIT_REV_ALT"
  echo "output: $output"
  echo "status: $status"
  assert_output_exists
}
