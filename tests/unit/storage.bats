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

  run /bin/bash -c "dokku storage:ensure-directory --chown packeto $TEST_APP"
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
  run /bin/bash -c "dokku storage:list $TEST_APP | grep -qe '^\s*/tmp/mount:/mount'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku storage:mount $TEST_APP /tmp/mount:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_output " !     Mount path already exists."
  assert_failure
  run /bin/bash -c "dokku storage:unmount $TEST_APP /tmp/mount:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku storage:list $TEST_APP | grep -vqe '^\s*/tmp/mount:/mount'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku storage:unmount $TEST_APP /tmp/mount:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_output " !     Mount path does not exist."
  assert_failure
  run /bin/bash -c "dokku storage:mount $TEST_APP mount_volume:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
