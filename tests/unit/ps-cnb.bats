#!/usr/bin/env bats

load test_helper

setup_file() {
  install_pack
}

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(ps) cnb env vars" {
  run /bin/bash -c "dokku config:set $TEST_APP DOKKU_CNB_EXPERIMENTAL=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@dokku.me:$TEST_APP add_requirements_txt
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "curl $(dokku url $TEST_APP)/env"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains '"DOKKU_CNB_EXPERIMENTAL": "1"'
}
