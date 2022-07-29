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

@test "(ps) herokuish" {
  deploy_app
  run /bin/bash -c "dokku ps:stop $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(<$CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_failure
  done

  run /bin/bash -c "dokku ps:start $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(<$CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done

  run /bin/bash -c "dokku ps:restart $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(<$CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(<$CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done
}
