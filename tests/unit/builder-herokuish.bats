#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(builder-herouish:build .env)" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'DOTENV_KEY=some_value'
}

@test "(builder-herokuish) builder-herokuish:set allowed" {
  if [[ "$(dpkg --print-architecture 2>/dev/null || true)" == "amd64" ]]; then
    skip "this test cannot be performed accurately on amd64 as it tests whether we can enable the plugin on arm64"
  fi

  run /bin/bash -c "dokku builder-herokuish:set --global allowed"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:set --global allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:set --global allowed"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(builder-herokuish) run" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected herokuish"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

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

@test "(builder-herokuish) cron:run" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected herokuish"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP cron_run_wrapper
  echo "output: $output"
  echo "status: $status"
  assert_success

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

cron_run_wrapper() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  APP_REPO_DIR="$(realpath "$APP_REPO_DIR")"

  mv -f "$APP_REPO_DIR/app-cron.json" "$APP_REPO_DIR/app.json"
}
