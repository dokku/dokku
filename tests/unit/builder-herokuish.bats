#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(builder-herouish:build .env)" {
  run deploy_app python dokku@dokku.me:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'DOTENV_KEY=some_value'
}

@test "(builder-herokuish) builder-herokuish:set allowed" {
  if [[ "$(dpkg --print-architecture 2>/dev/null || true)" == "amd64" ]]; then
    skip "this test cannot be performed accurately on amd64 as it tests whether we can enable the plugin on armhf/arm64"
  fi

  run /bin/bash -c "dokku builder-herokuish:set --global allowed"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:set --global allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:set --global allowed"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
