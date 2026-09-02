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

@test "(ps:scale) --replace scales down unspecified processes" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP worker=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker ps --filter label=com.dokku.app-name=$TEST_APP --filter label=com.dokku.process-type=worker --format '{{.Names}}' | wc -l"
  output=$(echo "$output" | tr -d " ")
  echo "output: ($output)"
  echo "status: $status"
  assert_output "1"

  run /bin/bash -c "dokku ps:scale --replace $TEST_APP web=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ps:scale $TEST_APP"
  output=$(echo "$output" | tr -s " ")
  echo "output: ($output)"
  echo "status: $status"
  assert_output $'cron: 0\ncustom: 0\nrelease: 0\ntask: 0\nweb: 1\nworker: 0'

  sleep 2

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker ps --filter label=com.dokku.app-name=$TEST_APP --filter label=com.dokku.process-type=worker --format '{{.Names}}' | wc -l"
  output=$(echo "$output" | tr -d " ")
  echo "output: ($output)"
  echo "status: $status"
  assert_output "0"
}

@test "(ps:scale) --clear without a web process type" {
  run /bin/bash -c "dokku ps:scale $TEST_APP web=0 console=0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python-console-only
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale --skip-deploy $TEST_APP console=2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ps:scale $TEST_APP"
  output=$(echo "$output" | tr -s " ")
  echo "output: ($output)"
  echo "status: $status"
  assert_output "console: 2"

  run /bin/bash -c "dokku ps:scale --clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ps:scale $TEST_APP"
  output=$(echo "$output" | tr -s " ")
  echo "output: ($output)"
  echo "status: $status"
  assert_output "console: 0"
}
