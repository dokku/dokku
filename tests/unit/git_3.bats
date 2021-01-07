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
