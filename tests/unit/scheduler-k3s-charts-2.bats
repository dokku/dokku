#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  dokku scheduler-k3s:charts:set traefik.service.annotations.prometheus.io/scrape >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:charts:set traefik.controller.config >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:charts:set cert-manager.installCRDs >/dev/null 2>/dev/null || true
  global_teardown
}

@test "(scheduler-k3s:charts:set) accepts property names containing /" {
  run /bin/bash -c "dokku scheduler-k3s:charts:set traefik.service.annotations.prometheus.io/scrape true"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:report traefik --format json | jq -r '.\"traefik.service.annotations.prometheus.io/scrape\"'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku scheduler-k3s:charts:set traefik.service.annotations.prometheus.io/scrape"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:report traefik --format json"
  assert_success
  assert_output "{}"
}

@test "(scheduler-k3s:charts:set) preserves multi-line values" {
  local value=$'line one\nline two\nline three'
  run /bin/bash -c "dokku scheduler-k3s:charts:set traefik.controller.config \"$value\""
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:report traefik --format json | jq -r '.\"traefik.controller.config\"'"
  assert_success
  assert_output "$value"
}

@test "(scheduler-k3s:set) deprecated chart.* form writes through map storage" {
  run /bin/bash -c "dokku scheduler-k3s:set --global chart.cert-manager.installCRDs false"
  assert_success
  assert_output_contains "deprecated"

  run /bin/bash -c "dokku scheduler-k3s:charts:report cert-manager --format json | jq -r '.\"cert-manager.installCRDs\"'"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku scheduler-k3s:set --global chart.cert-manager.installCRDs"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:report cert-manager --format json"
  assert_success
  assert_output "{}"
}
