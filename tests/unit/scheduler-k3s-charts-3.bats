#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  dokku scheduler-k3s:charts:clear cert-manager >/dev/null 2>/dev/null || true
  dokku scheduler-k3s:charts:clear traefik >/dev/null 2>/dev/null || true
  global_teardown
}

@test "(scheduler-k3s:charts:set) --replace writes the declared map in one call" {
  run /bin/bash -c "dokku scheduler-k3s:charts:set traefik.replicas 2"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:charts:set traefik.foo bar"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:charts:set cert-manager.installCRDs false"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:set --replace traefik replicas=3 service.annotations.prometheus.io/scrape=true opt=a=b empty="
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:report traefik --format json | jq -c -S ."
  assert_success
  assert_output '{"traefik.empty":"","traefik.opt":"a=b","traefik.replicas":"3","traefik.service.annotations.prometheus.io/scrape":"true"}'

  run /bin/bash -c "dokku scheduler-k3s:charts:report cert-manager --format json | jq -c -S ."
  assert_success
  assert_output '{"cert-manager.installCRDs":"false"}'
}

@test "(scheduler-k3s:charts:set) --replace rejects invalid input without writing" {
  run /bin/bash -c "dokku scheduler-k3s:charts:set traefik.replicas 2"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:set --replace"
  assert_failure
  assert_output_contains "Missing chart name"

  run /bin/bash -c "dokku scheduler-k3s:charts:set --replace not-a-chart replicas=3"
  assert_failure
  assert_output_contains "Invalid chart name"

  run /bin/bash -c "dokku scheduler-k3s:charts:set --replace replicas=3"
  assert_failure
  assert_output_contains "Invalid chart name"

  run /bin/bash -c "dokku scheduler-k3s:charts:set --replace traefik"
  assert_failure
  assert_output_contains "scheduler-k3s:charts:clear"

  run /bin/bash -c "dokku scheduler-k3s:charts:set --replace traefik replicas=3 malformed"
  assert_failure
  assert_output_contains "Invalid key=value pair: malformed"

  run /bin/bash -c "dokku scheduler-k3s:charts:set --replace traefik replicas=3 =3"
  assert_failure
  assert_output_contains "Invalid key=value pair: =3"

  run /bin/bash -c "dokku scheduler-k3s:charts:set --replace traefik replicas=3 replicas=4"
  assert_failure
  assert_output_contains "Duplicate key specified: replicas"

  run /bin/bash -c "dokku scheduler-k3s:charts:report traefik --format json | jq -c -S ."
  assert_success
  assert_output '{"traefik.replicas":"2"}'
}

@test "(scheduler-k3s:charts:clear) removes every override for a chart" {
  run /bin/bash -c "dokku scheduler-k3s:charts:set --replace traefik replicas=2 foo=bar"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:charts:set cert-manager.installCRDs false"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:clear traefik"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:report traefik --format json"
  assert_success
  assert_output "{}"

  run /bin/bash -c "dokku scheduler-k3s:charts:report cert-manager --format json | jq -c -S ."
  assert_success
  assert_output '{"cert-manager.installCRDs":"false"}'

  run /bin/bash -c "dokku scheduler-k3s:charts:clear traefik"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:charts:clear"
  assert_failure
  assert_output_contains "Missing chart name"

  run /bin/bash -c "dokku scheduler-k3s:charts:clear not-a-chart"
  assert_failure
  assert_output_contains "Invalid chart name"
}
