#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  mkdir -p /var/lib/dokku/plugins/available/cron-entries
  mkdir -p /var/lib/dokku/plugins/enabled/cron-entries
  cat >"/var/lib/dokku/plugins/available/cron-entries/plugin.toml" <<EOF
[plugin]
description = "dokku test cron-entries plugin"
version = "0.0.1"
[plugin.config]
EOF
  cp /var/lib/dokku/plugins/available/cron-entries/plugin.toml /var/lib/dokku/plugins/enabled/cron-entries/plugin.toml

}

teardown() {
  rm -rf /var/lib/dokku/plugins/available/cron-entries /var/lib/dokku/plugins/enabled/cron-entries
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

@test "(cron) installed" {
  run /bin/bash -c "test -f /etc/sudoers.d/dokku-cron"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(cron) invalid [missing-keys]" {
  run deploy_app python dokku@dokku.me:$TEST_APP template_cron_file_invalid
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(cron) invalid [schedule]" {
  run deploy_app python dokku@dokku.me:$TEST_APP template_cron_file_invalid_schedule
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(cron) invalid [seconds]" {
  run deploy_app python dokku@dokku.me:$TEST_APP template_cron_file_invalid_schedule_seconds
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(cron) create [empty]" {
  run deploy_app python dokku@dokku.me:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(cron) create [single-verbose]" {
  run deploy_app python dokku@dokku.me:$TEST_APP template_cron_file_valid
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python task.py schedule"
}

@test "(cron) create [single-short]" {
  run deploy_app python dokku@dokku.me:$TEST_APP template_cron_file_valid_short
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python task.py daily"
}

@test "(cron) create [multiple]" {
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

@test "(cron) injected entries" {
  echo "echo '@daily;/bin/true'" > /var/lib/dokku/plugins/enabled/cron-entries/cron-entries
  chmod +x /var/lib/dokku/plugins/enabled/cron-entries/cron-entries

  run /bin/bash -c "dokku plugin:trigger cron-write"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "@daily /bin/true"

  # log file
  echo "echo '@daily;/bin/true;/var/log/dokku/log.log'" > /var/lib/dokku/plugins/enabled/cron-entries/cron-entries
  chmod +x /var/lib/dokku/plugins/enabled/cron-entries/cron-entries

  run /bin/bash -c "dokku plugin:trigger cron-write"
  echo "output: $output"
  echo "status: $status"
  assert_success


  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "@daily /bin/true &>> /var/log/dokku/log.log"

  # specify matching scheduler
  echo "[[ \$1 == 'docker-local' ]] && echo '@daily;/bin/true'" > /var/lib/dokku/plugins/enabled/cron-entries/cron-entries
  chmod +x /var/lib/dokku/plugins/enabled/cron-entries/cron-entries

  run /bin/bash -c "dokku plugin:trigger cron-write"
  echo "output: $output"
  echo "status: $status"
  assert_success


  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "@daily /bin/true"

  # specify non-matching scheduler
  echo "[[ \$1 == 'kubernetes' ]] && echo '@daily;/bin/true'" > /var/lib/dokku/plugins/enabled/cron-entries/cron-entries
  chmod +x /var/lib/dokku/plugins/enabled/cron-entries/cron-entries

  run /bin/bash -c "dokku plugin:trigger cron-write"
  echo "output: $output"
  echo "status: $status"
  assert_success


  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
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
