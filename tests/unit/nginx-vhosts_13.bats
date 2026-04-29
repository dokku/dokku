#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) proxy:build-config generates 502 config for undeployed app" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "return 502" -1
}

@test "(nginx-vhosts) proxy:build-config undeployed app returns 502" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  sleep 2

  assert_http_localhost_response "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/" "" "502"
}

@test "(nginx-vhosts) proxy:build-config 502 config replaced after deploy" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "return 502" 0
  assert_output_contains "proxy_pass" -1

  assert_http_success "http://${TEST_APP}.${DOKKU_DOMAIN}"
}

@test "(nginx-vhosts) proxy:build-config domains:add updates 502 config" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP custom.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "custom.${DOKKU_DOMAIN}" -1
  assert_output_contains "return 502" -1
}

@test "(nginx-vhosts) proxy:build-config no nginx.conf when proxy disabled" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
