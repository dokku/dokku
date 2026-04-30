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

@test "(storage) storage:help" {
  run /bin/bash -c "dokku storage"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage mounted volumes"
  help_output="$output"

  run /bin/bash -c "dokku storage:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage mounted volumes"
  assert_output "$help_output"
}

@test "(storage) storage:ensure-directory" {
  run /bin/bash -c "test -d $DOKKU_LIB_ROOT/data/storage/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku storage:ensure-directory @"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP/"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 0
  assert_output_contains "Setting directory ownership to 2000:2000" 0
  assert_output_contains "Setting directory ownership to 32767:32767" 1

  run /bin/bash -c "test -d $DOKKU_LIB_ROOT/data/storage/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:ensure-directory --chown false $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 0
  assert_output_contains "Setting directory ownership to 2000:2000" 0
  assert_output_contains "Setting directory ownership to 32767:32767" 0

  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP --chown false"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 0
  assert_output_contains "Setting directory ownership to 2000:2000" 0
  assert_output_contains "Setting directory ownership to 32767:32767" 0

  run /bin/bash -c "dokku storage:ensure-directory --chown heroku $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 1
  assert_output_contains "Setting directory ownership to 2000:2000" 0
  assert_output_contains "Setting directory ownership to 32767:32767" 0

  run /bin/bash -c "dokku storage:ensure-directory --chown paketo $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 0
  assert_output_contains "Setting directory ownership to 2000:2000" 1
  assert_output_contains "Setting directory ownership to 32767:32767" 0

  run /bin/bash -c "dokku storage:ensure-directory --chown herokuish $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 0
  assert_output_contains "Setting directory ownership to 2000:2000" 0
  assert_output_contains "Setting directory ownership to 32767:32767" 1
}

@test "(storage) storage:mount, storage:list, storage:umount" {
  run /bin/bash -c "dokku storage:mount $TEST_APP /tmp/mount:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet storage:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/tmp/mount:/mount"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].host_path'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/tmp/mount"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].container_path'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/mount"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].volume_options'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku storage:mount $TEST_APP /tmp/mount:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_output " !     Mount path already exists."
  assert_failure

  run /bin/bash -c "dokku storage:unmount $TEST_APP /tmp/mount:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet storage:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku storage:unmount $TEST_APP /tmp/mount:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_output " !     Mount path does not exist."
  assert_failure

  run /bin/bash -c "dokku storage:mount $TEST_APP mount_volume:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet storage:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "mount_volume:/mount"
}

@test "(storage) storage:create / storage:list-entries / storage:destroy" {
  run /bin/bash -c "dokku storage:create rdmtest-entry"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-entry$'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-entry"

  run /bin/bash -c "dokku storage:info rdmtest-entry --format json | jq -r '.scheduler'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker-local"

  run /bin/bash -c "dokku storage:destroy rdmtest-entry"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(storage) storage:create rejects invalid names" {
  # underscore: rejected
  run /bin/bash -c "dokku storage:create rdmtest_invalid"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  # uppercase: rejected
  run /bin/bash -c "dokku storage:create RdmTest"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  # 46 chars: too long
  long_name=$(printf 'a%.0s' {1..46})
  run /bin/bash -c "dokku storage:create $long_name"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  # legacy- prefix: reserved
  run /bin/bash -c "dokku storage:create legacy-foo"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(storage) storage:create + storage:mount with named entry attaches multiple entries to one app" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-data --container-dir /data"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-cache --container-dir /cache"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].entry_name' | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-cache rdmtest-data"

  # cleanup
  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku storage:destroy rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-cache"
  assert_success
}

@test "(storage) storage:destroy refuses to remove a still-mounted entry" {
  run /bin/bash -c "dokku storage:create rdmtest-busy"
  assert_success
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-busy --container-dir /data"
  assert_success

  run /bin/bash -c "dokku storage:destroy rdmtest-busy"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "still mounted"

  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-busy"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-busy"
  assert_success
}

@test "(storage) storage:exec runs a non-interactive command and propagates exit code" {
  run /bin/bash -c "dokku storage:create rdmtest-exec"
  assert_success

  # docker pull image up-front so the test isn't gated on registry latency.
  run /bin/bash -c "docker pull alpine:3 >/dev/null"
  assert_success

  run /bin/bash -c "dokku storage:exec rdmtest-exec -- /bin/sh -c 'touch /data/marker && ls /data/marker'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "/data/marker"

  run /bin/bash -c "dokku storage:exec rdmtest-exec -- /bin/sh -c 'exit 42'"
  echo "output: $output"
  echo "status: $status"
  assert_equal "$status" 42

  run /bin/bash -c "dokku storage:destroy rdmtest-exec"
  assert_success
}

@test "(storage) storage:migrate converts legacy docker-options -v lines into attachments" {
  # Stage a legacy `-v` line directly via docker-options:add. This
  # bypasses the new storage:mount path so the line lives in
  # docker-options without a corresponding attachment, simulating an
  # app upgraded from a pre-PR Dokku version.
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy,run \"-v /tmp/legacy-mount:/legacy\""
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Before migration, storage:list shows nothing (attachment-only) and
  # docker-options has the -v line.
  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy"
  assert_success
  assert_output_contains "-v /tmp/legacy-mount:/legacy"

  # Run the migration for this app.
  run /bin/bash -c "dokku storage:migrate $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # storage:list now surfaces the synthesized colon form.
  run /bin/bash -c "dokku --quiet storage:list $TEST_APP"
  assert_success
  assert_output "/tmp/legacy-mount:/legacy"

  # The synthesized entry shows up under the legacy- prefix.
  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^legacy-' | head -1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # docker-options no longer holds the -v line on either phase.
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy"
  assert_success
  assert_output_not_contains "-v /tmp/legacy-mount:/legacy"

  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-run"
  assert_success
  assert_output_not_contains "-v /tmp/legacy-mount:/legacy"

  # Idempotency: re-running storage:migrate is a no-op (still one
  # attachment, no errors).
  run /bin/bash -c "dokku storage:migrate $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  assert_success
  assert_output "1"

  # Cleanup: unmount + destroy the synthesized legacy entry. We need
  # the entry name from list-entries since legacy-<hash> is content-derived.
  legacy_name=$(dokku storage:list-entries --format json | jq -r '.[] | select(.name | startswith("legacy-")) | .name' | head -1)
  if [[ -n "$legacy_name" ]]; then
    run /bin/bash -c "dokku storage:unmount $TEST_APP $legacy_name --container-dir /legacy"
    assert_success
    run /bin/bash -c "dokku storage:destroy $legacy_name"
    assert_success
  fi
}

@test "(storage) storage:ensure-directory emits deprecation warning" {
  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "deprecated"
}
