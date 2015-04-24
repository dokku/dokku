#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(domains) domains" {
  run bash -c "dokku domains $TEST_APP | grep ${TEST_APP}.dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_output "${TEST_APP}.dokku.me"
}

@test "(domains) domains:add" {
  run dokku domains:add $TEST_APP www.test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku domains:add $TEST_APP test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(domains) domains:add (multiple)" {
  run dokku domains:add $TEST_APP www.test.app.dokku.me test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(domains) domains:remove" {
  run dokku domains:add $TEST_APP test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku domains:remove $TEST_APP test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  refute_line "test.app.dokku.me"
}

@test "(domains) domains:clear" {
  run dokku domains:add $TEST_APP test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku domains:clear $TEST_APP
  echo "output: "$output
  echo "status: "$status
  assert_success
}
