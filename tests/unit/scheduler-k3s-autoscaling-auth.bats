#!/usr/bin/env bats

load test_helper

scheduler_k3s_seed_trigger_auth() {
  local scope="$1"
  local trigger="$2"
  local key="$3"
  local value="$4"
  local dir="/var/lib/dokku/config/scheduler-k3s/$scope"
  sudo mkdir -p "$dir"
  echo -n "$value" | sudo tee "$dir/trigger-auth.${trigger}.${key}" >/dev/null
}

scheduler_k3s_clear_trigger_auth() {
  local scope="$1"
  local trigger="$2"
  sudo rm -f "/var/lib/dokku/config/scheduler-k3s/$scope/trigger-auth.${trigger}."* 2>/dev/null || true
}

setup() {
  global_setup
  dokku apps:create $TEST_APP >/dev/null 2>/dev/null || true
  dokku apps:create ${TEST_APP}-2 >/dev/null 2>/dev/null || true
}

teardown() {
  dokku scheduler-k3s:autoscaling-auth:set $TEST_APP datadog >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:autoscaling-auth:set $TEST_APP memory >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:autoscaling-auth:set ${TEST_APP}-2 datadog >/dev/null 2>/dev/null || true
  scheduler_k3s_clear_trigger_auth "--global" "datadog"
  dokku --force apps:destroy $TEST_APP >/dev/null 2>/dev/null || true
  dokku --force apps:destroy ${TEST_APP}-2 >/dev/null 2>/dev/null || true
  global_teardown
}

@test "(scheduler-k3s:autoscaling-auth:report) no-arg loops all apps" {
  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:set $TEST_APP datadog --metadata apiKey=secret-1"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:set ${TEST_APP}-2 datadog --metadata apiKey=secret-2"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report"
  assert_success
  assert_output_contains "$TEST_APP autoscaling-auth information"
  assert_output_contains "${TEST_APP}-2 autoscaling-auth information"
  assert_output_contains "Datadog:" 2
}

@test "(scheduler-k3s:autoscaling-auth:report) lists metadata in json and clears after unset" {
  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:set $TEST_APP datadog --metadata apiKey=secret-1 --metadata appKey=secret-2"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --format json | jq -r '.\"datadog.apiKey\"'"
  assert_success
  assert_output "secret-1"

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --format json | jq -r '.\"datadog.appKey\"'"
  assert_success
  assert_output "secret-2"

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --scheduler-k3s-autoscaling-auth.datadog.apiKey"
  assert_success
  assert_output "*******"

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP"
  assert_success
  assert_output_contains "Datadog:"
  assert_output_not_contains "Datadog apiKey:"
  assert_output_not_contains "secret-1"
  assert_output_not_contains "secret-2"

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --include-metadata"
  assert_success
  assert_output_contains "Datadog apiKey:"
  assert_output_contains "secret-1"
  assert_output_contains "secret-2"

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:set $TEST_APP datadog"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --format json"
  assert_success
  assert_output "{}"
}

@test "(scheduler-k3s:autoscaling-auth:report) stdout includes metadata values with --include-metadata" {
  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:set $TEST_APP datadog --metadata apiKey=super-secret-value"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP"
  assert_success
  assert_output_contains "Datadog:"
  assert_output_contains "configured"
  assert_output_not_contains "Datadog apiKey:"
  assert_output_not_contains "super-secret-value"

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --include-metadata"
  assert_success
  assert_output_contains "Datadog apiKey:"
  assert_output_contains "super-secret-value"
}

@test "(scheduler-k3s:autoscaling-auth:report) reports multiple triggers in json" {
  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:set $TEST_APP datadog --metadata apiKey=secret-1"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:set $TEST_APP memory --metadata some-key=some-value"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --format json | jq -r 'keys | sort | join(\",\")'"
  assert_success
  assert_output "datadog.apiKey,memory.some-key"
}

@test "(scheduler-k3s:autoscaling-auth:report) supports --global scope" {
  scheduler_k3s_seed_trigger_auth "--global" "datadog" "apiKey" "global-secret"

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report --global --format json | jq -r '.\"datadog.apiKey\"'"
  assert_success
  assert_output "global-secret"

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report --global --include-metadata"
  assert_success
  assert_output_contains "--global autoscaling-auth information"
  assert_output_contains "Datadog:"
  assert_output_contains "Datadog apiKey:"
  assert_output_contains "global-secret"
}

@test "(scheduler-k3s:autoscaling-auth:report) returns empty json for unconfigured app" {
  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --format json"
  assert_success
  assert_output "{}"
}

@test "(scheduler-k3s:autoscaling-auth:report) rejects invalid format and info-flag combinations" {
  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --format yaml"
  assert_failure
  assert_output_contains "Invalid format"

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:set $TEST_APP datadog --metadata apiKey=secret-1"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --format json --scheduler-k3s-autoscaling-auth.datadog.apiKey"
  assert_failure
  assert_output_contains "--format flag cannot be specified when specifying an info flag"

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --scheduler-k3s-autoscaling-auth.datadog.missing"
  assert_failure
  assert_output_contains "Invalid flag passed"
}
