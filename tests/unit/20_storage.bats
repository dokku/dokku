#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
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
  assert_output "Mount path already exists."
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
  assert_output "Mount path does not exist."
  assert_failure
}
