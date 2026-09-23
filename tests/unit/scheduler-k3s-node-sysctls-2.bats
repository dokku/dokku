#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  dokku scheduler-k3s:node-sysctls:clear --global >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:node-sysctls:clear --profile edge-workers >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:profiles:remove edge-workers >/dev/null 2>/dev/null || true
  global_teardown
}

@test "(scheduler-k3s:node-sysctls:report) --stored returns only the map each scope owns" {
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --replace --global vm.overcommit_memory=1 vm.swappiness=20"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:profiles:add edge-workers --role worker"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --replace --profile edge-workers vm.swappiness=60"
  assert_success

  # the resolved set a profile's daemonset applies carries the global sysctls underneath its own
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --format json | jq -r '.\"edge-workers\" | to_entries | sort_by(.key) | map(\"\(.key)=\(.value)\") | join(\",\")'"
  assert_success
  assert_output "vm.overcommit_memory=1,vm.swappiness=60"

  # the stored map holds only what the profile itself was set to
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --stored --format json | jq -r '.\"edge-workers\" | to_entries | sort_by(.key) | map(\"\(.key)=\(.value)\") | join(\",\")'"
  assert_success
  assert_output "vm.swappiness=60"

  # the global scope stores exactly what it applies
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --stored --format json | jq -r '.\"--global\" | to_entries | sort_by(.key) | map(\"\(.key)=\(.value)\") | join(\",\")'"
  assert_success
  assert_output "vm.overcommit_memory=1,vm.swappiness=20"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --stored --global --format json | jq -rS '.'"
  assert_success
  stored_global="$output"
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --global --format json | jq -rS '.'"
  assert_success
  assert_output "$stored_global"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --stored --profile edge-workers --format json | jq -rS 'keys | join(\",\")'"
  assert_success
  assert_output "edge-workers"

  # the stdout renderer honours both flags too
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --stored --profile edge-workers"
  assert_success
  assert_output_contains "edge-workers"
  assert_output_not_contains "vm.overcommit_memory"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:clear --profile edge-workers"
  assert_success

  # a cleared profile keeps inheriting the global sysctls, so only the stored map converges
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --format json | jq -r '.\"edge-workers\" | length'"
  assert_success
  assert_output "2"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --stored --profile edge-workers --format json | jq -r '.\"edge-workers\" | length'"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --stored --profile edge-workers"
  assert_success
  assert_output_not_contains "vm.overcommit_memory"
  assert_output_not_contains "vm.swappiness"
}

@test "(scheduler-k3s:node-sysctls:report) rejects conflicting and unknown scopes" {
  run /bin/bash -c "dokku scheduler-k3s:profiles:add edge-workers --role worker"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --global --profile edge-workers"
  assert_failure
  assert_output_contains "Only one of --global and --profile may be specified"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --profile missing-profile"
  assert_failure
  assert_output_contains "Node profile missing-profile not found"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --stored --format bogus"
  assert_failure
  assert_output_contains "Invalid format: bogus"
}
