#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  rm -f "$DOKKU_ROOT/.dokkurc/DOKKU_TRACE"
}

teardown() {
  rm -f "$DOKKU_ROOT/.dokkurc/DOKKU_TRACE"
  global_teardown
}

@test "(trace) trace:on" {
  run /bin/bash -c "dokku trace:on"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/.dokkurc/DOKKU_TRACE"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(trace) trace on" {
  run /bin/bash -c "dokku trace on"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/.dokkurc/DOKKU_TRACE"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(trace) trace:off, trace off" {
  touch "$DOKKU_ROOT/.dokkurc/DOKKU_TRACE"
  run /bin/bash -c "dokku trace:off"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/.dokkurc/DOKKU_TRACE"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  touch "$DOKKU_ROOT/.dokkurc/DOKKU_TRACE"
  run /bin/bash -c "dokku trace off"
  echo "output: $output"
  echo "status: $status"
  assert_success


  run /bin/bash -c "test -f $DOKKU_ROOT/.dokkurc/DOKKU_TRACE"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
