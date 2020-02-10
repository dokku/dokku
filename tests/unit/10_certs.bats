#!/usr/bin/env bats

load test_helper

setup_local_tls() {
  TLS=$BATS_TMPDIR/tls
  mkdir -p $TLS
  tar xf $BATS_TEST_DIRNAME/server_ssl.tar -C $TLS
  tar xf $BATS_TEST_DIRNAME/domain_ssl.tar -C $TLS
  sudo chown -R dokku:dokku $TLS
}

teardown_local_tls() {
  TLS=$BATS_TMPDIR/tls
  rm -R $TLS
}

setup() {
  global_setup
  create_app
  setup_local_tls
}

teardown() {
  destroy_app
  teardown_local_tls
  global_teardown
}

@test "(certs) certs:help" {
  run /bin/bash -c "dokku certs"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage SSL (TLS) certs"
  help_output="$output"

  run /bin/bash -c "dokku certs:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage SSL (TLS) certs"
  assert_output "$help_output"
}

@test "(certs) certs:add" {
  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/server.crt $BATS_TMPDIR/tls/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(certs) certs:add with multiple dots in the filename" {
  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/domain.com.crt $BATS_TMPDIR/tls/domain.com.key"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(certs) certs:add tar:in" {
  run /bin/bash -c "dokku certs:add $TEST_APP < $BATS_TEST_DIRNAME/server_ssl.tar"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(certs) certs:add tar:in should ignore OSX hidden files" {
  run /bin/bash -c "dokku certs:add $TEST_APP < $BATS_TEST_DIRNAME/osx_ssl_tarred.tar"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(certs) certs:info" {
  run /bin/bash -c "dokku certs:add $TEST_APP < $BATS_TEST_DIRNAME/server_ssl.tar && dokku certs:info $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(certs) certs:remove" {
  run /bin/bash -c "dokku certs:add $TEST_APP < $BATS_TEST_DIRNAME/server_ssl.tar && dokku certs:remove $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
