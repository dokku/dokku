#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku apps:create $TEST_APP >/dev/null 2>/dev/null || true
}

teardown() {
  dokku --force apps:destroy $TEST_APP >/dev/null 2>/dev/null || true
  dokku --force apps:destroy ${TEST_APP}-2 >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:annotations:clear --global >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:node-sysctls:clear --global >/dev/null 2>/dev/null || true
  global_teardown
}

@test "(scheduler-k3s:annotations:report) lists annotations after set, clears them after unset" {
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deployment foo bar"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json | jq -r '.\"global.deployment.foo\"'"
  assert_success
  assert_output "bar"

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP"
  assert_success
  assert_output_contains "$TEST_APP annotations information"
  assert_output_contains "Annotation (global/deployment) foo:"
  assert_output_contains "bar"

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --scheduler-k3s-annotations.global.deployment.foo"
  assert_success
  assert_output "bar"

  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deployment foo"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json"
  assert_success
  assert_output "{}"
}

@test "(scheduler-k3s:annotations:report) honours --process-type and --resource-type filters" {
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deployment foo bar"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --process-type web --resource-type deployment baz qux"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type service alpha beta"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --resource-type deployment --format json | jq -r 'keys | sort | join(\",\")'"
  assert_success
  assert_output "global.deployment.foo,web.deployment.baz"

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --process-type web --format json | jq -r 'keys | sort | join(\",\")'"
  assert_success
  assert_output "web.deployment.baz"

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --process-type web --resource-type deployment --format json | jq -r '.\"web.deployment.baz\"'"
  assert_success
  assert_output "qux"
}

@test "(scheduler-k3s:annotations:report) supports --global scope and no-arg multi-app loop" {
  run /bin/bash -c "dokku apps:create ${TEST_APP}-2"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:set --global --resource-type deployment foo global-val"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deployment foo app1-val"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:annotations:set ${TEST_APP}-2 --resource-type deployment bar app2-val"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report --global --format json | jq -r '.\"global.deployment.foo\"'"
  assert_success
  assert_output "global-val"

  run /bin/bash -c "dokku scheduler-k3s:annotations:report"
  assert_success
  assert_output_contains "$TEST_APP annotations information"
  assert_output_contains "${TEST_APP}-2 annotations information"
  assert_output_contains "app1-val"
  assert_output_contains "app2-val"
}

@test "(scheduler-k3s:annotations:set) preserves multi-line values and / in keys" {
  local value=$'line one\nline two\nline three'
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deployment prometheus.io/scrape \"$value\""
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json | jq -r '.\"global.deployment.prometheus.io/scrape\"'"
  assert_success
  assert_output "$value"
}

@test "(scheduler-k3s:annotations:report) rejects invalid format and info-flag combinations" {
  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format yaml"
  assert_failure
  assert_output_contains "Invalid format"

  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deployment foo bar"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json --scheduler-k3s-annotations.global.deployment.foo"
  assert_failure
  assert_output_contains "--format flag cannot be specified when specifying an info flag"

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --scheduler-k3s-annotations.global.deployment.missing"
  assert_failure
  assert_output_contains "Invalid flag passed"
}

@test "(scheduler-k3s:annotations:set) --replace writes the declared map in one call" {
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deployment keep old-value"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deployment drop dropped-value"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:set --replace $TEST_APP --resource-type deployment keep=new-value add=added-value"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json | jq -r 'keys | sort | join(\",\")'"
  assert_success
  assert_output "global.deployment.add,global.deployment.keep"

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json | jq -r '.\"global.deployment.keep\"'"
  assert_success
  assert_output "new-value"

  run /bin/bash -c "dokku scheduler-k3s:annotations:set --replace $TEST_APP --resource-type deployment --process-type web scoped=scoped-value"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json | jq -r 'keys | sort | join(\",\")'"
  assert_success
  assert_output "global.deployment.add,global.deployment.keep,web.deployment.scoped"

  run /bin/bash -c "dokku scheduler-k3s:annotations:set --replace --global --resource-type deployment global-key=global-value"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report --global --format json | jq -r '.\"global.deployment.global-key\"'"
  assert_success
  assert_output "global-value"
}

@test "(scheduler-k3s:annotations:set) --replace stores empty values and rejects malformed pair lists" {
  run /bin/bash -c "dokku scheduler-k3s:annotations:set --replace $TEST_APP --resource-type deployment empty="
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json | jq -r '.\"global.deployment.empty\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku scheduler-k3s:annotations:set --replace $TEST_APP --resource-type deployment"
  assert_failure
  assert_output_contains "Must specify at least one key=value pair, use scheduler-k3s:annotations:clear to remove all annotations"

  run /bin/bash -c "dokku scheduler-k3s:annotations:set --replace $TEST_APP --resource-type deployment nodelimiter"
  assert_failure
  assert_output_contains "Invalid key=value pair: nodelimiter"

  run /bin/bash -c "dokku scheduler-k3s:annotations:set --replace $TEST_APP --resource-type deployment foo=bar foo=qux"
  assert_failure
  assert_output_contains "Duplicate key specified: foo"

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json | jq -r 'keys | sort | join(\",\")'"
  assert_success
  assert_output "global.deployment.empty"
}

@test "(scheduler-k3s:annotations:set) rejects an unknown resource-type and a missing app" {
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deploymnet foo bar"
  assert_failure
  assert_output_contains "Invalid resource-type specified, valid resource types include:"

  run /bin/bash -c "dokku scheduler-k3s:annotations:set --replace $TEST_APP --resource-type deploymnet foo=bar"
  assert_failure
  assert_output_contains "Invalid resource-type specified, valid resource types include:"

  run /bin/bash -c "dokku scheduler-k3s:annotations:set ${TEST_APP}-missing --resource-type deployment foo bar"
  assert_failure
  assert_output_contains "App ${TEST_APP}-missing does not exist"

  run /bin/bash -c "dokku scheduler-k3s:annotations:set --replace ${TEST_APP}-missing --resource-type deployment foo=bar"
  assert_failure
  assert_output_contains "App ${TEST_APP}-missing does not exist"
}

@test "(scheduler-k3s:annotations:clear) removes annotations, honouring filters" {
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deployment foo bar"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --process-type web --resource-type deployment baz qux"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type service alpha beta"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:clear $TEST_APP --process-type web"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json | jq -r 'keys | sort | join(\",\")'"
  assert_success
  assert_output "global.deployment.foo,global.service.alpha"

  run /bin/bash -c "dokku scheduler-k3s:annotations:clear $TEST_APP --resource-type service"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json | jq -r 'keys | sort | join(\",\")'"
  assert_success
  assert_output "global.deployment.foo"

  run /bin/bash -c "dokku scheduler-k3s:annotations:clear $TEST_APP"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json"
  assert_success
  assert_output "{}"
}

@test "(scheduler-k3s:annotations:clear) leaves neighbouring scheduler-k3s properties alone" {
  run /bin/bash -c "dokku scheduler-k3s:set $TEST_APP deploy-timeout 300s"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:set --global vm.max_map_count 262144"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:set $TEST_APP datadog --metadata apiKey=secret-1"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deployment foo bar"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:clear $TEST_APP"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report $TEST_APP --format json"
  assert_success
  assert_output "{}"

  run /bin/bash -c "dokku --quiet scheduler-k3s:report $TEST_APP --scheduler-k3s-deploy-timeout"
  assert_success
  assert_output "300s"

  run /bin/bash -c "dokku scheduler-k3s:node-sysctls:report --format json | jq -r '.\"--global\".\"vm.max_map_count\"'"
  assert_success
  assert_output "262144"

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:report $TEST_APP --format json | jq -r '.\"datadog.apiKey\"'"
  assert_success
  assert_output "secret-1"
}

@test "(scheduler-k3s:annotations:clear) supports --global and requires an app" {
  run /bin/bash -c "dokku scheduler-k3s:annotations:set --global --resource-type deployment foo global-val"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:clear"
  assert_failure
  assert_output_contains "Please specify an app to run the command on"

  run /bin/bash -c "dokku scheduler-k3s:annotations:clear ${TEST_APP}-missing"
  assert_failure
  assert_output_contains "App ${TEST_APP}-missing does not exist"

  run /bin/bash -c "dokku scheduler-k3s:annotations:clear $TEST_APP --resource-type deploymnet"
  assert_failure
  assert_output_contains "Invalid resource-type specified, valid resource types include:"

  run /bin/bash -c "dokku scheduler-k3s:annotations:report --global --format json | jq -r '.\"global.deployment.foo\"'"
  assert_success
  assert_output "global-val"

  run /bin/bash -c "dokku scheduler-k3s:annotations:clear --global"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:report --global --format json"
  assert_success
  assert_output "{}"
}
