#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) pre-validate fails fast on broken nginx.conf.sigil" {
  run deploy_app nodejs-express dokku@$DOKKU_DOMAIN:$TEST_APP bad_custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Pre-validating custom nginx.conf.sigil"
  assert_output_contains "Custom nginx.conf.sigil failed nginx -t validation"
  assert_output_not_contains "Building $TEST_APP"
  assert_output_not_contains "Releasing $TEST_APP"
}

@test "(nginx-vhosts) pre-validate succeeds on first deploy with valid nginx.conf.sigil" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Pre-validating custom nginx.conf.sigil"
}

@test "(nginx-vhosts) pre-validate is skipped when disable-custom-config=true" {
  run /bin/bash -c "dokku nginx:set $TEST_APP disable-custom-config true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app nodejs-express dokku@$DOKKU_DOMAIN:$TEST_APP bad_custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "Pre-validating custom nginx.conf.sigil"
}

@test "(nginx-vhosts) pre-validate is skipped when proxy is not nginx" {
  run /bin/bash -c "dokku proxy:set $TEST_APP caddy"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app nodejs-express dokku@$DOKKU_DOMAIN:$TEST_APP bad_custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "Pre-validating custom nginx.conf.sigil"
}
