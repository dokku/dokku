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

@test "(cron) cron:help" {
  run /bin/bash -c "dokku cron"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage scheduled cron tasks"
  help_output="$output"

  run /bin/bash -c "dokku cron:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage scheduled cron tasks"
  assert_output "$help_output"
}

@test "(cron) invalid" {
  run deploy_app python dokku@dokku.me:$TEST_APP template_cron_file_invalid
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run deploy_app python dokku@dokku.me:$TEST_APP template_cron_file_invalid_schedule
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run deploy_app python dokku@dokku.me:$TEST_APP template_cron_file_invalid_schedule_seconds
  echo "output: $output"
  echo "status: $status"
  assert_failure
}


template_cron_file_invalid() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting invalid cron app.json -> $APP_REPO_DIR/app.json"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "key": "value"
    }
  ]
}
EOF
}

template_cron_file_invalid_schedule() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting invalid cron app.json -> $APP_REPO_DIR/app.json"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "command": "python task.py",
      "schedule": "@nonstandard"
    }
  ]
}
EOF
}

template_cron_file_invalid_schedule_seconds() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting invalid cron app.json -> $APP_REPO_DIR/app.json"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "command": "python task.py",
      "schedule": "0 5 * * * *"
    }
  ]
}
EOF
}
