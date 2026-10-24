#!/usr/bin/env bats

load test_helper

APP_FILTER_PLUGIN_NAME="app-filter-test"

setup() {
  global_setup
}

teardown() {
  release_deploy_lock
  rm -rf "${PLUGIN_ENABLED_PATH:?}/$APP_FILTER_PLUGIN_NAME" "${PLUGIN_AVAILABLE_PATH:?}/$APP_FILTER_PLUGIN_NAME"
  dokku --force apps:destroy "$TEST_APP-source" &>/dev/null || true
  global_teardown
}

@test "(apps) apps:help" {
  run /bin/bash -c "dokku apps --help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage apps"
  help_output="$output"

  run /bin/bash -c "dokku apps:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage apps"
  assert_output "$help_output"
}

@test "(apps) apps:list" {
  run /bin/bash -c "dokku apps:list 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "You haven't deployed any applications yet"
  assert_output_contains "$TEST_APP" 0

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:list 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "You haven't deployed any applications yet" 0
  assert_output_contains "$TEST_APP"

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps) apps:list --format json" {
  run /bin/bash -c "dokku apps:list --format json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "[]"

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:list --format json | jq '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "dokku apps:list --format json | jq -r '.[0]'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP"

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps) apps:create" {
  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:list | grep $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output $TEST_APP

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:create 1994testapp"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:create testapp:latest"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create testApp:latest"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create TestApp"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create test/app"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create test_app"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create 'test;id'"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create 'test\$cmd'"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create 'test|cmd'"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create 'test&cmd'"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku --app $TEST_APP apps:create"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:list | grep $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output $TEST_APP

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps) app autocreate disabled" {
  run /bin/bash -c "dokku apps:set --global disable-autocreation true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_failure
  run /bin/bash -c "dokku apps:set --global disable-autocreation"
}

@test "(apps) apps:destroy" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force --app $TEST_APP apps:destroy"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps) apps:rename" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "cat $DOKKU_ROOT/$TEST_APP/hooks/pre-receive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "git-hook \"$TEST_APP\""
  run /bin/bash -c "dokku apps:rename $TEST_APP great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "cat $DOKKU_ROOT/great-test-name/hooks/pre-receive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "git-hook \"great-test-name\""
  run /bin/bash -c "dokku apps:list | grep $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists
  run /bin/bash -c "curl --silent --write-out '%{http_code}\n' $(dokku url great-test-name) | grep 404"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists
  run /bin/bash -c "dokku --force apps:destroy great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku buildpacks:add $TEST_APP https://github.com/heroku/heroku-buildpack-ruby.git"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP KEY=VALUE-$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy --no-healthcheck"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku git:set $TEST_APP deploy-branch SOME_BRANCH_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku network:set $TEST_APP attach-post-create test-network"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku nginx:set $TEST_APP  bind-address-ipv4 127.0.0.1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku resource:limit --memory 100 $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:rename $TEST_APP great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku buildpacks:report great-test-name --buildpacks-list"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "https://github.com/heroku/heroku-buildpack-ruby.git"
  run /bin/bash -c "dokku config:get great-test-name KEY"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "VALUE-$TEST_APP"
  run /bin/bash -c "dokku docker-options:report great-test-name --docker-options-deploy"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--no-healthcheck"
  run /bin/bash -c "dokku git:report great-test-name --git-deploy-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "SOME_BRANCH_NAME"
  run /bin/bash -c "dokku network:report great-test-name --network-attach-post-create"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test-network"
  run /bin/bash -c "dokku nginx:report great-test-name --nginx-bind-address-ipv4"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "127.0.0.1"
  run /bin/bash -c "dokku resource:report great-test-name --resource-_default_.limit.memory"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "100"

  run /bin/bash -c "dokku --force apps:destroy great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps) apps:rename with tls" {
  run setup_test_tls
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:rename $TEST_APP great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps) apps:exists" {
  run /bin/bash -c "dokku apps:exists $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:exists $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

}

@test "(apps) apps:lock/locked/unlock" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:report $TEST_APP --app-locked"
  echo "output: $output"
  echo "status: $status"
  assert_output "false"

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:lock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:report $TEST_APP --app-locked"
  echo "output: $output"
  echo "status: $status"
  assert_output "true"

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:unlock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:report $TEST_APP --app-locked"
  echo "output: $output"
  echo "status: $status"
  assert_output "false"

  destroy_app
}

@test "(apps) apps:lock does not modify the deploy lock file" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo holder-id >$DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock && chown dokku:dokku $DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:lock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat $DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "holder-id"

  run /bin/bash -c "dokku apps:unlock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  destroy_app
}

@test "(apps:locked) ignores a deploy lock file that no deploy holds" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo stale-id >$DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock && chown dokku:dokku $DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:report $TEST_APP --app-locked"
  echo "output: $output"
  echo "status: $status"
  assert_output "false"

  destroy_app
}

@test "(apps:locked) reports a deploy lock held by a running deploy" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  hold_deploy_lock

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:unlock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "A deploy may be in progress."

  release_deploy_lock
  destroy_app
}

@test "(apps:unlock) succeeds for an app that is not locked" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:unlock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "A deploy may be in progress."

  destroy_app
}

@test "(apps:rename) fails for a locked app" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:lock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:rename $TEST_APP great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Unable to rename a locked app"

  run /bin/bash -c "dokku apps:exists great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  destroy_app
}

@test "(apps) [install] converts a lock created by apps:lock in older versions" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sudo -u dokku touch $DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "DOKKU_ROOT=$DOKKU_ROOT PLUGIN_PATH=$PLUGIN_PATH PLUGIN_AVAILABLE_PATH=$PLUGIN_AVAILABLE_PATH PLUGIN_ENABLED_PATH=$PLUGIN_ENABLED_PATH PLUGIN_CORE_AVAILABLE_PATH=$PLUGIN_CORE_AVAILABLE_PATH DOKKU_LIB_ROOT=$DOKKU_LIB_ROOT $PLUGIN_ENABLED_PATH/apps/install"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "[[ ! -e $DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock ]]"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "[[ -f $DOKKU_LIB_ROOT/data/apps/$TEST_APP/.manual-deploy.lock ]]"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  destroy_app
}

@test "(apps:report) emits new stripped JSON keys alongside legacy" {
  create_app

  run /bin/bash -c "dokku apps:report $TEST_APP --format json | jq -r 'has(\"dir\") and has(\"app-dir\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku apps:report $TEST_APP --format json | jq -r 'has(\"locked\") and has(\"app-locked\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku apps:report $TEST_APP --format json | jq -r 'has(\"deploy-source\") and has(\"app-deploy-source\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku apps:report $TEST_APP --format json | jq -r 'has(\"global-disable-autocreation\") and has(\"app-global-disable-autocreation\")'"
  assert_success
  assert_output "true"

  destroy_app
}

@test "(apps:report --global) emits global-disable-autocreation" {
  run /bin/bash -c "dokku apps:report --global --app-global-disable-autocreation"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku apps:set --global disable-autocreation true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:report --global --app-global-disable-autocreation"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku apps:report --global --format json | jq -r '.\"global-disable-autocreation\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku apps:report --global --format json | jq -r '.\"app-global-disable-autocreation\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  create_app

  run /bin/bash -c "dokku apps:report $TEST_APP --format json | jq -r '.\"global-disable-autocreation\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  destroy_app

  run /bin/bash -c "dokku apps:set --global disable-autocreation"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps:set) rejects dead writes" {
  run /bin/bash -c "dokku apps:set --global deploy-source git"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Property cannot be specified globally"

  run /bin/bash -c "dokku apps:set --global deploy-source-metadata foo"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Property cannot be specified globally"

  create_app

  run /bin/bash -c "dokku apps:set $TEST_APP disable-autocreation true"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Property can only be specified globally"

  destroy_app
}

@test "(apps) apps:create, apps:clone, and apps:rename with an existing filtered app" {
  create_app
  create_app "$TEST_APP-source"

  setup_app_filter_plugin "$TEST_APP"

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Name is already taken"
  assert_output_not_contains "Creating $TEST_APP"

  run /bin/bash -c "dokku apps:clone $TEST_APP-source $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Name is already taken"

  run /bin/bash -c "dokku apps:rename $TEST_APP-source $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Name is already taken"

  rm -rf "${PLUGIN_ENABLED_PATH:?}/$APP_FILTER_PLUGIN_NAME" "${PLUGIN_AVAILABLE_PATH:?}/$APP_FILTER_PLUGIN_NAME"

  destroy_app 0 "$TEST_APP-source"
  destroy_app
}

setup_app_filter_plugin() {
  declare desc="installs a plugin that filters out a single app"
  declare FILTERED_APP="$1"
  local PLUGIN_DIR="$PLUGIN_AVAILABLE_PATH/$APP_FILTER_PLUGIN_NAME"
  mkdir -p "$PLUGIN_DIR"

  cat <<EOF >"$PLUGIN_DIR/plugin.toml"
[plugin]
description = "test plugin filtering out a single app"
version = "0.1.0"
[plugin.config]
EOF

  cat <<EOF >"$PLUGIN_DIR/user-auth-app"
#!/usr/bin/env bash
shift 2
for app in "\$@"; do
  [[ "\$app" == "$FILTERED_APP" ]] || echo "\$app"
done
EOF
  chmod +x "$PLUGIN_DIR/user-auth-app"

  dokku plugin:enable "$APP_FILTER_PLUGIN_NAME"
}
