#!/usr/bin/env bats

load test_helper

setup() {
  create_app
  setup_test_tls
  deploy_app
}

teardown() {
  destroy_app
}

@test "nginx:build-config (with SSL CN mismatch)" {
  run /bin/bash -c "dokku domains $TEST_APP | grep node-js-app.dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_output "node-js-app.dokku.me"
}
