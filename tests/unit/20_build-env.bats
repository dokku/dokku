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
  run dokku config:set --no-restart $TEST_APP NEWRELIC_APP_NAME="$TEST_APP (Staging)"
  echo "output: "$output
  echo "status: "$status
  assert_success

  deploy_app
  run dokku config $TEST_APP
  assert_success
}

@test "(build-env) default curl timeouts" {
  run dokku config:unset --global CURL_CONNECT_TIMEOUT
  echo "output: "$output
  echo "status: "$status
  assert_success

  run dokku config:unset --global CURL_TIMEOUT
  echo "output: "$output
  echo "status: "$status
  assert_success

  deploy_app
  run /bin/bash -c "dokku config:get --global CURL_CONNECT_TIMEOUT | grep 90"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run /bin/bash -c "dokku config:get --global CURL_TIMEOUT | grep 60"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(build-env) buildpack failure" {
  run dokku config:set --no-restart $TEST_APP BUILDPACK_URL='https://github.com/dokku/fake-buildpack'
  echo "output: "$output
  echo "status: "$status
  assert_success

  run deploy_app
  echo "output: "$output
  echo "status: "$status
  assert_failure
}

@test "(build-env) buildpack deploy with Dockerfile" {
  run dokku config:set --no-restart $TEST_APP BUILDPACK_URL='https://github.com/heroku/heroku-buildpack-nodejs'
  echo "output: "$output
  echo "status: "$status
  assert_success

  deploy_app dockerfile
  run dokku --quiet config:get $TEST_APP DOKKU_APP_TYPE
  echo "output: "$output
  echo "status: "$status
  assert_output "herokuish"
}

@test "(build-env) app autocreate disabled" {
  run dokku config:set --no-restart --global DOKKU_DISABLE_APP_AUTOCREATION='true'
  echo "output: "$output
  echo "status: "$status
  assert_success

  run deploy_app
  echo "output: "$output
  echo "status: "$status
  assert_failure
}
