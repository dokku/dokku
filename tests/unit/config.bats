#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "config:set" {
  run dokku config:set $TEST_APP test_var=true test_var2='hello world'
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "config:get" {
  run dokku config:set $TEST_APP test_var=true test_var2='hello world'
  echo "output: "$output
  echo "status: "$status
  run dokku config:get $TEST_APP test_var2
  echo "output: "$output
  echo "status: "$status
  assert_output 'hello world'
}

@test "config:unset" {
  run dokku config:set $TEST_APP test_var=true test_var2='hello world'
  echo "output: "$output
  echo "status: "$status
  run dokku config:get $TEST_APP test_var
  echo "output: "$output
  echo "status: "$status
  run dokku config:unset $TEST_APP test_var
  echo "output: "$output
  echo "status: "$status
  run dokku config:get $TEST_APP test_var
  echo "output: "$output
  echo "status: "$status
  assert_output ""
}
