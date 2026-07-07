#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
  dokku openresty:set --global log-level >/dev/null 2>&1 || true
  dokku openresty:set --global proxy-buffer-size >/dev/null 2>&1 || true
  dokku openresty:set --global client-max-body-size >/dev/null 2>&1 || true
}

@test "(openresty:report) info-flag works before deploy" {
  run /bin/bash -c "dokku openresty:set --global log-level"
  assert_success

  run /bin/bash -c "dokku openresty:report $TEST_APP --openresty-computed-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ERROR"

  run /bin/bash -c "dokku openresty:report $TEST_APP --openresty-invalid-flag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid flag passed"
}

@test "(openresty:report) --format json" {
  run /bin/bash -c "dokku openresty:set --global proxy-buffer-size"
  assert_success

  run /bin/bash -c "dokku openresty:report --global --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"computed-proxy-buffer-size\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "4k"

  run /bin/bash -c "dokku openresty:report --global"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "global openresty information"
}

@test "(openresty:report) client-max-body-size raw global computed" {
  run /bin/bash -c "dokku openresty:set --global client-max-body-size"
  assert_success

  run /bin/bash -c "dokku openresty:report $TEST_APP --openresty-client-max-body-size"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku openresty:report $TEST_APP --openresty-computed-client-max-body-size"
  assert_success
  assert_output "1m"

  run /bin/bash -c "dokku openresty:set --global client-max-body-size 2m"
  assert_success

  run /bin/bash -c "dokku openresty:report $TEST_APP --openresty-computed-client-max-body-size"
  assert_success
  assert_output "2m"

  run /bin/bash -c "dokku openresty:set $TEST_APP client-max-body-size 5m"
  assert_success

  run /bin/bash -c "dokku openresty:report $TEST_APP --openresty-client-max-body-size"
  assert_success
  assert_output "5m"

  run /bin/bash -c "dokku openresty:report $TEST_APP --openresty-computed-client-max-body-size"
  assert_success
  assert_output "5m"
}
