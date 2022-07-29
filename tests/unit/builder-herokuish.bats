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
