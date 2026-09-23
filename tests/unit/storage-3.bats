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

@test "(storage:mount) --replace varies mount-time fields per spec" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage:Z,volume-subpath=uploads,volume-chown=herokuish rdmtest-cache:/cache:ro,noexec,phase=deploy --process-type web"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.phases\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "deploy,run"

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

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.2.phases\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "deploy"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.2.subpath\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.2.volume-chown\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

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

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.process-type\", .\"attachment.2.process-type\"' | sort -u | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "web"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-data --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-cache --force"
  assert_success
}

@test "(storage:mount) --replace rejects mount-time flags" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-data --container-dir /app/storage"
  echo "output: $output"
  echo "status: $status"
  assert_success

  for flag in "--phase deploy" "--volume-subpath uploads" "--volume-readonly" "--volume-chown herokuish" "--volume-options Z"; do
    run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-cache:/cache $flag"
    echo "flag: $flag"
    echo "output: $output"
    echo "status: $status"
    assert_failure
    assert_output_contains "cannot be used with --replace" -1
  done

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

@test "(storage:mount) --replace rejects malformed spec options" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage:volume-subpth=uploads"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "unknown key" -1

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage:volume-subpath="
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "empty value" -1

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage:volume-subpath=a,volume-subpath=b"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "more than once" -1

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage:ro,rw"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "sets both ro and rw" -1

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage:phase=build"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "unsupported phase" -1

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage:volume-chown=bogus"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Unsupported chown permissions" -1

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku storage:destroy rdmtest-data --force"
  assert_success
}

@test "(storage:mount) --replace rw leaves the option list alone" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:mount --replace $TEST_APP rdmtest-data:/app/storage:rw rdmtest-cache:/cache:rw,Z"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.readonly\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.volume-options\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.2.volume-options\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "Z"

  run /bin/bash -c "dokku storage:report $TEST_APP --storage-deploy-mounts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "rw"

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
