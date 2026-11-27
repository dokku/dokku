#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_bats_envgroups
  dokku config:unset --global --no-restart ENVGROUP_SHARED >/dev/null 2>&1 || true
  destroy_app
  rm -f /tmp/dokku-backup-*.tar.gz 2>/dev/null || true
  global_teardown
}

@test "(envgroup) group config is available at runtime (herokuish)" {
  dokku builder-herokuish:set "$TEST_APP" allowed true
  dokku config:set --global --no-restart ENVGROUP_SHARED=global
  dokku envgroup:create bats-dev
  dokku envgroup:set bats-dev ENVGROUP_SHARED=group GROUP_ONLY=first
  dokku envgroup:add bats-dev "$TEST_APP"

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP env | grep -E '^(ENVGROUP_SHARED|GROUP_ONLY)=' | sort | tr '\n' ' '"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ENVGROUP_SHARED=group GROUP_ONLY=first "

  run /bin/bash -c "dokku envgroup:set bats-dev GROUP_ONLY=second"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "The following apps must be redeployed for the change to take effect: $TEST_APP"
  assert_output_contains "Restarting app $TEST_APP" 0

  run /bin/bash -c "dokku envgroup:set --restart bats-dev GROUP_ONLY=third"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Restarting app $TEST_APP"
  assert_output_contains "The following apps must be redeployed" 0

  run /bin/bash -c "dokku run $TEST_APP env | grep -E '^GROUP_ONLY='"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "GROUP_ONLY=third"

  run /bin/bash -c "dokku envgroup:remove --restart bats-dev $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Restarting app $TEST_APP"

  run /bin/bash -c "dokku run $TEST_APP env | grep -E '^ENVGROUP_SHARED='"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ENVGROUP_SHARED=global"

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP env | grep -E '^(ENVGROUP_SHARED|GROUP_ONLY)=' | tr '\n' ' '"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ENVGROUP_SHARED=global "
}

@test "(envgroup) group config is available at runtime (dockerfile)" {
  dokku envgroup:create bats-dev
  dokku envgroup:set bats-dev GROUP_ONLY=first
  dokku envgroup:add bats-dev "$TEST_APP"

  run deploy_app dockerfile dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP env | grep -E '^GROUP_ONLY='"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "GROUP_ONLY=first"

  run /bin/bash -c "dokku envgroup:destroy --force bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "The following apps must be redeployed for the change to take effect: $TEST_APP"
}

@test "(backup) envgroup round-trip restores groups and memberships" {
  dokku envgroup:create bats-first
  dokku envgroup:create bats-second
  dokku envgroup:create bats-empty
  dokku envgroup:set bats-first KEY=first SPACED="hello world"
  dokku envgroup:set bats-second KEY=second
  dokku envgroup:add bats-second "$TEST_APP"
  dokku envgroup:add bats-first "$TEST_APP"

  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  local backup_file="$output"

  dokku envgroup:destroy --force bats-first
  dokku envgroup:destroy --force bats-empty
  dokku envgroup:set bats-second KEY=changed STALE=value

  run /bin/bash -c "dokku backup:import --force '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet envgroup:list | grep '^bats-' | tr '\n' ' '"
  echo "output: $output"
  echo "status: $status"
  assert_output "bats-empty bats-first bats-second "

  run /bin/bash -c "dokku envgroup:get bats-first SPACED"
  echo "output: $output"
  echo "status: $status"
  assert_output "hello world"

  run /bin/bash -c "dokku envgroup:keys bats-second | tr '\n' ' '"
  echo "output: $output"
  echo "status: $status"
  assert_output "KEY "

  run /bin/bash -c "dokku envgroup:get bats-second KEY"
  echo "output: $output"
  echo "status: $status"
  assert_output "second"

  run /bin/bash -c "dokku envgroup:report $TEST_APP --envgroup-groups"
  echo "output: $output"
  echo "status: $status"
  assert_output "bats-second bats-first"

  run /bin/bash -c "dokku config:export --merged --format json $TEST_APP | jq -r .KEY"
  echo "output: $output"
  echo "status: $status"
  assert_output "first"
}

destroy_bats_envgroups() {
  declare desc="destroys every env group created by these tests"
  local group
  for group in $(dokku --quiet envgroup:list 2>/dev/null | grep '^bats-'); do
    dokku envgroup:destroy --force "$group" >/dev/null 2>&1 || true
  done
}
