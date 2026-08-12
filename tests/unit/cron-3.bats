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

@test "(cron) invalid [command]" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_injection
  echo "output: $output"
  echo "status: $status"
  assert_failure
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

@test "(cron:run) retires containers past their active deadline" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_long_running
  echo "output: $output"
  echo "status: $status"
  assert_success

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"

  run /bin/bash -c "dokku cron:run $TEST_APP $cron_id --detach --ttl-seconds 1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  expiring_container="$output"

  run /bin/bash -c "docker container inspect $expiring_container --format '{{ index .Config.Labels \"com.dokku.active-deadline-seconds\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "docker container inspect $expiring_container --format '{{ index .Config.Labels \"com.dokku.container-type\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "cron"

  run /bin/bash -c "dokku cron:run $TEST_APP $cron_id --detach --ttl-seconds 300"
  echo "output: $output"
  echo "status: $status"
  assert_success
  surviving_container="$output"

  run /bin/bash -c "sleep 2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container ls --filter \"label=com.dokku.container-type=cron\" --format '{{ .Names }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "$expiring_container"
  assert_output_contains "$surviving_container"
}

@test "(cron:run) --ttl-seconds must be a positive integer" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_long_running
  echo "output: $output"
  echo "status: $status"
  assert_success

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"

  run /bin/bash -c "dokku cron:run $TEST_APP $cron_id --detach --ttl-seconds 0"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "--ttl-seconds must be a positive integer"

  run /bin/bash -c "dokku cron:run $TEST_APP $cron_id --detach --ttl-seconds -1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "--ttl-seconds must be a positive integer"
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
