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
