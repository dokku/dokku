#!/usr/bin/env bats

load test_helper
TEST_PLUGIN_NAME=smoke-test-plugin
TEST_PLUGIN_GIT_REPO=https://github.com/dokku/${TEST_PLUGIN_NAME}.git

remove_test_plugin() {
  rm -rf $PLUGIN_ENABLED_PATH/$TEST_PLUGIN_NAME $PLUGIN_AVAILABLE_PATH/$TEST_PLUGIN_NAME
}

teardown() {
  remove_test_plugin || true
}

@test "(plugin) plugin:install, plugin:disable, plugin:uninstall" {
  run bash -c "dokku plugin:install $TEST_PLUGIN_GIT_REPO"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku plugin | grep enabled | grep $TEST_PLUGIN_NAME"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku plugin:disable $TEST_PLUGIN_NAME"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku plugin | grep disabled | grep $TEST_PLUGIN_NAME"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku plugin:uninstall $TEST_PLUGIN_NAME"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku plugin | grep $TEST_PLUGIN_NAME"
  echo "output: "$output
  echo "status: "$status
  assert_failure
}
