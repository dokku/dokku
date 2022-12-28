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

@test "(ps:rebuild) old app name" {
  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  mkdir -p /home/dokku/test_app
  sudo chown -R dokku:dokku /home/dokku/test_app

  run /bin/bash -c "dokku plugin:trigger post-create test_app"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:sync --build test_app https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild test_app"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:rename test_app $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(ps:scale) console-only app" {
  run /bin/bash -c "dokku ps:scale $TEST_APP web=0 console=0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:report $TEST_APP --deployed"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run deploy_app python-console-only
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:report $TEST_APP --deployed"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku run $TEST_APP console"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Hello world!"

  run /bin/bash -c "dokku run $TEST_APP printenv FOO"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku config:set $TEST_APP FOO=bar"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Releasing $TEST_APP"

  run /bin/bash -c "dokku run $TEST_APP printenv FOO"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "bar"
}

@test "(ps:set) procfile-path" {
  run deploy_app dockerfile-procfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Arg: web.py"
  assert_output_contains "Arg: second.Procfile" 0
  assert_success

  run /bin/bash -c "dokku ps:set $TEST_APP procfile-path nonexistent-procfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Could not start due to"

  run /bin/bash -c "dokku ps:set $TEST_APP procfile-path second.Procfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'SECRET_KEY:' 0

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Arg: web.py"
  assert_output_contains "Arg: second.Procfile"
  assert_success

  run /bin/bash -c "dokku ps:set $TEST_APP procfile-path"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'SECRET_KEY:'

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Arg: web.py"
  assert_output_contains "Arg: second.Procfile" 0
  assert_success
}

@test "(ps:scale) remove zerod processes" {
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

  run /bin/bash -c "dokku --quiet ps:scale $TEST_APP"
  output=$(echo "$output" | tr -s " ")
  echo "output: ($output)"
  assert_output $'cron: 0\ncustom: 0\nrelease: 0\nweb: 1\nworker: 1'

  run /bin/bash -c "dokku ps:scale $TEST_APP worker=0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ps:scale $TEST_APP"
  output=$(echo "$output" | tr -s " ")
  echo "output: ($output)"
  assert_output $'cron: 0\ncustom: 0\nrelease: 0\nweb: 1\nworker: 0'

  run /bin/bash -c "dokku ps:set $TEST_APP procfile-path second.Procfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ps:scale $TEST_APP"
  output=$(echo "$output" | tr -s " ")
  echo "output: ($output)"
  assert_output 'web: 1'
}
