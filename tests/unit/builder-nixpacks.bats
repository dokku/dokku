#!/usr/bin/env bats

load test_helper

setup_file() {
  install_nixpacks
}

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(builder-nixpacks:report) info-flag works before deploy" {
  run /bin/bash -c "dokku builder-nixpacks:report $TEST_APP --builder-nixpacks-nixpackstoml-path"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-nixpacks:set $TEST_APP nixpackstoml-path nixpacks.alt.toml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-nixpacks:report $TEST_APP --builder-nixpacks-nixpackstoml-path"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "nixpacks.alt.toml"

  run /bin/bash -c "dokku builder-nixpacks:report $TEST_APP --builder-nixpacks-invalid-flag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid flag passed"
}

@test "(builder-nixpacks:report) --global --builder-nixpacks-global-nixpackstoml-path" {
  run /bin/bash -c "dokku builder-nixpacks:set --global nixpackstoml-path nixpacks.alt.toml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-nixpacks:report --global --builder-nixpacks-global-nixpackstoml-path"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "nixpacks.alt.toml"

  run /bin/bash -c "dokku builder-nixpacks:set --global nixpackstoml-path"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(builder-nixpacks:report) nixpackstoml-path raw vs computed vs global" {
  run /bin/bash -c "dokku builder-nixpacks:set --global nixpackstoml-path"
  assert_success

  run /bin/bash -c "dokku builder-nixpacks:report $TEST_APP --builder-nixpacks-nixpackstoml-path"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku builder-nixpacks:report $TEST_APP --builder-nixpacks-global-nixpackstoml-path"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku builder-nixpacks:report $TEST_APP --builder-nixpacks-computed-nixpackstoml-path"
  assert_success
  assert_output "nixpacks.toml"

  run /bin/bash -c "dokku builder-nixpacks:set --global nixpackstoml-path nixpacks.global.toml"
  assert_success

  run /bin/bash -c "dokku builder-nixpacks:report $TEST_APP --builder-nixpacks-global-nixpackstoml-path"
  assert_success
  assert_output "nixpacks.global.toml"

  run /bin/bash -c "dokku builder-nixpacks:report $TEST_APP --builder-nixpacks-computed-nixpackstoml-path"
  assert_success
  assert_output "nixpacks.global.toml"

  run /bin/bash -c "dokku builder-nixpacks:set $TEST_APP nixpackstoml-path nixpacks.app.toml"
  assert_success

  run /bin/bash -c "dokku builder-nixpacks:report $TEST_APP --builder-nixpacks-nixpackstoml-path"
  assert_success
  assert_output "nixpacks.app.toml"

  run /bin/bash -c "dokku builder-nixpacks:report $TEST_APP --builder-nixpacks-global-nixpackstoml-path"
  assert_success
  assert_output "nixpacks.global.toml"

  run /bin/bash -c "dokku builder-nixpacks:report $TEST_APP --builder-nixpacks-computed-nixpackstoml-path"
  assert_success
  assert_output "nixpacks.app.toml"

  run /bin/bash -c "dokku builder-nixpacks:set $TEST_APP nixpackstoml-path"
  assert_success

  run /bin/bash -c "dokku builder-nixpacks:set --global nixpackstoml-path"
  assert_success
}

@test "(builder-nixpacks:set)" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected nixpacks"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_requirements_txt
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'load build definition from Dockerfile' -1
  assert_output_contains "SECRET_KEY: fjdkslafjdk"

  run /bin/bash -c "dokku builder-nixpacks:set $TEST_APP nixpackstoml-path nonexistent.toml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'load build definition from Dockerfile' -1
}

@test "(builder-nixpacks) run" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected nixpacks"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:add $TEST_APP build '--label=com.dokku.build-test=safe'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_requirements_txt
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'load build definition from Dockerfile' -1
  assert_output_contains "SECRET_KEY: fjdkslafjdk"

  run /bin/bash -c "dokku run $TEST_APP python3 task.py test"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "['task.py', 'test']"

  run /bin/bash -c "dokku --quiet run $TEST_APP task"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "['task.py', 'test']"

  run /bin/bash -c "dokku run $TEST_APP env"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "SECRET_KEY=fjdkslafjdk"
}

@test "(builder-nixpacks) cron:run" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected nixpacks"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP cron_run_wrapper
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'load build definition from Dockerfile' -1
  assert_output_contains "SECRET_KEY: fjdkslafjdk"

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"
  run /bin/bash -c "echo $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku cron:run $TEST_APP $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "['task.py', 'some', 'cron', 'task']"
}

@test "(builder-nixpacks) core-post-extract renames the configured nixpacks.toml" {
  echo "" >"$BATS_TEST_TMPDIR/nixpacks.alt.toml"

  run /bin/bash -c "dokku builder-nixpacks:set $TEST_APP nixpackstoml-path nixpacks.alt.toml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run_plugin_script builder-nixpacks core-post-extract "$TEST_APP" "$BATS_TEST_TMPDIR" HEAD
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $BATS_TEST_TMPDIR/nixpacks.toml"
  assert_success

  run /bin/bash -c "test ! -e $BATS_TEST_TMPDIR/nixpacks.alt.toml"
  assert_success
}

cron_run_wrapper() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  APP_REPO_DIR="$(realpath "$APP_REPO_DIR")"

  add_requirements_txt "$APP" "$APP_REPO_DIR"
  mv -f "$APP_REPO_DIR/app-cron.json" "$APP_REPO_DIR/app.json"
}
