#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  [[ -f "$DOKKU_ROOT/HOSTNAME" ]] && cp -fp "$DOKKU_ROOT/HOSTNAME" "$DOKKU_ROOT/HOSTNAME.bak"
  create_app
}

teardown() {
  destroy_app 0 $TEST_APP
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  [[ -f "$DOKKU_ROOT/HOSTNAME.bak" ]] && mv "$DOKKU_ROOT/HOSTNAME.bak" "$DOKKU_ROOT/HOSTNAME" && chown dokku:dokku "$DOKKU_ROOT/HOSTNAME"
  global_teardown
}

assert_nonssl_domain() {
  local domain=$1
  assert_app_domain "${domain}"
  assert_http_success "http://${domain}"
}

assert_app_domain() {
  local domain=$1
  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null | grep -xF ${domain}"
  echo "output: $output"
  echo "status: $status"
  assert_output "${domain}"
}

assert_external_port() {
  local CID="$1"; local exit_status="$2"
  local EXTERNAL_PORT_COUNT=$(docker port $CID | wc -l)
  run /bin/bash -c "[[ $EXTERNAL_PORT_COUNT -gt 0 ]]"
  if [[ "$exit_status" == "success" ]]; then
    assert_success
  else
    assert_failure
  fi
}

@test "(proxy) network:set bind-all-interfaces" {
  deploy_app
  assert_nonssl_domain "${TEST_APP}.dokku.me"

  run /bin/bash -c "dokku network:set $TEST_APP bind-all-interfaces true"
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success "${TEST_APP}.dokku.me"

  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    assert_external_port $(< $CID_FILE) success
  done

  run /bin/bash -c "dokku network:set $TEST_APP bind-all-interfaces false"
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success "${TEST_APP}.dokku.me"

  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    assert_external_port $(< $CID_FILE) failure
  done
}

@test "(proxy) network host-mode" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy \"--network=host\""
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "curl --silent --write-out '%{http_code}\n' `dokku url $TEST_APP` | grep 200"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
