#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

# @test "(ps) dockerfile" {
#   # CI support: 'Ah. I just spoke with our Docker expert --
#   # looks like docker exec is built to work with docker-under-libcontainer,
#   # but we're using docker-under-lxc. I don't have an estimated time for the fix, sorry
#   skip "circleci does not support docker exec at the moment."
#   run bash -c "dokku ps $TEST_APP | grep -q \"node web.js\""
#   echo "output: "$output
#   echo "status: "$status
#   assert_success
# }

@test "(ps) dockerfile" {
  deploy_app dockerfile
  run bash -c "dokku ps:stop $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run bash -c "docker ps -q --no-trunc | grep -q $(< $CID_FILE)"
    echo "output: "$output
    echo "status: "$status
    assert_failure
  done

  run bash -c "dokku ps:start $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run bash -c "docker ps -q --no-trunc | grep -q $(< $CID_FILE)"
    echo "output: "$output
    echo "status: "$status
    assert_success
  done

  run bash -c "dokku ps:restart $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run bash -c "docker ps -q --no-trunc | grep -q $(< $CID_FILE)"
    echo "output: "$output
    echo "status: "$status
    assert_success
  done

  run bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run bash -c "docker ps -q --no-trunc | grep -q $(< $CID_FILE)"
    echo "output: "$output
    echo "status: "$status
    assert_success
  done
}

@test "(ps:scale) dockerfile" {
  run bash -c "dokku ps:scale $TEST_APP web=2"
  echo "output: "$output
  echo "status: "$status
  assert_success

  deploy_app dockerfile
  CIDS=""
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    CIDS+=$(< $CID_FILE)
    CIDS+=" "
  done
  CIDS_PATTERN=$(echo $CIDS | sed -e "s: :|:g")
  run bash -c "docker ps -q --no-trunc | egrep \"$CIDS_PATTERN\" | wc -l | grep 2"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku ps:scale $TEST_APP web=1"
  echo "output: "$output
  echo "status: "$status
  assert_success
  CIDS=""
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    CIDS+=$(< $CID_FILE)
    CIDS+=" "
  done
  CIDS_PATTERN=$(echo $CIDS | sed -e "s: :|:g")
  run bash -c "docker ps -q --no-trunc | egrep \"$CIDS_PATTERN\" | wc -l | grep 1"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku ps:scale $TEST_APP web=0"
  echo "output: "$output
  echo "status: "$status
  assert_success
  CIDS=""
  shopt -s nullglob
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    CIDS+=$(< $CID_FILE)
    CIDS+=" "
  done
  run bash -c "[[ -z \"$CIDS\" ]]"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
