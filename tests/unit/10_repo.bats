#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  deploy_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(repo) repo:gc, repo:purge-cache" {
  run /bin/bash -c "dokku repo:gc $TEST_APP"
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
}
