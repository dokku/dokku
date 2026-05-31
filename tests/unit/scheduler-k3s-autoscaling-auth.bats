#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku apps:create $TEST_APP >/dev/null 2>/dev/null || true
  dokku apps:create ${TEST_APP}-2 >/dev/null 2>/dev/null || true
}

teardown() {
  dokku scheduler-k3s:autoscaling-auth:set $TEST_APP datadog >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:autoscaling-auth:set ${TEST_APP}-2 datadog >/dev/null 2>/dev/null || true
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
  assert_output_contains "$TEST_APP scheduler-k3s information"
  assert_output_contains "${TEST_APP}-2 scheduler-k3s information"
  assert_output_contains "Datadog:" 2
}
