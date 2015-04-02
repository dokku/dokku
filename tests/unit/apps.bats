#!/usr/bin/env bats

load test_helper

@test "(apps) apps" {
  create_app
  run bash -c "dokku apps | grep $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_output $TEST_APP
  destroy_app
}

@test "(apps) apps:create" {
  run dokku apps:create $TEST_APP
  echo "output: "$output
  echo "status: "$status
  assert_success
  destroy_app
}

@test "(apps) apps:destroy" {
  create_app
  run bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
