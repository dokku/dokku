#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  destroy_app 0 "$MYAPP" || true
  global_teardown
}

# @test "(ps) herokuish base" {
#   # CI support: 'Ah. I just spoke with our Docker expert --
#   # looks like docker exec is built to work with docker-under-libcontainer,
#   # but we're using docker-under-lxc. I don't have an estimated time for the fix, sorry
#   skip "circleci does not support docker exec at the moment."
#   deploy_app
#   run /bin/bash -c "dokku ps $TEST_APP | grep -q \"node web.js\""
#   echo "output: $output"
#   echo "status: $status"
#   assert_success
# }

@test "(ps) herokuish" {
  deploy_app
  run /bin/bash -c "dokku ps:stop $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(< $CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_failure
  done

  run /bin/bash -c "dokku ps:start $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(< $CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done

  run /bin/bash -c "dokku ps:restart $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(< $CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(< $CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done
}

@test "(ps:scale) herokuish" {
  run /bin/bash -c "dokku ps:scale $TEST_APP web=2 worker=2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  deploy_app
  for PROC_TYPE in web worker; do
    CIDS=""
    for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.$PROC_TYPE.*; do
      CIDS+=$(< $CID_FILE)
      CIDS+=" "
    done
    CIDS_PATTERN=$(echo $CIDS | sed -e "s: :|:g")
    run /bin/bash -c "docker ps -q --no-trunc | grep -E \"$CIDS_PATTERN\" | wc -l | grep 2"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 worker=1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for PROC_TYPE in web worker; do
    CIDS=""
    for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.$PROC_TYPE.*; do
      CIDS+=$(< $CID_FILE)
      CIDS+=" "
    done
    CIDS_PATTERN=$(echo $CIDS | sed -e "s: :|:g")
    run /bin/bash -c "docker ps -q --no-trunc | grep -E \"$CIDS_PATTERN\" | wc -l | grep 1"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done

  run /bin/bash -c "dokku ps:scale $TEST_APP web=0 worker=0 invalid-process=0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for PROC_TYPE in web worker invalid-process; do
    CIDS=""
    shopt -s nullglob
    for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.$PROC_TYPE.*; do
      CIDS+=$(< $CID_FILE)
      CIDS+=" "
    done
    shopt -u nullglob
    run /bin/bash -c "[[ -z \"$CIDS\" ]]"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done
}

@test "(ps:restore) herokuish" {
  MYAPP="manual-randomtestapp-1"
  create_app "$MYAPP"
  deploy_app

  run /bin/bash -c "dokku ps:stop $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:list"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:restore"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ps:report $TEST_APP | grep -q exited"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
