#!/usr/bin/env bats

load test_helper
TEST_PLUGIN_NAME=smoke-test-plugin
TEST_PLUGIN_GIT_REPO=https://github.com/dokku/${TEST_PLUGIN_NAME}.git
TEST_PLUGIN_LOCAL_REPO="$(mktemp -d)/$TEST_PLUGIN_NAME"

clone_test_plugin() {
  git clone "$TEST_PLUGIN_GIT_REPO" "$TEST_PLUGIN_LOCAL_REPO"
}

setup() {
  global_setup
  clone_test_plugin
}

remove_test_plugin() {
  rm -rf $PLUGIN_ENABLED_PATH/$TEST_PLUGIN_NAME $PLUGIN_AVAILABLE_PATH/$TEST_PLUGIN_NAME
  rm -rf $TEST_PLUGIN_LOCAL_REPO
}

teardown() {
  remove_test_plugin || true
  global_teardown
}

@test "(plugin) plugin:help" {
  run /bin/bash -c "dokku plugin"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage installed plugins"
  help_output="$output"

  run /bin/bash -c "dokku plugin:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage installed plugins"
  assert_output "$help_output"
}

@test "(plugin) plugin:install, plugin:disable, plugin:update plugin:uninstall" {
  run /bin/bash -c "dokku plugin:installed $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku plugin:install $TEST_PLUGIN_GIT_REPO --name $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:installed $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:update $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:list | grep enabled | grep $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sudo -E -u nobody dokku plugin:uninstall $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku plugin:disable $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:list | grep disabled | grep $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:uninstall $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:list | grep $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(plugin) plugin:install, plugin:disable, plugin:update plugin:uninstall (with file://)" {
  run /bin/bash -c "dokku plugin:install file://$TEST_PLUGIN_LOCAL_REPO --name $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:update $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:list | grep enabled | grep $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sudo -E -u nobody dokku plugin:uninstall $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku plugin:disable $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:list | grep disabled | grep $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:uninstall $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:list | grep $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(plugin) plugin:install plugin:update (with tag)" {
  run /bin/bash -c "dokku plugin:install $TEST_PLUGIN_GIT_REPO --committish v0.2.0 --name $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:list | grep enabled | grep $TEST_PLUGIN_NAME | grep 0.2.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:update $TEST_PLUGIN_NAME v0.3.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:list | grep enabled | grep $TEST_PLUGIN_NAME | grep 0.2.0"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku plugin:list | grep enabled | grep $TEST_PLUGIN_NAME | grep 0.3.0"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(plugin) plugin:install, plugin:disable, plugin:uninstall as non-root user failure" {
  run /bin/bash -c "sudo -E -u nobody dokku plugin:install $TEST_PLUGIN_GIT_REPO"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku plugin:install $TEST_PLUGIN_GIT_REPO"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:list | grep enabled | grep $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sudo -E -u nobody dokku plugin:disable $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(plugin) plugin:install permissions set properly" {
  run /bin/bash -c "ls -lah /var/lib/dokku/core-plugins/available/git/commands"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "root" "0"
  assert_output_contains "dokku" "3"

  run /bin/bash -c "chown -R root:root /var/lib/dokku/core-plugins/available/git"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "ls -lah /var/lib/dokku/core-plugins/available/git/commands"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "root" "2"
  assert_output_contains "dokku" "1"

  run /bin/bash -c "dokku plugin:install"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "ls -lah /var/lib/dokku/core-plugins/available/git/commands"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "root" "0"
  assert_output_contains "dokku" "3"
}

@test "(plugin) plugin:update [errors]" {
  run /bin/bash -c "dokku plugin:install $TEST_PLUGIN_GIT_REPO --name $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:disable $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:update $TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Specified plugin not enabled or installed"

  run /bin/bash -c "dokku plugin:update invalid"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Specified plugin not enabled or installed"

  run /bin/bash -c "dokku plugin:update $TEST_PLUGIN_GIT_REPO"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid plugin name specified"

  run /bin/bash -c "dokku plugin:update app-json"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "please update Dokku instead"
}

@test "(plugin) plugin:update permissions set properly" {
  run /bin/bash -c "ls -lah /var/lib/dokku/core-plugins/available/git/commands"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "root" "0"
  assert_output_contains "dokku" "3"

  run /bin/bash -c "chown -R root:root /var/lib/dokku/core-plugins/available/git"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "ls -lah /var/lib/dokku/core-plugins/available/git/commands"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "root" "2"
  assert_output_contains "dokku" "1"

  run /bin/bash -c "dokku plugin:update"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "ls -lah /var/lib/dokku/core-plugins/available/git/commands"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "root" "0"
  assert_output_contains "dokku" "3"
}
