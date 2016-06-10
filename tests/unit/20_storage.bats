#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(storage) storage:mount, storage:list, storage:umount" {
  run /bin/bash -c "dokku storage:mount $TEST_APP /tmp/mount:/mount"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku storage:list $TEST_APP | grep -qe '^\s*/tmp/mount:/mount'"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku storage:mount $TEST_APP /tmp/mount:/mount"
  echo "output: "$output
  echo "status: "$status
  assert_output "Mount path already exists."
  assert_failure
  run /bin/bash -c "dokku storage:unmount $TEST_APP /tmp/mount:/mount"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku storage:list $TEST_APP | grep -vqe '^\s*/tmp/mount:/mount'"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku storage:unmount $TEST_APP /tmp/mount:/mount"
  echo "output: "$output
  echo "status: "$status
  assert_output "Mount path does not exist."
  assert_failure
}

@test "(storage) storage:mount (dockerfile)" {
  run /bin/bash -c "dokku storage:mount $TEST_APP /tmp/mount:/mount"
  echo "output: "$output
  echo "status: "$status
  assert_success

  deploy_app dockerfile

  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect -f '{{ .HostConfig.Binds }}' $CID"
  echo "output: "$output
  echo "status: "$status
  assert_output "[/tmp/mount:/mount]"
}
