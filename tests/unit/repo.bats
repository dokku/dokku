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

@test "(repo) repo:gc" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku repo:gc $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(repo) repo:purge-cache" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker volume ls -q --filter label=com.dokku.app-name=$TEST_APP | wc -l"
  echo "count: '$output'"
  echo "status: $status"
  assert_output 1

  run /bin/bash -c "dokku repo:purge-cache $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker volume ls -q --filter label=com.dokku.app-name=$TEST_APP | wc -l"
  echo "count: '$output'"
  echo "status: $status"
  assert_output 0
}
