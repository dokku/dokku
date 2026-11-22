#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  rm -f /tmp/dokku-backup-*.tar.gz 2>/dev/null || true
  global_teardown
}

@test "(backup:import) prompts before overwriting and restores once confirmed" {
  dokku config:set --no-restart $TEST_APP PROMPTKEY=original

  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  dokku config:set --no-restart $TEST_APP PROMPTKEY=changed

  run /bin/bash -c "echo '$(basename "$backup_file")' | script -qec \"dokku backup:import '$backup_file'\" /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "This restore will overwrite the following"
  assert_output_contains "$TEST_APP" -1
  assert_output_contains "global configuration"

  run /bin/bash -c "dokku config:get $TEST_APP PROMPTKEY"
  echo "output: $output"
  assert_output "original"
}

@test "(backup:import) aborts without restoring when the confirmation does not match" {
  dokku config:set --no-restart $TEST_APP PROMPTKEY=original

  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  dokku config:set --no-restart $TEST_APP PROMPTKEY=changed

  run /bin/bash -c "echo 'not-the-name' | script -qec \"dokku backup:import '$backup_file'\" /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "confirmation did not match"

  run /bin/bash -c "dokku config:get $TEST_APP PROMPTKEY"
  echo "output: $output"
  assert_output "changed"
}

@test "(backup:import) --force restores without prompting" {
  dokku config:set --no-restart $TEST_APP PROMPTKEY=original

  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  dokku config:set --no-restart $TEST_APP PROMPTKEY=changed

  run /bin/bash -c "script -qec \"dokku backup:import --force '$backup_file'\" /dev/null </dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "This restore will overwrite the following" 0

  run /bin/bash -c "dokku config:get $TEST_APP PROMPTKEY"
  echo "output: $output"
  assert_output "original"
}
