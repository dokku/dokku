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

@test "(apps:lock) blocks ps:rebuild until apps:unlock" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP convert_to_dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builds:list $TEST_APP --format json | jq length"
  echo "output: $output"
  echo "status: $status"
  assert_success
  local BUILD_COUNT="$output"

  run /bin/bash -c "dokku apps:lock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "$TEST_APP has been locked via apps:lock" -1
  assert_output_not_contains "Building $TEST_APP"

  run /bin/bash -c "dokku builds:list $TEST_APP --format json | jq length"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$BUILD_COUNT"

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:unlock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps:lock) blocks git:sync before fetching" {
  run /bin/bash -c "dokku apps:lock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "$TEST_APP has been locked via apps:lock"

  run /bin/bash -c "sudo -u dokku git -C $DOKKU_ROOT/$TEST_APP rev-parse --verify -q master"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku builds:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "No builds recorded"
}
