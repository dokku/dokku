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

@test "(nginx-vhosts) nginx (no server tokens)" {
  deploy_app
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
  assert_output_contains "Enabling HSTS"

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
  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
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
  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
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
  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
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

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
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
  assert_output_contains "Enabling HSTS"

  # disable hsts globally
  run /bin/bash -c "dokku nginx:set --global hsts false"
  # check it is now not applied
  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Enabling HSTS" 0

  # apply on app
  run /bin/bash -c "dokku nginx:set $TEST_APP hsts true"
  # check it is now applied
  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Enabling HSTS" 1

  # disable globally
  run /bin/bash -c "dokku nginx:set --global hsts"
  # check it is still applied
  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Enabling HSTS" 1
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

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
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

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "[::1]:80;" 0
  assert_output_contains "127.0.0.1:80;" 0
}

@test "(nginx-vhosts) nginx:set x-forwarded-for-value" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "X-Forwarded-For \$remote_addr;"

  run /bin/bash -c "dokku nginx:set $TEST_APP x-forwarded-for-value '\$http_x_forwarded_for'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "X-Forwarded-For \$http_x_forwarded_for;"
}

@test "(nginx-vhosts) nginx:set x-forwarded-port-value" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "X-Forwarded-Port \$server_port;"

  run /bin/bash -c "dokku nginx:set $TEST_APP x-forwarded-port-value '\$http_x_forwarded_port'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "X-Forwarded-Port \$http_x_forwarded_port;"
}

@test "(nginx-vhosts) nginx:set x-forwarded-proto-value" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "X-Forwarded-Proto \$scheme;"

  run /bin/bash -c "dokku nginx:set $TEST_APP x-forwarded-proto-value '\$http_x_forwarded_proto'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "X-Forwarded-Proto \$http_x_forwarded_proto;"
}

@test "(nginx-vhosts) nginx:validate-config" {
  deploy_app
  run /bin/bash -c "dokku nginx:validate-config"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:validate-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  echo "invalid config" > "/home/dokku/${TEST_APP}/nginx.conf"

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

  echo "invalid config" > "/home/dokku/${TEST_APP}/nginx.conf"

  run /bin/bash -c "dokku nginx:validate-config $TEST_APP --clean"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:validate-config"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
