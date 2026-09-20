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
  # DOKKU_TRACE is set in the environment rather than passed as --trace so that the
  # parent traces too: the sudo line naming the forwarded variables is only emitted there
  run /bin/bash -c "DOKKU_TRACE=1 DOKKU_TEST_CROSSING=1 dokku apps:list 2>&1 >/dev/null | grep -o -- '--preserve-env=[^ ]*' | head -1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--preserve-env="
  assert_output_not_contains "DOKKU_TEST_CROSSING"

  run /bin/bash -c "DOKKU_TRACE=1 DOKKU_PRESERVE_ENV=DOKKU_TEST_CROSSING DOKKU_TEST_CROSSING=1 dokku apps:list 2>&1 >/dev/null | grep -o -- '--preserve-env=[^ ]*' | head -1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "DOKKU_TEST_CROSSING"
}

@test "(core) [global-flags] the system sudo supports a named preserve-env list" {
  run /bin/bash -c "DOKKU_TEST_CROSSING=1 sudo -u dokku --preserve-env=DOKKU_TEST_CROSSING -H /usr/bin/env | grep -cxE 'DOKKU_TEST_CROSSING=1|HOME=/home/dokku'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"
}
