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

@test "(resource) resource:limit" {
  run /bin/bash -c "dokku resource:limit $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "resource limits $TEST_APP information"

  deploy_app
  run /bin/bash -c "dokku resource:limit --memory 512MB $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.Memory}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "0"

  dokku ps:rebuild "$TEST_APP"
  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.Memory}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "536870912"

  run /bin/bash -c "dokku resource:limit --memory 1024MB --process-type worker $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku ps:rebuild "$TEST_APP"
  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.Memory}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "536870912"

  run /bin/bash -c "dokku resource:limit-clear --process-type worker $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku ps:rebuild "$TEST_APP"
  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.Memory}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "536870912"

  run /bin/bash -c "dokku resource:limit-clear --process-type web $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku ps:rebuild "$TEST_APP"
  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.Memory}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "536870912"

  run /bin/bash -c "dokku resource:limit-clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku ps:rebuild "$TEST_APP"
  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.Memory}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "0"
}

@test "(resource) resource:reserve" {
  run /bin/bash -c "dokku resource:reserve $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "resource reservation $TEST_APP information"

  deploy_app
  run /bin/bash -c "dokku resource:reserve --memory 512MB $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.MemoryReservation}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "0"

  dokku ps:rebuild "$TEST_APP"
  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.MemoryReservation}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "536870912"

  run /bin/bash -c "dokku resource:reserve --memory 1024MB --process-type worker $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku ps:rebuild "$TEST_APP"
  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.MemoryReservation}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "536870912"

  run /bin/bash -c "dokku resource:reserve-clear --process-type worker $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku ps:rebuild "$TEST_APP"
  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.MemoryReservation}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "536870912"

  run /bin/bash -c "dokku resource:reserve-clear --process-type web $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku ps:rebuild "$TEST_APP"
  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.MemoryReservation}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "536870912"

  run /bin/bash -c "dokku resource:reserve-clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku ps:rebuild "$TEST_APP"
  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect --format '{{.HostConfig.MemoryReservation}}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "0"
}

@test "(resource) resource:limit-clear" {
  run /bin/bash -c "dokku resource:limit-clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku resource:limit-clear --process-type web $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(resource) resource:reserve-clear" {
  run /bin/bash -c "dokku resource:reserve-clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku resource:reserve-clear --process-type web $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
