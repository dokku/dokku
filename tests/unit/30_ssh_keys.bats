#!/usr/bin/env bats

load test_helper

setup() {
  global_setup

  # create a temporary key and save in a variable
  create_key
  local KEY="$(cat /tmp/testkey.pub)"
  destroy_key

  # now create key that will be really used
  create_key

  # Test key without a trailing newline
  echo -n "$KEY" > /tmp/testkey-no-newline.pub

  # the temporary key is useful for adding in the file with two keys
  # useful for a negative test
  { cat /tmp/testkey.pub ; echo "$KEY" ; } > /tmp/testkey-double.pub

  # another negative test input
  echo 'invalid!' > /tmp/testkey-invalid.pub

  # save current authorized_keys to remove all changes afterwards
  cp "${DOKKU_ROOT:-/home/dokku}/.ssh/authorized_keys" /tmp/testkey-authorized_keys
}

teardown() {
  # restore authorized_keys to its contents before the tests
  cp /tmp/testkey-authorized_keys "${DOKKU_ROOT:-/home/dokku}/.ssh/authorized_keys"

  # normal cleanup after here
  destroy_key
  global_teardown
}

@test "(ssh-keys) ssh-keys:add, ssh-keys:list, ssh-keys:remove" {
  run /bin/bash -c "dokku ssh-keys:add name1 /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "cat /tmp/testkey.pub | dokku ssh-keys:add name2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku ssh-keys:list | grep name1 && dokku ssh-keys:list | grep name2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku ssh-keys:remove name1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c 'echo "" >> "${DOKKU_ROOT:-/home/dokku}/.ssh/authorized_keys"'
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku ssh-keys:list | grep name1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  run /bin/bash -c "dokku ssh-keys:list | grep name2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku ssh-keys:add name3 /tmp/testkey-double.pub"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  run /bin/bash -c "dokku ssh-keys:add name4 /tmp/testkey-invalid.pub"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  run /bin/bash -c "dokku ssh-keys:add name4 /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku ssh-keys:add name5 /tmp/testkey-no-newline.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # leave this as the last test in the sequence! It introduces an error in authorized_keys
  run /bin/bash -c 'echo invalid >> "${DOKKU_ROOT:-/home/dokku}/.ssh/authorized_keys"'
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku ssh-keys:add name5 /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
