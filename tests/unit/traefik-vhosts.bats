#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
  dokku traefik:set --global challenge-mode >/dev/null 2>&1 || true
  dokku traefik:set --global dns-provider-test_key >/dev/null 2>&1 || true
}

@test "(traefik:report) info-flag works before deploy" {
  run /bin/bash -c "dokku traefik:set --global challenge-mode"
  assert_success

  run /bin/bash -c "dokku traefik:report $TEST_APP --traefik-computed-challenge-mode"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "tls"

  run /bin/bash -c "dokku traefik:report $TEST_APP --traefik-invalid-flag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid flag passed"
}

@test "(traefik:report) --format json" {
  run /bin/bash -c "dokku traefik:report --global --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:report --global --format json | jq -r '.\"computed-api-enabled\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku traefik:report --global"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "global traefik information"
}

@test "(traefik:report) dns-provider values are masked unless queried" {
  run /bin/bash -c "dokku traefik:set --global dns-provider-test_key secretvalue"
  assert_success

  run /bin/bash -c "dokku traefik:report --global | grep test_key"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "*******"

  run /bin/bash -c "dokku traefik:report --global --traefik-dns-provider-test_key"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "secretvalue"

  run /bin/bash -c "dokku traefik:report --global --format json | jq -r '.\"dns-provider-test_key\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "secretvalue"
}
