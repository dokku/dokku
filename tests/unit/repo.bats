#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  global_teardown
}

@test "(repo) repo:help" {
  run /bin/bash -c "dokku repo"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the app's repo"
  help_output="$output"

  run /bin/bash -c "dokku repo:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the app's repo"
  assert_output "$help_output"
}

@test "(repo) repo:gc, repo:purge-cache" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku repo:gc $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "touch $DOKKU_ROOT/$TEST_APP/cache/derp"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "find $DOKKU_ROOT/$TEST_APP/cache -type f | wc -l | grep 0"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  run /bin/bash -c "dokku repo:purge-cache $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "find $DOKKU_ROOT/$TEST_APP/cache -type f | wc -l | grep 0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
}
