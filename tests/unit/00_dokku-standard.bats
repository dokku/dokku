#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  rm -rf /home/dokku/$TEST_APP/tls /home/dokku/tls
  destroy_app
  disable_tls_wildcard
}

@test "run (with tty)" {
  deploy_app
  run /bin/bash -c "dokku run $TEST_APP ls /app/package.json"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "run (without tty)" {
  deploy_app
  run /bin/bash -c ": |dokku run $TEST_APP ls /app/package.json"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "urls (non-ssl)" {
  run bash -c "dokku urls $TEST_APP | grep dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_output "http://dokku.me"
}

@test "urls (app ssl)" {
  setup_test_tls
  run bash -c "dokku urls $TEST_APP | grep dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_output "https://dokku.me"
}

@test "urls (wildcard ssl)" {
  setup_test_tls_wildcard
  run bash -c "dokku urls $TEST_APP | grep dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_output "https://dokku.me"
}
