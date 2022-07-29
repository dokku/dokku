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

@test "(ps) dockerfile" {
  deploy_app dockerfile
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

@test "(ps:scale) dockerfile non-existent process" {
  run /bin/bash -c "dokku ps:scale $TEST_APP non-existent=2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  destroy_app
  create_app
  deploy_app dockerfile-procfile
  run /bin/bash -c "dokku ps:scale $TEST_APP non-existent=2"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(ps:scale) dockerfile" {
  run /bin/bash -c "dokku ps:scale $TEST_APP web=2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  deploy_app dockerfile
  CIDS=""
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    CIDS+=$(<$CID_FILE)
    CIDS+=" "
  done
  CIDS_PATTERN=$(echo $CIDS | sed -e "s: :|:g")
  run /bin/bash -c "docker ps -q --no-trunc | grep -E \"$CIDS_PATTERN\" | wc -l | grep 2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  CIDS=""
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    CIDS+=$(<$CID_FILE)
    CIDS+=" "
  done
  CIDS_PATTERN=$(echo $CIDS | sed -e "s: :|:g")
  run /bin/bash -c "docker ps -q --no-trunc | grep -E \"$CIDS_PATTERN\" | wc -l | grep 1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  CIDS=""
  shopt -s nullglob
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    CIDS+=$(<$CID_FILE)
    CIDS+=" "
  done
  run /bin/bash -c "[[ -z \"$CIDS\" ]]"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
