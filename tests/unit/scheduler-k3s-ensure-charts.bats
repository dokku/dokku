#!/usr/bin/env bats

load test_helper

setup() {
  uninstall_k3s || true
  global_setup
  dokku nginx:stop
  export KUBECONFIG="/etc/rancher/k3s/k3s.yaml"
}

teardown() {
  global_teardown
  dokku nginx:start
  uninstall_k3s || true
}

@test "(scheduler-k3s:ensure-charts) upgrades an existing release to the pinned chart version" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  install_k3s

  run /bin/bash -c "helm list -n traefik -o json | jq -r '.[] | select(.name == \"traefik\") | .chart'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "traefik-26.0.0"

  run /bin/bash -c "helm list -n traefik -o json | jq -r '.[] | select(.name == \"traefik\") | .revision'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  local initial_revision="$output"

  run /bin/bash -c "dokku scheduler-k3s:ensure-charts --charts traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "helm list -n traefik -o json | jq -r '.[] | select(.name == \"traefik\") | .chart'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "traefik-26.0.0"

  run /bin/bash -c "helm list -n traefik -o json | jq -r '.[] | select(.name == \"traefik\") | .revision'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  local upgraded_revision="$output"
  [[ "$upgraded_revision" -gt "$initial_revision" ]]

  run /bin/bash -c "dokku scheduler-k3s:ensure-charts"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "helm list -n traefik -o json | jq -r '.[] | select(.name == \"traefik\") | .revision'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$upgraded_revision"
}
