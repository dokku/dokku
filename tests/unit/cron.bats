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
  assert_failure
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

@test "(cron) create [empty]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(cron) create [single-verbose]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid_single
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python3 task.py schedule"
}

@test "(cron) create [single-short]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid_short
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python3 task.py daily"
}

@test "(cron) create [multiple]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid_multiple
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python3 task.py first"
  assert_output_contains "python3 task.py second"
}

@test "(cron) injected entries" {
  echo "echo '@daily;/bin/true'" >/var/lib/dokku/plugins/enabled/cron-entries/cron-entries
  chmod +x /var/lib/dokku/plugins/enabled/cron-entries/cron-entries

  run /bin/bash -c "dokku plugin:trigger scheduler-cron-write"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku cron:list --global --format json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output '[{"id":"5cruaotm4yzzpnjlsdunblj8qyjp","command":"/bin/true","global":true,"schedule":"@daily","concurrency_policy":"","app-in-maintenance":false,"task-in-maintenance":false,"maintenance":false}]'

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "@daily /bin/true"

  # log file
  echo "echo '@daily;/bin/true;/var/log/dokku/log.log'" >/var/lib/dokku/plugins/enabled/cron-entries/cron-entries
  chmod +x /var/lib/dokku/plugins/enabled/cron-entries/cron-entries

  run /bin/bash -c "dokku plugin:trigger scheduler-cron-write"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "@daily /bin/true &>> /var/log/dokku/log.log"

  # specify matching scheduler
  echo "[[ \$1 == 'docker-local' ]] && echo '@daily;/bin/true'" >/var/lib/dokku/plugins/enabled/cron-entries/cron-entries
  chmod +x /var/lib/dokku/plugins/enabled/cron-entries/cron-entries

  run /bin/bash -c "dokku plugin:trigger scheduler-cron-write"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "@daily /bin/true"

  # specify non-matching scheduler
  echo "[[ \$1 == 'kubernetes' ]] && echo '@daily;/bin/true'" >/var/lib/dokku/plugins/enabled/cron-entries/cron-entries
  chmod +x /var/lib/dokku/plugins/enabled/cron-entries/cron-entries

  run /bin/bash -c "dokku plugin:trigger scheduler-cron-write"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(cron) cron:list --format json" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku cron:list $TEST_APP --format json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"
  run /bin/bash -c "echo $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists
}

@test "(cron) cron:run" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid
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
  assert_output "['task.py', 'schedule']"

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[1].id')"
  run /bin/bash -c "echo $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku cron:run $TEST_APP $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "['task.py', 'schedule', 'now']"
}

@test "(cron) cron:run concurrency_policy forbid" {
  run deploy_app dockerfile dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_concurrency_forbid
  echo "output: $output"
  echo "status: $status"
  assert_success

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"
  run /bin/bash -c "echo cron $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku cron:run $TEST_APP $cron_id --detach"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker ps --filter "label=com.dokku.cron-id=$cron_id" -q | xargs docker inspect"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku cron:run $TEST_APP $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(cron) cron:suspend cron:resume" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid_multiple
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python3 task.py first"
  assert_output_contains "python3 task.py second"

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"
  run /bin/bash -c "echo $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  first_command="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].command')"
  run /bin/bash -c "echo $first_command"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku cron:report $TEST_APP --cron-maintenance-$cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku cron:suspend $TEST_APP $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku cron:report $TEST_APP --cron-maintenance-$cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python3 task.py first" 0
  assert_output_contains "python3 task.py second"

  run /bin/bash -c "dokku cron:resume $TEST_APP $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku cron:report $TEST_APP --cron-maintenance-$cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python3 task.py first"
  assert_output_contains "python3 task.py second"
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

template_cron_file_valid() {
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
    },
    {
      "command": "python3 task.py schedule now",
      "schedule": "6 5 5 5 5"
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

template_cron_file_valid_multiple() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting valid cron app.json -> $APP_REPO_DIR/app.json"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "command": "python3 task.py first",
      "schedule": "5 5 5 5 5"
    },
    {
      "command": "python3 task.py second",
      "schedule": "@daily"
    }
  ]
}
EOF
}

template_cron_file_concurrency_forbid() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting valid cron app.json -> $APP_REPO_DIR/app.json"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "command": "sleep 30",
      "schedule": "0 0 * * *",
      "concurrency_policy": "forbid"
    }
  ]
}
EOF
}
