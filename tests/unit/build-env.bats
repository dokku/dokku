#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(build-env) special characters" {
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP NEWRELIC_APP_NAME='$TEST_APP (Staging)'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config $TEST_APP"
  assert_success
}

@test "(build-env) default curl timeouts" {
  run /bin/bash -c "dokku config:unset --global CURL_CONNECT_TIMEOUT"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:unset --global CURL_TIMEOUT"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config:get --global CURL_CONNECT_TIMEOUT | grep 90"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get --global CURL_TIMEOUT | grep 600"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(build-env) buildpack failure" {
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP BUILDPACK_URL='https://github.com/dokku/fake-buildpack'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(build-env) buildpack deploy with Dockerfile" {
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP BUILDPACK_URL='https://github.com/dokku/heroku-buildpack-null'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP move_dockerfile_into_place
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-detected"
  echo "output: $output"
  echo "status: $status"
  assert_output "herokuish"
}
