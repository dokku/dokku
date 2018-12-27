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
  echo "output: "$output
  echo "status: "$status
  assert_success

  GIT_REMOTE_BRANCH=global-branch deploy_app
  echo "output: "$output
  echo "status: "$status
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run /bin/bash -c "dokku git:set $TEST_APP deploy-branch app-branch"
  GIT_REMOTE_BRANCH=app-branch deploy_app
  echo "output: "$output
  echo "status: "$status
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run /bin/bash -c "dokku git:set --global deploy-branch"
}

@test "(git) ensure GIT_REV env var is set" {
  deploy_app
  echo "output: "$output
  echo "status: "$status
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP GIT_REV"
  echo "output: "$output
  echo "status: "$status
  assert_output_exists
}

@test "(git) disable GIT_REV" {
  run /bin/bash -c "dokku git:set $TEST_APP rev-env-var"
  echo "output: "$output
  echo "status: "$status
  assert_success

  deploy_app
  echo "output: "$output
  echo "status: "$status
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP GIT_REV"
  echo "output: "$output
  echo "status: "$status
  assert_output ""
}

@test "(git) customize the GIT_REV environment variable" {
  run /bin/bash -c "dokku git:set $TEST_APP rev-env-var GIT_REV_ALT"
  echo "output: "$output
  echo "status: "$status
  assert_success

  deploy_app
  echo "output: "$output
  echo "status: "$status
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP GIT_REV_ALT"
  echo "output: "$output
  echo "status: "$status
  assert_output_exists
}

@test "(git) git:initialize" {
  run /bin/bash -c "test -d $DOKKU_ROOT/$TEST_APP/refs"
  echo "output: "$output
  echo "status: "$status
  assert_failure

  run /bin/bash -c "dokku git:initialize $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run /bin/bash -c "test -d $DOKKU_ROOT/$TEST_APP/refs"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(git) git:initialize via deploy" {
  run /bin/bash -c "test -d $DOKKU_ROOT/$TEST_APP/refs"
  echo "output: "$output
  echo "status: "$status
  assert_failure

  deploy_app

  run /bin/bash -c "test -d $DOKKU_ROOT/$TEST_APP/refs"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
