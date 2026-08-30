#!/usr/bin/env bats

load test_helper

PROFILE_PROPERTY_PATH="/var/lib/dokku/config/scheduler-k3s/--global"

setup() {
  global_setup
}

teardown() {
  rm -f "$PROFILE_PROPERTY_PATH"/node-profile-*.json || true
  global_teardown
}

@test "(scheduler-k3s:profiles:add) rejects names that cannot derive a valid helm release name" {
  run /bin/bash -c "dokku scheduler-k3s:profiles:add a-twenty-seven-char-profile --role worker"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Profile name is too long, must be at most 26 characters"

  run /bin/bash -c "dokku scheduler-k3s:profiles:add EdgePool --role worker"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "must only contain lowercase alphanumeric characters"

  run /bin/bash -c "dokku scheduler-k3s:profiles:list --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"
}

@test "(scheduler-k3s:profiles:add) accepts a name at the maximum length" {
  run /bin/bash -c "dokku scheduler-k3s:profiles:add aaaaaaaaaaaaaaaaaaaaaaaaaa --role worker"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:profiles:list --format json | jq -r '.[0].name'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "aaaaaaaaaaaaaaaaaaaaaaaaaa"

  run /bin/bash -c "dokku scheduler-k3s:profiles:remove aaaaaaaaaaaaaaaaaaaaaaaaaa"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:profiles:list --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"
}

@test "(scheduler-k3s:profiles:remove) removes a profile stored under a legacy name" {
  seed_legacy_node_profile EdgePool

  run /bin/bash -c "dokku scheduler-k3s:profiles:list --format json | jq -r '.[0].name'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "EdgePool"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --profile EdgePool vm.max_map_count 262144"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "cannot carry node sysctls"

  run /bin/bash -c "dokku scheduler-k3s:profiles:remove EdgePool"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:profiles:list --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"
}

seed_legacy_node_profile() {
  declare desc="writes a node profile property under a name profiles:add no longer accepts"
  declare PROFILE_NAME="$1"

  mkdir -p "$PROFILE_PROPERTY_PATH"
  echo "{\"name\":\"$PROFILE_NAME\",\"role\":\"worker\"}" >"$PROFILE_PROPERTY_PATH/node-profile-$PROFILE_NAME.json"
  chown dokku:dokku "$PROFILE_PROPERTY_PATH/node-profile-$PROFILE_NAME.json"
}
