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

@test "(certs:report) info-flag works before deploy" {
  run /bin/bash -c "dokku certs:report $TEST_APP --ssl-hostnames"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku certs:report $TEST_APP --ssl-dir"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "/$TEST_APP/tls"

  run /bin/bash -c "dokku certs:report $TEST_APP --ssl-invalid-flag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid flag passed"
}

@test "(certs:report) --global --format json" {
  run /bin/bash -c "dokku certs:report --global --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "{}"

  run /bin/bash -c "dokku certs:report --global"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "global ssl information"

  run /bin/bash -c "dokku certs:report --global --ssl-bogus"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid flag passed"
}

@test "(certs:report) reports hostnames and subject for a CN-only certificate" {
  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/server.crt $BATS_TMPDIR/tls/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku certs:report $TEST_APP --ssl-hostnames"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku.me"

  run /bin/bash -c "dokku certs:report $TEST_APP --ssl-subject"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "CN=dokku.me"
}

@test "(certs:report) preserves multi-field subject order" {
  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/domain.com.crt $BATS_TMPDIR/tls/domain.com.key"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku certs:report $TEST_APP --ssl-subject"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "C=US; ST=California; L=San Francisco; O=Expa; OU=Operations; CN=node-js-app.dokku.me"

  run /bin/bash -c "dokku certs:report $TEST_APP --ssl-hostnames"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "node-js-app.dokku.me"
}

@test "(certs:report) reports the CN and all SANs as hostnames" {
  local SANS_TLS="$BATS_TMPDIR/tls-sans"
  mkdir -p "$SANS_TLS"
  tar xf "$BATS_TEST_DIRNAME/server_ssl_sans.tar" -C "$SANS_TLS"
  sudo chown -R dokku:dokku "$SANS_TLS"

  run /bin/bash -c "dokku certs:add $TEST_APP $SANS_TLS/server.crt $SANS_TLS/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku certs:report $TEST_APP --ssl-hostnames"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test.dokku.me www.test.app.dokku.me www.test.dokku.me"

  rm -rf "$SANS_TLS"
}

@test "(certs) get_ssl_hostnames parses a CN-only certificate" {
  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/server.crt $BATS_TMPDIR/tls/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success

  source "$PLUGIN_CORE_AVAILABLE_PATH/certs/functions"
  run get_ssl_hostnames "$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku.me"
}

@test "(certs) get_ssl_hostnames includes the CN and SANs" {
  local SANS_TLS="$BATS_TMPDIR/tls-sans"
  mkdir -p "$SANS_TLS"
  tar xf "$BATS_TEST_DIRNAME/server_ssl_sans.tar" -C "$SANS_TLS"
  sudo chown -R dokku:dokku "$SANS_TLS"

  run /bin/bash -c "dokku certs:add $TEST_APP $SANS_TLS/server.crt $SANS_TLS/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success

  source "$PLUGIN_CORE_AVAILABLE_PATH/certs/functions"
  run get_ssl_hostnames "$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line "test.dokku.me"
  assert_line "www.test.dokku.me"
  assert_line "www.test.app.dokku.me"
  assert_line_count 3

  rm -rf "$SANS_TLS"
}

@test "(certs) certs:add" {
  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/server.crt $BATS_TMPDIR/tls/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(certs:add) preserves explicit https:443 port mapping" {
  run /bin/bash -c "dokku ports:set $TEST_APP http:80:80 https:443:443"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/server.crt $BATS_TMPDIR/tls/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:80 https:443:443"
}

@test "(certs:add) adds default https:443 mapping when none exists" {
  run /bin/bash -c "dokku ports:set $TEST_APP http:80:5000"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/server.crt $BATS_TMPDIR/tls/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:5000 https:443:5000"
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

@test "(certs) certs:add with symbolic link for certificate" {
  ln -s $BATS_TMPDIR/tls/server.crt $BATS_TMPDIR/tls/linked_server.crt
  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/linked_server.crt $BATS_TMPDIR/tls/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(certs) certs:add with symbolic link for private key" {
  ln -s $BATS_TMPDIR/tls/server.key $BATS_TMPDIR/tls/linked_server.key
  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/server.crt $BATS_TMPDIR/tls/linked_server.key"
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

@test "(certs) certs-set plugin trigger" {
  run_plugn_trigger certs-set "$TEST_APP" "$BATS_TMPDIR/tls/server.crt" "$BATS_TMPDIR/tls/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success
  [[ -f "$DOKKU_ROOT/$TEST_APP/tls/server.crt" ]]
  [[ -f "$DOKKU_ROOT/$TEST_APP/tls/server.key" ]]
}

@test "(certs) certs-set plugin trigger missing key file" {
  run_plugn_trigger certs-set "$TEST_APP" "$BATS_TMPDIR/tls/server.crt" "/nonexistent/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "KEY file specified not found"
}

@test "(certs) certs-remove plugin trigger" {
  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/server.crt $BATS_TMPDIR/tls/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run_plugn_trigger certs-remove "$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  [[ ! -d "$DOKKU_ROOT/$TEST_APP/tls" ]]
}

@test "(certs) certs-remove plugin trigger without endpoint" {
  run_plugn_trigger certs-remove "$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "An app-specific SSL endpoint is not defined"
}

@test "(certs) certs:show" {
  run /bin/bash -c "dokku certs:show fake-app-name 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "App fake-app-name does not exist"
  assert_failure

  run /bin/bash -c "dokku certs:show $TEST_APP fake-var-name"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "specify either 'key' or 'crt'"
  assert_failure

  run /bin/bash -c "dokku certs:add $TEST_APP < $BATS_TEST_DIRNAME/server_ssl.tar && dokku certs:show $TEST_APP crt"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "-----END CERTIFICATE-----"
  assert_success

  run /bin/bash -c "dokku certs:add $TEST_APP < $BATS_TEST_DIRNAME/server_ssl.tar && dokku certs:show $TEST_APP key"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "-----END PRIVATE KEY-----"
  assert_success
}
