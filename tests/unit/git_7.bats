#!/usr/bin/env bats

load test_helper

TEST_PLUGIN_APP=smoke-test-app
TEST_PLUGIN_GIT_REPO=https://github.com/dokku/${TEST_PLUGIN_APP}.git
TEST_PLUGIN_LOCAL_REPO="$(mktemp -d)/$TEST_PLUGIN_APP"

clone_test_app() {
  git clone "$TEST_PLUGIN_GIT_REPO" "$TEST_PLUGIN_LOCAL_REPO"
}

remove_test_app() {
  rm -rf $TEST_PLUGIN_LOCAL_REPO
}

setup() {
  global_setup
  create_app
  clone_test_app
}

teardown() {
  remove_test_app || true
  destroy_app
  global_teardown
}

@test "(git) push tags and branches" {
  # https://github.com/dokku/dokku/issues/5188
  local GIT_REMOTE=${GIT_REMOTE:="dokku@${DOKKU_DOMAIN}:$TEST_APP"}
  run git -C "$TEST_PLUGIN_LOCAL_REPO" remote add target "$GIT_REMOTE"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$TEST_PLUGIN_LOCAL_REPO" push target 1.0.0:master
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$TEST_PLUGIN_LOCAL_REPO" push target 2.0.0:master -f
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$TEST_PLUGIN_LOCAL_REPO" push target master -f
  echo "output: $output"
  echo "status: $status"
  assert_success
}
