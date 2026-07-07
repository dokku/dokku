#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  clone_test_plugin
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

@test "(plugin) plugin:list --format json" {
  run /bin/bash -c "dokku plugin:list --format nonsense"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid output format"

  run /bin/bash -c "dokku plugin:list --format json | jq -r '.[] | select(.name == \"apps\") | .core'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku plugin:list --format json | jq -r '.[] | select(.name == \"apps\") | .enabled'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku plugin:list --format json | jq -r '.[] | select(.name == \"apps\") | .source_url'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  # Clone a git-based third-party plugin directly into the available path so the
  # git install source is present without running the (build-dependent) install
  # trigger. plugn lists it as a disabled plugin.
  run /bin/bash -c "git clone $TEST_PLUGIN_GIT_REPO $PLUGIN_AVAILABLE_PATH/$TEST_PLUGIN_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:list --format json | jq -r '.[] | select(.name == \"$TEST_PLUGIN_NAME\") | .source_url'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_PLUGIN_GIT_REPO"

  run /bin/bash -c "dokku plugin:list --format json | jq -r '.[] | select(.name == \"$TEST_PLUGIN_NAME\") | .core'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku plugin:list --format json | jq -r '.[] | select(.name == \"$TEST_PLUGIN_NAME\") | .enabled'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku plugin:list --format json | jq -r '.[] | select(.name == \"$TEST_PLUGIN_NAME\") | .committish | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "40"

  run /bin/bash -c "dokku plugin:list --format json | jq -r '.[] | select(.name == \"$TEST_PLUGIN_NAME\") | .branch'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists
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

@test "(plugin) plugin:install [errors]" {
  run /bin/bash -c "dokku plugin:install YABBA_DABBA_DOO XXXX YYYY"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Please retry with valid arguments"

  run /bin/bash -c "dokku plugin:install ZXZX --random-flag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Please retry with valid arguments"

  run /bin/bash -c "dokku plugin:install http://www.example.com/ --random-flag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Please retry with valid arguments"

  run /bin/bash -c "dokku plugin:install http://www.example.com/gives-a-404"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Please retry with valid arguments"

  run /bin/bash -c "dokku plugin:install http://xxxx/ --random-flag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Please retry with valid arguments"

  run /bin/bash -c "dokku plugin:install /path/to/nonexistent/dir"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Please retry with valid arguments"
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
