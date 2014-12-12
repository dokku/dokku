#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "config:set" {
  run dokku config:set $TEST_APP test_var=true
  assert_success
}

@test "config:get" {
  run dokku config:set $TEST_APP test_var=true
  run dokku config:get $TEST_APP test_var
  assert_output true
}

@test "config:unset" {
  run dokku config:set $TEST_APP test_var=true
  run dokku config:get $TEST_APP test_var
  run dokku config:unset $TEST_APP test_var
  run dokku config:get $TEST_APP test_var
  assert_output ""
}
