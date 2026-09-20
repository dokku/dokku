#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  rm -rf "$DOKKU_LIB_ROOT/data/storage/rdmtestapp*"
}

teardown() {
  destroy_app
  global_teardown
}

@test "(storage:mount) --replace replaces the whole mount set" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-logs"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-data --container-dir /app/storage"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-cache:/cache rdmtest-logs:/logs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].entry_name' | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-cache rdmtest-logs"

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-cache:/cache rdmtest-logs:/logs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-data --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-cache --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-logs --force"
  assert_success
}

@test "(storage:mount) --replace leaves other process types untouched" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-cache --container-dir /cache --process-type web"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].entry_name' | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-cache rdmtest-data"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.process-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "web"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.2.process-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "_default_"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-data --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-cache --force"
  assert_success
}

@test "(storage:mount) --replace applies scope flags to every spec" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage:Z rdmtest-cache:/cache:ro,noexec --phase deploy --process-type web --volume-subpath uploads --volume-chown herokuish"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.phases\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "deploy"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.subpath\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "uploads"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.volume-chown\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "herokuish"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.volume-options\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "Z"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.2.readonly\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.2.volume-options\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "noexec"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-data --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-cache --force"
  assert_success
}

@test "(storage:mount) --replace leaves the existing set untouched when a spec is invalid" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-data --container-dir /app/storage"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-cache:/cache rdmtest-missing:/missing"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "does not exist" -1

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].entry_name' | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-data"

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage rdmtest-cache:/app/storage"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "specified more than once" -1

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].entry_name' | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-data"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-data --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-cache --force"
  assert_success
}

@test "(storage:mount) --replace rejects an empty spec list" {
  run /bin/bash -c "dokku storage:mount --replace $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "storage:unmount --all" -1
}

@test "(storage:unmount) removes multiple mounts in one call" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-logs"
  assert_success

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage rdmtest-cache:/cache rdmtest-logs:/logs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-data rdmtest-logs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].entry_name' | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-cache"

  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-cache rdmtest-missing"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].entry_name' | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-cache"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-data --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-cache --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-logs --force"
  assert_success
}

@test "(storage:unmount) --all removes every mount and is idempotent" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage rdmtest-cache:/cache"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP rdmtest-data"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "cannot be specified with --all" -1

  run /bin/bash -c "dokku storage:destroy rdmtest-data --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-cache --force"
  assert_success
}

@test "(storage:unmount) --all --process-type removes only that scope" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-data --container-dir /app/storage"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-cache --container-dir /cache --process-type web"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP --process-type web"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].entry_name' | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-data"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP --process-type worker"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-data --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-cache --force"
  assert_success
}
