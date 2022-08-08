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
  echo -n "$KEY" >/tmp/testkey-no-newline.pub

  # the temporary key is useful for adding in the file with two keys
  # useful for a negative test
  {
    cat /tmp/testkey.pub
    echo "$KEY"
  } >/tmp/testkey-double.pub

  # another negative test input
  echo 'invalid!' >/tmp/testkey-invalid.pub

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

@test "(ssh-keys) ssh-keys:help" {
  run /bin/bash -c "dokku ssh-keys"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage public ssh keys used for deployment"
  help_output="$output"

  run /bin/bash -c "dokku ssh-keys:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage public ssh keys used for deployment"
  assert_output "$help_output"
}

@test "(ssh-keys) ssh-keys:add" {
  run /bin/bash -c "dokku ssh-keys:add name1 /tmp/testkey-double.pub"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ssh-keys:add name2 /tmp/testkey-invalid.pub"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ssh-keys:add name3 /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list name3"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:add name4 /tmp/testkey-no-newline.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list name4"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(ssh-keys) ssh-keys:add FILE" {
  run /bin/bash -c "dokku ssh-keys:add name /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:add name /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "cat /tmp/testkey.pub | dokku ssh-keys:add name"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ssh-keys:add other-name /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "cat /tmp/testkey.pub | dokku ssh-keys:add other-name"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ssh-keys:list name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list other-name"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(ssh-keys) ssh-keys:add stdin" {
  run /bin/bash -c "cat /tmp/testkey.pub | dokku ssh-keys:add name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:add name /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "cat /tmp/testkey.pub | dokku ssh-keys:add name"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ssh-keys:add other-name /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "cat /tmp/testkey.pub | dokku ssh-keys:add other-name"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ssh-keys:list name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list other-name"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(ssh-keys) ssh-keys:add invalid" {
  run /bin/bash -c 'echo invalid >> "${DOKKU_ROOT:-/home/dokku}/.ssh/authorized_keys"'
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:add name5 /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ssh-keys:list"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(ssh-keys) ssh-keys:remove" {
  run /bin/bash -c "dokku ssh-keys:add new-user /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list new-user | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "1"

  run /bin/bash -c "dokku ssh-keys:remove new-user"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list new-user | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_equal "$output" "0"

  run /bin/bash -c "dokku ssh-keys:remove new-user"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:add new-user /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list new-user | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_equal "$output" "1"

  run /bin/bash -c "dokku ssh-keys:list new-user | cut -d' ' -f1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  local fingerprint="$output"
  run /bin/bash -c "dokku ssh-keys:remove --fingerprint ${fingerprint}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list new-user | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "0"
}

@test "(ssh-keys) ssh-keys:list" {
  run /bin/bash -c "dokku ssh-keys:list"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c 'echo "" >> "${DOKKU_ROOT:-/home/dokku}/.ssh/authorized_keys"'
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list | grep name1"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ssh-keys:list name1"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c 'echo "" > "${DOKKU_ROOT:-/home/dokku}/.ssh/authorized_keys"'
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:add new-user /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "1"

  run /bin/bash -c "dokku ssh-keys:add another-user /tmp/testkey-no-newline.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "2"

  run /bin/bash -c "dokku ssh-keys:list new-user | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "1"
}

@test "(ssh-keys) ssh-keys:list --format invalid" {
  run /bin/bash -c "dokku ssh-keys:list --format invalid"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(ssh-keys) ssh-keys:list --format text" {
  run /bin/bash -c "dokku ssh-keys:list --format text"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c 'echo "" >> "${DOKKU_ROOT:-/home/dokku}/.ssh/authorized_keys"'
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list --format text"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list --format text | grep name1"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ssh-keys:list --format text name1"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c 'echo "" > "${DOKKU_ROOT:-/home/dokku}/.ssh/authorized_keys"'
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:add new-user /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list --format text | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "1"

  run /bin/bash -c "dokku ssh-keys:add another-user /tmp/testkey-no-newline.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list --format text | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "2"

  run /bin/bash -c "dokku ssh-keys:list --format text new-user | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "1"
}

@test "(ssh-keys) ssh-keys:list --format json" {
  run /bin/bash -c "dokku ssh-keys:list --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c 'echo "" >> "${DOKKU_ROOT:-/home/dokku}/.ssh/authorized_keys"'
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list --format json name1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "[]"

  run /bin/bash -c 'echo "" > "${DOKKU_ROOT:-/home/dokku}/.ssh/authorized_keys"'
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:add new-user /tmp/testkey.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list --format json | jq length"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "1"

  run /bin/bash -c "dokku ssh-keys:add another-user /tmp/testkey-no-newline.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ssh-keys:list --format json | jq length"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "2"

  run /bin/bash -c "dokku ssh-keys:list --format json new-user | jq length"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "1"
}
