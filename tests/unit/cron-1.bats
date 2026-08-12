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
  # restore the default scheduler before destroy: a k3s-scheduled app cannot be
  # torn down cleanly without a cluster, and would otherwise leak into the next
  # test's cleanup_apps
  dokku scheduler:set "$TEST_APP" selected docker-local 2>/dev/null || true
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
  assert_failure
}

@test "(cron) scheduler-uses-host-cron" {
  run /bin/bash -c "dokku plugin:trigger scheduler-uses-host-cron docker-local"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku plugin:trigger scheduler-uses-host-cron k3s"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku plugin:trigger scheduler-uses-host-cron null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku plugin:trigger scheduler-uses-host-cron bogus-scheduler"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""
}

@test "(cron) invalid [missing-keys]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_invalid
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(cron) invalid [schedule]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_invalid_schedule
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(cron) invalid [seconds]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_invalid_schedule_seconds
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(cron) create [single-verbose]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid_single
  echo "output: $output"
  echo "status: $status"
  assert_success

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku cron:run $TEST_APP $cron_id"
  assert_output_contains "python3 task.py schedule" 0
}

@test "(cron) create [single-short]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid_short
  echo "output: $output"
  echo "status: $status"
  assert_success

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku cron:run $TEST_APP $cron_id"
  assert_output_contains "python3 task.py daily" 0
}

@test "(cron) k3s-scheduled app skips host crontab" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid_single
  echo "output: $output"
  echo "status: $status"
  assert_success

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku cron:run $TEST_APP $cron_id"

  # switch the app to a scheduler that manages its own cron backend
  run /bin/bash -c "dokku scheduler:set $TEST_APP selected k3s"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # regenerating the host crontab must now skip the k3s app's tasks, leaving
  # no host-cron tasks and therefore no crontab file
  run /bin/bash -c "dokku plugin:trigger scheduler-cron-write"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(cron) crontab format [no raw command leakage]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid_single
  echo "output: $output"
  echo "status: $status"
  assert_success

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku cron:run $TEST_APP $cron_id"
  # User command must never appear verbatim in the crontab - the crontab
  # only references the cron ID, and cron:run resolves the command at run
  # time and exec's it inside the container.
  assert_output_contains "python3 task.py schedule" 0
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
      "command": "python3 task.py",
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
      "command": "python3 task.py",
      "schedule": "0 5 * * * *"
    }
  ]
}
EOF
}

template_cron_file_valid_single() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting valid cron app.json -> $APP_REPO_DIR/app.json"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "command": "python3 task.py schedule",
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
      "command": "python3 task.py daily",
      "schedule": "@daily"
    }
  ]
}
EOF
}
