#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
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

  run /bin/bash -c "dokku git:clone $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:clone $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(git) git:clone [--no-build]" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:clone $TEST_APP https://github.com/heroku/node-js-getting-started.git"
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

  run /bin/bash -c "dokku git:clone $TEST_APP https://github.com/heroku/node-js-getting-started.git main"
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

  run /bin/bash -c "dokku git:clone $TEST_APP https://github.com/heroku/node-js-getting-started.git 1"
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

  run /bin/bash -c "dokku git:clone $TEST_APP https://github.com/heroku/node-js-getting-started.git 97e6c72491c7531507bfc5413903e0e00e31e1b0"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:clone [--build]" {
  run /bin/bash -c "dokku git:clone --build $TEST_APP https://github.com/heroku/node-js-getting-started.git"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:clone --build $TEST_APP https://github.com/heroku/node-js-getting-started.git main"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:clone --build $TEST_APP https://github.com/heroku/node-js-getting-started.git 1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:clone --build $TEST_APP https://github.com/heroku/node-js-getting-started.git 97e6c72491c7531507bfc5413903e0e00e31e1b0"
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
