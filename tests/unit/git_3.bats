#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  touch /home/dokku/.ssh/known_hosts
  chown dokku:dokku /home/dokku/.ssh/known_hosts
}

teardown() {
  rm -f /home/dokku/.ssh/id_rsa.pub || true
  destroy_app
  global_teardown
}

@test "(git) git:allow-host" {
  run /bin/bash -c "dokku git:allow-host"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "cat /home/dokku/.ssh/known_hosts | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  start_lines=$output

  run /bin/bash -c "dokku git:allow-host github.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/.ssh/known_hosts | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f /home/dokku/.ssh/known_hosts"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/.ssh/known_hosts | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "$((start_lines + 1))"

  run /bin/bash -c "dokku git:allow-host github.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/.ssh/known_hosts | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "$((start_lines + 2))"
}

@test "(git) git:clone [errors]" {
  run /bin/bash -c "dokku git:clone"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku git:clone $TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run create_app "$TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:clone $TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run destroy_app 0 "$TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:clone [--no-build]" {
  run /bin/bash -c "dokku git:clone $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:clone $TEST_APP https://github.com/dokku/smoke-test-app.git other-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:clone $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:clone $TEST_APP https://github.com/dokku/smoke-test-app.git 9cf71bba639c4f1671dfd42685338b762d3354f2"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:clone [--build noarg]" {
  run /bin/bash -c "dokku git:clone --build $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"
}

@test "(git) git:clone [--build branch]" {
  run /bin/bash -c "dokku git:clone --build $TEST_APP https://github.com/dokku/smoke-test-app.git other-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"
}

@test "(git) git:clone [--build tag]" {
  run /bin/bash -c "dokku git:clone --build $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"
}

@test "(git) git:clone [--build commit]" {
  run /bin/bash -c "dokku git:clone --build $TEST_APP https://github.com/dokku/smoke-test-app.git 9cf71bba639c4f1671dfd42685338b762d3354f2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"
}

@test "(git) git:public-key" {
  run /bin/bash -c "dokku git:public-key"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "cp /root/.ssh/dokku_test_rsa.pub /home/dokku/.ssh/id_rsa.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:public-key"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
