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

# covers https://github.com/dokku/dokku/issues/4665
@test "(proxied-app) default" {
  run /bin/bash -c "dokku builder:set $TEST_APP selected null"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler:set $TEST_APP selected null"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:set $TEST_APP static-web-listener 127.0.0.1:8080"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:ports-set $TEST_APP http:80:8080"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "server 127.0.0.1:8080;"
}
