#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku apps:create $TEST_APP >/dev/null 2>/dev/null || true
}

teardown() {
  dokku --force apps:destroy $TEST_APP >/dev/null 2>/dev/null || true
  dokku --force apps:destroy ${TEST_APP}-2 >/dev/null 2>/dev/null || true
  for resource in deployment service pod cronjob job secret ingress serviceaccount; do
    dokku scheduler-k3s:labels:set --global --resource-type "$resource" foo "" >/dev/null 2>/dev/null || true
    dokku scheduler-k3s:labels:set --global --resource-type "$resource" bar "" >/dev/null 2>/dev/null || true
  done
  global_teardown
}

@test "(scheduler-k3s:labels:report) lists labels after set, clears them after unset" {
  run /bin/bash -c "dokku scheduler-k3s:labels:set $TEST_APP --resource-type deployment foo bar"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:labels:report $TEST_APP --format json | jq -r '.\"global.deployment.foo\"'"
  assert_success
  assert_output "bar"

  run /bin/bash -c "dokku scheduler-k3s:labels:report $TEST_APP"
  assert_success
  assert_output_contains "$TEST_APP labels information"
  assert_output_contains "Label (global/deployment) foo:"
  assert_output_contains "bar"

  run /bin/bash -c "dokku scheduler-k3s:labels:report $TEST_APP --scheduler-k3s-labels.global.deployment.foo"
  assert_success
  assert_output "bar"

  run /bin/bash -c "dokku scheduler-k3s:labels:set $TEST_APP --resource-type deployment foo"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:labels:report $TEST_APP --format json"
  assert_success
  assert_output "{}"
}

@test "(scheduler-k3s:labels:report) honours --process-type and --resource-type filters" {
  run /bin/bash -c "dokku scheduler-k3s:labels:set $TEST_APP --resource-type deployment foo bar"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:labels:set $TEST_APP --process-type web --resource-type deployment baz qux"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:labels:set $TEST_APP --resource-type service alpha beta"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:labels:report $TEST_APP --resource-type deployment --format json | jq -r 'keys | sort | join(\",\")'"
  assert_success
  assert_output "global.deployment.foo,web.deployment.baz"

  run /bin/bash -c "dokku scheduler-k3s:labels:report $TEST_APP --process-type web --format json | jq -r 'keys | sort | join(\",\")'"
  assert_success
  assert_output "web.deployment.baz"

  run /bin/bash -c "dokku scheduler-k3s:labels:report $TEST_APP --process-type web --resource-type deployment --format json | jq -r '.\"web.deployment.baz\"'"
  assert_success
  assert_output "qux"
}

@test "(scheduler-k3s:labels:report) supports --global scope and no-arg multi-app loop" {
  run /bin/bash -c "dokku apps:create ${TEST_APP}-2"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:labels:set --global --resource-type deployment foo global-val"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:labels:set $TEST_APP --resource-type deployment foo app1-val"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:labels:set ${TEST_APP}-2 --resource-type deployment bar app2-val"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:labels:report --global --format json | jq -r '.\"global.deployment.foo\"'"
  assert_success
  assert_output "global-val"

  run /bin/bash -c "dokku scheduler-k3s:labels:report"
  assert_success
  assert_output_contains "$TEST_APP labels information"
  assert_output_contains "${TEST_APP}-2 labels information"
  assert_output_contains "app1-val"
  assert_output_contains "app2-val"
}

@test "(scheduler-k3s:labels:set) preserves multi-line values and / in keys" {
  local value=$'line one\nline two\nline three'
  run /bin/bash -c "dokku scheduler-k3s:labels:set $TEST_APP --resource-type deployment app.kubernetes.io/part-of \"$value\""
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:labels:report $TEST_APP --format json | jq -r '.\"global.deployment.app.kubernetes.io/part-of\"'"
  assert_success
  assert_output "$value"
}
