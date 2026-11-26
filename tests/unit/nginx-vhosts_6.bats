#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
}

teardown() {
  dokku nginx:set --global http2 || true
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) nginx (no server tokens)" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "curl -s -D - $(dokku url $TEST_APP) -o /dev/null | grep -E '^Server' | grep -E '[0-9]+'"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(nginx-vhosts) nginx:set hsts" {
  setup_test_tls wildcard
  local HSTS_CONF="/home/dokku/${TEST_APP}/nginx.conf.d/hsts.conf"

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Enabling HSTS" 2

  run /bin/bash -c "test -f $HSTS_CONF"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "cat $HSTS_CONF | grep includeSubdomains"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "cat $HSTS_CONF | grep 'max-age=15724800'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "cat $HSTS_CONF | grep preload"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku nginx:set $TEST_APP hsts-include-subdomains false"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "cat $HSTS_CONF | grep includeSubdomains"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku nginx:set $TEST_APP hsts-max-age 120"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "cat $HSTS_CONF | grep 'max-age=120'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set $TEST_APP hsts-preload true"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "cat $HSTS_CONF | grep preload"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set $TEST_APP hsts false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Enabling HSTS" 0

  run /bin/bash -c "test -f $DOKKU_ROOT/$TEST_APP/nginx.conf.d/hsts.conf"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(nginx-vhosts) nginx:set --global hsts" {
  setup_test_tls wildcard
  local HSTS_CONF="/home/dokku/${TEST_APP}/nginx.conf.d/hsts.conf"

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Enabling HSTS" 2

  # disable hsts globally
  run /bin/bash -c "dokku nginx:set --global hsts false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # check it is now not applied
  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Enabling HSTS" 0

  # apply on app
  run /bin/bash -c "dokku nginx:set $TEST_APP hsts true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # check it is now applied
  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Enabling HSTS" 1

  # set global value to default
  run /bin/bash -c "dokku nginx:set --global hsts"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # check it is still applied
  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Enabling HSTS" 1
}

@test "(nginx-vhosts) nginx:set http2" {
  setup_test_tls wildcard

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP | grep -c http2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set $TEST_APP http2 false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  if dpkg --compare-versions "$(nginx -v 2>&1 | cut -d'/' -f 2 | awk '{print $1}')" lt 1.25.1; then
    assert_output_contains "enables http2 per listen address"
  else
    assert_output_contains "enables http2 per listen address" 0
  fi

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "ssl;" 2
  assert_output_contains "http2" 0

  run /bin/bash -c "dokku nginx:set $TEST_APP http2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP | grep -c http2"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(nginx-vhosts) nginx:set --global http2" {
  setup_test_tls wildcard

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set --global http2 false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "http2" 0

  run /bin/bash -c "dokku nginx:set $TEST_APP http2 true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP | grep -c http2"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(nginx-vhosts) nginx:set bind-address" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set $TEST_APP bind-address-ipv4 127.0.0.1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set $TEST_APP bind-address-ipv6 ::1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "[::1]:80;"
  assert_output_contains "127.0.0.1:80;"

  run /bin/bash -c "dokku nginx:set $TEST_APP bind-address-ipv4"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set $TEST_APP bind-address-ipv6"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "[::1]:80;" 0
  assert_output_contains "127.0.0.1:80;" 0
}

@test "(nginx-vhosts) nginx:validate-config" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku nginx:validate-config"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:validate-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  echo "invalid config" >"/home/dokku/${TEST_APP}/nginx.conf"

  run /bin/bash -c "dokku nginx:validate-config"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku nginx:validate-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku nginx:validate-config --clean"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:validate-config"
  echo "output: $output"
  echo "status: $status"
  assert_success

  echo "invalid config" >"/home/dokku/${TEST_APP}/nginx.conf"

  run /bin/bash -c "dokku nginx:validate-config $TEST_APP --clean"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:validate-config"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
