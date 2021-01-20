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

@test "(cron) creation" {
  run deploy_app python dokku@dokku.me:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run deploy_app python dokku@dokku.me:$TEST_APP template_cron_file_valid
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python task.py schedule"

  run deploy_app python dokku@dokku.me:$TEST_APP template_cron_file_valid_short
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python task.py daily"

  run deploy_app python dokku@dokku.me:$TEST_APP template_cron_file_valid_multiple
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python task.py first"
  assert_output_contains "python task.py second"
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

template_cron_file_valid() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting valid cron app.json -> $APP_REPO_DIR/app.json"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "command": "python task.py schedule",
      "schedule": "5 5 5 5 5"
    }
  ]
}
EOF
}

template_cron_file_valid_short() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting valid cron app.json -> $APP_REPO_DIR/app.json"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "command": "python task.py daily",
      "schedule": "@daily"
    }
  ]
}
EOF
}

template_cron_file_valid_multiple() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting valid cron app.json -> $APP_REPO_DIR/app.json"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "command": "python task.py first",
      "schedule": "5 5 5 5 5"
    },
    {
      "command": "python task.py second",
      "schedule": "@daily"
    }
  ]
}
EOF
}
