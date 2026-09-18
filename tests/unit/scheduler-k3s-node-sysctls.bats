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

@test "(scheduler-k3s:node-sysctls:set) --replace writes the declared map in one call" {
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --global vm.max_map_count 262144"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --global vm.swappiness 10"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --replace --global vm.swappiness=20 vm.overcommit_memory=1"
  assert_success
  assert_output_contains "Removing vm.max_map_count stops dokku managing it"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --format json | jq -r '.\"--global\" | keys | sort | join(\",\")'"
  assert_success
  assert_output "vm.overcommit_memory,vm.swappiness"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --format json | jq -r '.\"--global\".\"vm.swappiness\"'"
  assert_success
  assert_output "20"

  run /bin/bash -c "dokku scheduler-k3s:profiles:add edge-workers --role worker"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --replace --profile edge-workers vm.swappiness=60"
  assert_success

  # a profile scope inherits the global sysctls and overrides them on conflict
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --format json | jq -r '.\"edge-workers\" | to_entries | sort_by(.key) | map(\"\(.key)=\(.value)\") | join(\",\")'"
  assert_success
  assert_output "vm.overcommit_memory=1,vm.swappiness=60"
}

@test "(scheduler-k3s:node-sysctls:set) --replace rejects malformed pair lists without writing" {
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --global vm.max_map_count 262144"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --replace --global"
  assert_failure
  assert_output_contains "Must specify at least one key=value pair, use scheduler-k3s:node-sysctls:clear to remove all sysctls"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --replace --global nodelimiter"
  assert_failure
  assert_output_contains "Invalid key=value pair: nodelimiter"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --replace --global vm.swappiness=10 vm.swappiness=20"
  assert_failure
  assert_output_contains "Duplicate key specified: vm.swappiness"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --format json | jq -r '.\"--global\" | keys | join(\",\")'"
  assert_success
  assert_output "vm.max_map_count"
}

@test "(scheduler-k3s:node-sysctls:clear) removes every sysctl for a scope" {
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --replace --global vm.max_map_count=262144 vm.swappiness=10"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:clear --global --profile edge-workers"
  assert_failure
  assert_output_contains "Only one of --global and --profile may be specified"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:clear --profile missing-profile"
  assert_failure
  assert_output_contains "Node profile missing-profile not found"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:clear --global"
  assert_success
  assert_output_contains "Removing vm.max_map_count stops dokku managing it"
  assert_output_contains "Removing vm.swappiness stops dokku managing it"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --format json | jq -r '.\"--global\" | length'"
  assert_success
  assert_output "0"
}
