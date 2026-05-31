#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  dokku scheduler-k3s:charts:set cert-manager.installCRDs >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:charts:set cert-manager.version >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:charts:set traefik.replicas >/dev/null 2>/dev/null || true
  global_teardown
}

@test "(scheduler-k3s:charts:set) sets and unsets a chart property" {
  run /bin/bash -c "dokku scheduler-k3s:charts:set cert-manager.installCRDs false"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:report cert-manager --format json | jq -r '.\"cert-manager.installCRDs\"'"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku scheduler-k3s:charts:report cert-manager"
  assert_success
  assert_output_contains "cert-manager chart information"
  assert_output_contains "Chart property installCRDs:"
  assert_output_contains "false"

  run /bin/bash -c "dokku scheduler-k3s:charts:set cert-manager.installCRDs"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:report cert-manager --format json"
  assert_success
  assert_output "{}"

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-global-chart.cert-manager.installCRDs\" // \"missing\"'"
  assert_success
  assert_output "missing"
}

@test "(scheduler-k3s:charts:set) rejects invalid chart names and malformed property args" {
  run /bin/bash -c "dokku scheduler-k3s:charts:set"
  assert_failure
  assert_output_contains "Invalid property"

  run /bin/bash -c "dokku scheduler-k3s:charts:set cert-manager"
  assert_failure
  assert_output_contains "Invalid property"

  run /bin/bash -c "dokku scheduler-k3s:charts:set .version 1.0"
  assert_failure
  assert_output_contains "Invalid property"

  run /bin/bash -c "dokku scheduler-k3s:charts:set cert-manager. 1.0"
  assert_failure
  assert_output_contains "Invalid property"

  run /bin/bash -c "dokku scheduler-k3s:charts:set not-a-chart.version 1.0"
  assert_failure
  assert_output_contains "Invalid chart name"

  run /bin/bash -c "dokku scheduler-k3s:set --global chart.cert-manager.version 1.13.3"
  assert_success
  assert_output_contains "deprecated"

  run /bin/bash -c "dokku scheduler-k3s:set --global chart.cert-manager.version"
  assert_success
}

@test "(scheduler-k3s:charts:report) per-chart, single-flag query, and flat JSON" {
  run /bin/bash -c "dokku scheduler-k3s:charts:set cert-manager.version 1.13.3"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:charts:set traefik.replicas 2"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:report"
  assert_success
  assert_output_contains "cert-manager chart information"
  assert_output_contains "traefik chart information"
  assert_output_contains "longhorn chart information"
  assert_output_contains "ingress-nginx chart information"
  assert_output_contains "keda chart information"
  assert_output_contains "keda-add-ons-http chart information"
  assert_output_contains "vector chart information"

  run /bin/bash -c "dokku scheduler-k3s:charts:report cert-manager"
  assert_success
  assert_output_contains "cert-manager chart information"
  assert_output_contains "Chart property version:"
  assert_output_contains "1.13.3"

  run /bin/bash -c "dokku scheduler-k3s:charts:report cert-manager"
  assert_output_not_contains "traefik chart information"

  run /bin/bash -c "dokku scheduler-k3s:charts:report --scheduler-k3s-charts-cert-manager.version"
  assert_success
  assert_output "1.13.3"

  run /bin/bash -c "dokku scheduler-k3s:charts:report --format json | jq -r '.\"cert-manager.version\"'"
  assert_success
  assert_output "1.13.3"

  run /bin/bash -c "dokku scheduler-k3s:charts:report --format json | jq -r '.\"traefik.replicas\"'"
  assert_success
  assert_output "2"
}
