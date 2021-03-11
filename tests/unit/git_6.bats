#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  touch /home/dokku/.ssh/known_hosts
  chown dokku:dokku /home/dokku/.ssh/known_hosts
  touch /home/dokku/data/git/$TEST_APP
}

teardown() {
  rm -f /home/dokku/.ssh/id_rsa.pub || true
  destroy_app
  global_teardown
}

@test "(git) git:unlock [succes]" {
  run /bin/bash -c "dokku git:unlock $TEST_APP --force" 
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:unlock [missing]" {
  run /bin/bash -c "dokku git:unlock"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
