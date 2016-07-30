#!/usr/bin/env bats

load test_helper

setup() {
  create_key
}

teardown() {
  destroy_key
}

@test "(ssh-keys) ssh-keys:add, ssh-keys:list, ssh-keys:remove" {
  run /bin/bash -c "dokku ssh-keys:add name1 /tmp/testkey.pub"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "cat /tmp/testkey.pub | dokku ssh-keys:add name2"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku ssh-keys:list | grep name1 && dokku ssh-keys:list | grep name2"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku ssh-keys:remove name1"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku ssh-keys:list | grep name1"
  echo "output: "$output
  echo "status: "$status
  assert_failure
  run /bin/bash -c "dokku ssh-keys:list | grep name2"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
