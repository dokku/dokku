#!/usr/bin/env bats

load test_helper

setup() {
  deploy_app
}

teardown() {
  destroy_app
}

@test "ps" {
  skip "circleci does not support docker exec at the moment"
  run bash -c "dokku ps $TEST_APP | grep -q \"node web.js\""
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "ps:start" {
  run bash -c "dokku ps:stop $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "dokku ps:start $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "docker ps -q --no-trunc | grep -q $(< $DOKKU_ROOT/$TEST_APP/CONTAINER)"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "ps:stop" {
  run bash -c "dokku ps:stop $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "docker ps -q --no-trunc | grep -q $(< $DOKKU_ROOT/$TEST_APP/CONTAINER)"
  echo "output: "$output
  echo "status: "$status
  assert_failure
}

@test "ps:restart" {
  run bash -c "dokku ps:restart $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "docker ps -q --no-trunc | grep -q $(< $DOKKU_ROOT/$TEST_APP/CONTAINER)"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
