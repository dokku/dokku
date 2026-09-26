#!/usr/bin/env bats

load test_helper

APP_FILTER_PLUGIN_NAME="app-filter-test"

setup() {
  global_setup
  create_app
}

teardown() {
  rm -rf "${PLUGIN_ENABLED_PATH:?}/$APP_FILTER_PLUGIN_NAME" "${PLUGIN_AVAILABLE_PATH:?}/$APP_FILTER_PLUGIN_NAME"
  rm -rf /tmp/hidden-app-clone
  destroy_app
  global_teardown
}

@test "(git) push to filtered app" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  setup_app_filter_plugin "$TEST_APP"

  run /bin/bash -c "dokku apps:exists $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "cd /tmp && rm -rf hidden-app-clone && cp -r $BATS_TEST_DIRNAME/../../tests/apps/python hidden-app-clone && cd hidden-app-clone && git init -q && git add . && git -c user.email=robot@example.com -c user.name=Robot commit -qm second && git push -f dokku@$DOKKU_DOMAIN:$TEST_APP HEAD:master"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "App $TEST_APP does not exist"
  assert_output_not_contains "pre-receive hook declined"
}

@test "(git) clone of filtered app" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  setup_app_filter_plugin "$TEST_APP"

  run /bin/bash -c "git clone dokku@$DOKKU_DOMAIN:$TEST_APP /tmp/hidden-app-clone"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "App $TEST_APP does not exist"
}

@test "(git) archive of filtered app" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  setup_app_filter_plugin "$TEST_APP"

  run /bin/bash -c "set -o pipefail; git archive --remote=dokku@$DOKKU_DOMAIN:$TEST_APP master | tar -t"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "App $TEST_APP does not exist"
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
