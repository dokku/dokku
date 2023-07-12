#!/usr/bin/env bats

load test_helper
source "$PLUGIN_CORE_AVAILABLE_PATH/config/functions"

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
}

teardown() {
  detach_delete_network
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) proxy:build-config (sslip.io style hostnames)" {
  echo "127.0.0.1.sslip.io.dokku.me" >"$DOKKU_ROOT/VHOST"
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  check_urls http://${TEST_APP}.127.0.0.1.sslip.io.dokku.me
  assert_http_success http://${TEST_APP}.127.0.0.1.sslip.io.dokku.me
}

@test "(nginx-vhosts) proxy:build-config (dockerfile expose)" {
  run deploy_app dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  check_urls http://${TEST_APP}.dokku.me:3000
  check_urls http://${TEST_APP}.dokku.me:3003
  check_urls http://www.test.app.dokku.me:3000
  check_urls http://www.test.app.dokku.me:3003
  assert_http_localhost_success "http" "${TEST_APP}.dokku.me" "3000"
  assert_http_localhost_success "http" "${TEST_APP}.dokku.me" "3003"
  assert_http_localhost_success "http" "www.test.app.dokku.me" "3000"
  assert_http_localhost_success "http" "www.test.app.dokku.me" "3003"
}

@test "(nginx-vhosts) proxy:build-config (multiple networks)" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  create_attach_network
  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(nginx-vhosts) proxy:build-config (global DOKKU_PROXY_PORT)" {
  local GLOBAL_PORT=30999
  run /bin/bash -c "dokku config:set --global DOKKU_PROXY_PORT=${GLOBAL_PORT}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  check_urls http://${TEST_APP}.dokku.me:${GLOBAL_PORT}
  assert_http_success http://${TEST_APP}.dokku.me:${GLOBAL_PORT}

  run /bin/bash -c "dokku config:unset --global DOKKU_PROXY_PORT"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
