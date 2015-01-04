#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "nginx (no server tokens)" {
  deploy_app
  run /bin/bash -c "curl -s -D - $(dokku url $TEST_APP) -o /dev/null | egrep '^Server' | egrep '[0-9]+'"
  echo "output: "$output
  echo "status: "$status
  assert_failure
}

@test "nginx:build-config (with SSL CN mismatch)" {
  setup_test_tls
  deploy_app
  run /bin/bash -c "dokku domains $TEST_APP | grep node-js-app.dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_output "node-js-app.dokku.me"
}
