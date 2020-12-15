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

@test "(git) deploy specific branch" {
  run /bin/bash -c "dokku git:set --global deploy-branch global-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success

  GIT_REMOTE_BRANCH=global-branch run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "deploy did not complete"

  run /bin/bash -c "dokku git:set $TEST_APP deploy-branch app-branch"
  GIT_REMOTE_BRANCH=app-branch run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:set --global deploy-branch"
}

@test "(git) git:initialize" {
  run /bin/bash -c "test -d $DOKKU_ROOT/$TEST_APP/refs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:initialize $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -d $DOKKU_ROOT/$TEST_APP/refs"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:initialize via deploy" {
  run /bin/bash -c "test -d $DOKKU_ROOT/$TEST_APP/refs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  deploy_app

  run /bin/bash -c "test -d $DOKKU_ROOT/$TEST_APP/refs"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
