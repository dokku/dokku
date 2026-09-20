#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(core) [global-flags] --quiet crosses the privilege drop" {
  run /bin/bash -c "dokku apps:list"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "My Apps"

  run /bin/bash -c "dokku --quiet apps:list"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "My Apps"
  assert_output_contains "$TEST_APP"
}

@test "(core) [global-flags] --trace crosses the privilege drop" {
  run /bin/bash -c "dokku --trace apps:list 2>&1 >/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "execute_dokku_cmd"
}

@test "(core) [global-flags] --force crosses the privilege drop" {
  local force_app="${TEST_APP}-force"
  run /bin/bash -c "dokku apps:create $force_app"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy $force_app </dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "To proceed, type"
}

@test "(core) [global-flags] --app crosses the privilege drop" {
  run /bin/bash -c "dokku --app $TEST_APP config:set --no-restart TEST_PRESERVED=crossed"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP TEST_PRESERVED"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "crossed"
}

@test "(core) [global-flags] a clean invocation writes nothing to stderr" {
  run /bin/bash -c "dokku apps:list 2>&1 >/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""
}

@test "(core) [global-flags] the caller environment crosses the privilege drop" {
  run /bin/bash -c "DOKKU_QUIET_OUTPUT=1 dokku apps:list"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "My Apps"
  assert_output_contains "$TEST_APP"
}

@test "(core) [global-flags] an unlisted variable only crosses when DOKKU_PRESERVE_ENV names it" {
  # DOKKU_IMAGE is not carried across the re-exec and is re-derived by the child, so the
  # traced value shows whether the caller's value survived. The flag form of --trace is
  # used deliberately: DOKKU_TRACE=1 would trace the parent as well as the child.
  run /bin/bash -c "DOKKU_IMAGE=probe/image dokku --trace apps:list 2>&1 >/dev/null | grep -m1 'export DOKKU_IMAGE'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "probe/image"

  run /bin/bash -c "DOKKU_PRESERVE_ENV=DOKKU_IMAGE DOKKU_IMAGE=probe/image dokku --trace apps:list 2>&1 >/dev/null | grep -m1 'export DOKKU_IMAGE'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "probe/image"
}

@test "(core) [global-flags] the system sudo supports a named preserve-env list" {
  run /bin/bash -c "DOKKU_TEST_CROSSING=1 sudo -u dokku --preserve-env=DOKKU_TEST_CROSSING -H /usr/bin/env | grep -cxE 'DOKKU_TEST_CROSSING=1|HOME=/home/dokku'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"
}
