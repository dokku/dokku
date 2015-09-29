#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(build-env) special characters" {
  run dokku config:set --no-restart $TEST_APP NEWRELIC_APP_NAME="$TEST_APP (Staging)"
  echo "output: "$output
  echo "status: "$status
  assert_success
  deploy_app
  run dokku config $TEST_APP
  assert_success
}

@test "(build-env) failure" {
  run dokku config:set --no-restart $TEST_APP BUILDPACK_URL='https://github.com/dokku/fake-buildpack'
  echo "output: "$output
  echo "status: "$status
  assert_success
  run deploy_app
  echo "output: "$output
  echo "status: "$status
  assert_failure
}
