#!/usr/bin/env bats

load test_helper

setup() {
  dokku traefik:set --global basic-auth-password
  dokku traefik:set --global dns-provider-test_key
  create_app
}

teardown() {
  destroy_app
  dokku traefik:set --global basic-auth-password >/dev/null 2>&1 || true
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

  run /bin/bash -c "dokku traefik:report --global --traefik-global-dns-provider-test_key"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "secretvalue"

  run /bin/bash -c "dokku traefik:report --global --format json | jq -r '.\"global-dns-provider-test_key\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "secretvalue"
}

@test "(traefik:report) basic-auth-password is masked unless queried" {
  run /bin/bash -c "dokku traefik:report --global --format json | jq -r '.\"global-basic-auth-password\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku traefik:report --global | grep 'basic auth password'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "*******"

  run /bin/bash -c "dokku traefik:set --global basic-auth-password secretpassword"
  assert_success

  run /bin/bash -c "dokku traefik:report --global | grep 'basic auth password'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "*******" 2
  assert_output_not_contains "secretpassword"

  run /bin/bash -c "dokku traefik:report --global --traefik-global-basic-auth-password"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "secretpassword"

  run /bin/bash -c "dokku traefik:report --global --traefik-computed-basic-auth-password"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "secretpassword"

  run /bin/bash -c "dokku traefik:report --global --format json | jq -r '.\"global-basic-auth-password\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "secretpassword"
}

@test "(traefik:report) credentials are masked in the aggregate report" {
  run /bin/bash -c "dokku traefik:set --global dns-provider-test_key secretvalue"
  assert_success

  run /bin/bash -c "dokku traefik:set --global basic-auth-password secretpassword"
  assert_success

  run /bin/bash -c "dokku report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "global dns provider test_key"
  assert_output_contains "*******" -1
  assert_output_not_contains "secretvalue"
  assert_output_not_contains "secretpassword"
}
