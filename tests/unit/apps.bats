#!/usr/bin/env bats

load test_helper

APP=lifecycle-app

@test "apps:create" {
  run dokku apps:create $APP
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "apps" {
  run bash -c "dokku apps | grep $APP"
  echo "output: "$output
  echo "status: "$status
  assert_output $APP
}

@test "apps:destroy" {
  run bash -c "dokku --force apps:destroy $APP"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
