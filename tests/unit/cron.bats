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

@test "(cron:report) --global --format json" {
  run /bin/bash -c "dokku cron:report --global --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"cron-maintenance\")'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"cron-global-maintenance\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"global-maintenance\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"cron-computed-maintenance\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"computed-maintenance\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"cron-global-mailfrom\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"global-mailfrom\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"cron-computed-mailfrom\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"computed-mailfrom\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"cron-global-mailto\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"global-mailto\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"cron-computed-mailto\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r 'has(\"computed-mailto\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report --global"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(cron:set) --global mailto" {
  run /bin/bash -c "dokku cron:set --global mailto admin@example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "unknown flag"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r '.\"cron-global-mailto\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "admin@example.com"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r '.\"cron-computed-mailto\"'"
  assert_success
  assert_output "admin@example.com"

  run /bin/bash -c "dokku cron:set --global mailto"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "unknown flag"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r '.\"cron-global-mailto\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku cron:report --global --format json | jq -r '.\"cron-computed-mailto\"'"
  assert_success
  assert_output ""
}

@test "(cron:set) --global mailfrom" {
  run /bin/bash -c "dokku cron:set --global mailfrom dokku@example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "unknown flag"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r '.\"cron-global-mailfrom\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku@example.com"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r '.\"cron-computed-mailfrom\"'"
  assert_success
  assert_output "dokku@example.com"

  run /bin/bash -c "dokku cron:set --global mailfrom"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "unknown flag"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r '.\"cron-global-mailfrom\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku cron:report --global --format json | jq -r '.\"cron-computed-mailfrom\"'"
  assert_success
  assert_output ""
}

@test "(cron:set) --global maintenance" {
  run /bin/bash -c "dokku cron:set --global maintenance true"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "unknown flag"

  run /bin/bash -c "dokku cron:report --global --format json | jq -r '.\"cron-global-maintenance\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:set --global maintenance"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "unknown flag"
}

@test "(cron:report) maintenance raw vs computed vs global" {
  run /bin/bash -c "dokku cron:set --global maintenance"
  assert_success

  run /bin/bash -c "dokku cron:report $TEST_APP --format json | jq -r '.\"cron-maintenance\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku cron:report $TEST_APP --format json | jq -r '.\"cron-global-maintenance\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku cron:report $TEST_APP --format json | jq -r '.\"cron-computed-maintenance\"'"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku cron:set --global maintenance true"
  assert_success

  run /bin/bash -c "dokku cron:report $TEST_APP --format json | jq -r '.\"cron-global-maintenance\"'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report $TEST_APP --format json | jq -r '.\"cron-computed-maintenance\"'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:set $TEST_APP maintenance true"
  assert_success

  run /bin/bash -c "dokku cron:report $TEST_APP --format json | jq -r '.\"cron-maintenance\"'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:report $TEST_APP --format json | jq -r '.\"cron-global-maintenance\"'"
  assert_success
  assert_output "true"

  # cron computed-maintenance returns "true" if either per-app or global is
  # "true" (cron's logic at plugins/cron/report.go::reportComputedMaintenance
  # is OR-not-override: any "true" wins). This is intentional for maintenance
  # mode - a per-app "false" does NOT override a global "true".
  run /bin/bash -c "dokku cron:report $TEST_APP --format json | jq -r '.\"cron-computed-maintenance\"'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku cron:set $TEST_APP maintenance"
  assert_success

  run /bin/bash -c "dokku cron:set --global maintenance"
  assert_success
}

@test "(cron:set) --global rejects task maintenance properties" {
  run /bin/bash -c "dokku cron:set --global maintenance.fakeid true"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Task maintenance properties cannot be set globally"
}

@test "(cron) create [multiple]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid_multiple
  echo "output: $output"
  echo "status: $status"
  assert_success

  first_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"
  second_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[1].id')"

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku cron:run $TEST_APP $first_id"
  assert_output_contains "dokku cron:run $TEST_APP $second_id"
  assert_output_contains "python3 task.py first" 0
  assert_output_contains "python3 task.py second" 0
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

  run /bin/bash -c "docker ps --filter "label=com.dokku.cron-id=$cron_id" -q | xargs docker inspect -f '{{ index .Config.Labels \"com.dokku.concurrency-policy\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "forbid"

  run /bin/bash -c "dokku cron:run $TEST_APP $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "currently has a cron lock in place for $cron_id"
  assert_failure
}

@test "(cron) cron:suspend cron:resume" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_valid_multiple
  echo "output: $output"
  echo "status: $status"
  assert_success

  first_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"
  second_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[1].id')"

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku cron:run $TEST_APP $first_id"
  assert_output_contains "dokku cron:run $TEST_APP $second_id"

  run /bin/bash -c "dokku cron:report $TEST_APP --cron-maintenance-$first_id"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku cron:suspend $TEST_APP $first_id"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku cron:report $TEST_APP --cron-maintenance-$first_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku cron:run $TEST_APP $first_id" 0
  assert_output_contains "dokku cron:run $TEST_APP $second_id"

  run /bin/bash -c "dokku cron:resume $TEST_APP $first_id"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku cron:report $TEST_APP --cron-maintenance-$first_id"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "cat /var/spool/cron/crontabs/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku cron:run $TEST_APP $first_id"
  assert_output_contains "dokku cron:run $TEST_APP $second_id"
}

@test "(cron) invalid [command]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_injection
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

@test "(cron) container labels regression" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_long_running
  echo "output: $output"
  echo "status: $status"
  assert_success

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"

  run /bin/bash -c "dokku cron:run $TEST_APP $cron_id --detach"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker ps --filter \"label=com.dokku.cron-id=$cron_id\" -q | xargs docker inspect -f '{{ index .Config.Labels \"com.dokku.cron-id\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$cron_id"

  run /bin/bash -c "docker ps --filter \"label=com.dokku.cron-id=$cron_id\" -q | xargs docker inspect -f '{{ index .Config.Labels \"com.dokku.concurrency-policy\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "allow"

  run /bin/bash -c "docker ps --filter \"label=com.dokku.cron-id=$cron_id\" -q | xargs docker inspect -f '{{ index .Config.Labels \"com.dokku.container-type\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "cron"

  run /bin/bash -c "docker ps --filter \"label=com.dokku.cron-id=$cron_id\" -q | xargs docker inspect -f '{{ index .Config.Labels \"com.dokku.app-name\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP"

  run /bin/bash -c "docker ps --filter \"label=com.dokku.cron-id=$cron_id\" -q | xargs docker inspect -f '{{ index .Config.Labels \"com.dokku.active-deadline-seconds\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "86400"
}

template_cron_file_long_running() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting long-running cron app.json -> $APP_REPO_DIR/app.json"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "command": "sleep 30",
      "schedule": "0 0 * * *"
    }
  ]
}
EOF
}

template_cron_file_injection() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting injection-attempt cron app.json -> $APP_REPO_DIR/app.json"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "command": "echo CRON_OK; echo hi > /tmp/appjson-injection-test.txt",
      "schedule": "* * * * *"
    }
  ]
}
EOF
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
