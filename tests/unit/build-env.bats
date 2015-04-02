#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(build-env) special characters" {
  run dokku config:set $TEST_APP NEWRELIC_APP_NAME="$TEST_APP (Staging)"
  echo "output: "$output
  echo "status: "$status
  assert_success
  deploy_app
  run dokku config $TEST_APP
  assert_success
}
