#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  install_fake_datastore
}

teardown() {
  remove_fake_datastore
  remove_user_auth_stub
  destroy_app
  rm -rf "/var/lib/dokku/data/storage/$TEST_APP-data" 2>/dev/null || true
  rm -f /tmp/dokku-backup-*.tar.gz 2>/dev/null || true
  global_teardown
}

@test "(backup:export) writes a tarball path to stdout and logs to stderr" {
  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku-backup-$TEST_APP-"
  assert_output_contains "Exporting backup" 0
  [[ -f "$output" ]]
}

@test "(backup) global config round-trip" {
  dokku config:set --global --no-restart BACKUP_GLOBAL_KEY=global_value

  run /bin/bash -c "dokku backup:export --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  dokku config:unset --global --no-restart BACKUP_GLOBAL_KEY

  run /bin/bash -c "dokku backup:import --force '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get --global BACKUP_GLOBAL_KEY"
  echo "output: $output"
  assert_output "global_value"

  dokku config:unset --global --no-restart BACKUP_GLOBAL_KEY 2>/dev/null || true
}

@test "(backup:import) refuses to overwrite an existing app without --force" {
  dokku config:set --no-restart $TEST_APP KEEP=me

  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp 2>/dev/null"
  assert_success
  local backup_file="$output"

  run /bin/bash -c "dokku backup:import '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "re-run with --force"
}

@test "(backup) service discovery and round-trip via datastore-list and service-list" {
  create_fake_service "mydb" "service-payload"

  run /bin/bash -c "dokku backup:export --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  run /bin/bash -c "tar tzf '$backup_file'"
  echo "output: $output"
  assert_output_contains "services/fake/mydb/data/data"

  rm -rf "/var/lib/dokku/services/fake/mydb"

  run /bin/bash -c "dokku backup:import --service fake:mydb '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  [[ "$(cat /var/lib/dokku/services/fake/mydb/data)" == "service-payload" ]]
}

@test "(backup) --include-storage round-trips managed volume data and entries" {
  dokku storage:ensure-directory $TEST_APP-data
  echo "vol-payload" >"/var/lib/dokku/data/storage/$TEST_APP-data/sentinel.txt"
  dokku storage:mount $TEST_APP "/var/lib/dokku/data/storage/$TEST_APP-data:/app/storage"

  run /bin/bash -c "dokku backup:export --app $TEST_APP --include-storage --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  run /bin/bash -c "tar tzf '$backup_file'"
  echo "output: $output"
  assert_output_contains "data/storage-data/" -1
  assert_output_contains "data/storage-entries/" -1

  rm -rf "/var/lib/dokku/data/storage/$TEST_APP-data"
  dokku --force apps:destroy $TEST_APP

  run /bin/bash -c "dokku backup:import '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  [[ "$(cat /var/lib/dokku/data/storage/$TEST_APP-data/sentinel.txt)" == "vol-payload" ]]

  run /bin/bash -c "dokku storage:list $TEST_APP"
  echo "output: $output"
  assert_output_contains "/app/storage"
}

@test "(backup:export) denies an app the caller cannot access" {
  install_user_auth_stub "$TEST_APP"

  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(backup:import) denies restoring an inaccessible existing app with a generic message" {
  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp 2>/dev/null"
  assert_success
  local backup_file="$output"

  install_user_auth_stub "$TEST_APP"

  run /bin/bash -c "dokku backup:import --force '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Unable to restore $TEST_APP"
}

install_fake_datastore() {
  cp -r "${BATS_TEST_DIRNAME}/../fake-datastore" /var/lib/dokku/plugins/available/fake-datastore
  mkdir -p /var/lib/dokku/plugins/enabled/fake-datastore
  cp -r /var/lib/dokku/plugins/available/fake-datastore/* /var/lib/dokku/plugins/enabled/fake-datastore/
  chmod +x /var/lib/dokku/plugins/available/fake-datastore/* /var/lib/dokku/plugins/enabled/fake-datastore/* 2>/dev/null || true
}

remove_fake_datastore() {
  rm -rf /var/lib/dokku/plugins/available/fake-datastore /var/lib/dokku/plugins/enabled/fake-datastore
  rm -rf /var/lib/dokku/services/fake
}

create_fake_service() {
  local name="$1" payload="$2"
  mkdir -p "/var/lib/dokku/services/fake/$name"
  echo "$payload" >"/var/lib/dokku/services/fake/$name/data"
}

install_user_auth_stub() {
  local forbidden="$1"
  mkdir -p /var/lib/dokku/plugins/available/user-auth-stub /var/lib/dokku/plugins/enabled/user-auth-stub
  cat >"/var/lib/dokku/plugins/available/user-auth-stub/plugin.toml" <<EOF
[plugin]
description = "user-auth stub for backup tests"
version = "0.1.0"
[plugin.config]
EOF
  cat >"/var/lib/dokku/plugins/available/user-auth-stub/user-auth-app" <<EOF
#!/usr/bin/env bash
set -eo pipefail
shift 2
for app in "\$@"; do
  [[ "\$app" == "$forbidden" ]] && continue
  echo "\$app"
done
EOF
  cat >"/var/lib/dokku/plugins/available/user-auth-stub/user-auth" <<EOF
#!/usr/bin/env bash
set -eo pipefail
exit 0
EOF
  chmod +x /var/lib/dokku/plugins/available/user-auth-stub/user-auth-app /var/lib/dokku/plugins/available/user-auth-stub/user-auth
  cp -r /var/lib/dokku/plugins/available/user-auth-stub/* /var/lib/dokku/plugins/enabled/user-auth-stub/
}

remove_user_auth_stub() {
  rm -rf /var/lib/dokku/plugins/available/user-auth-stub /var/lib/dokku/plugins/enabled/user-auth-stub
}
