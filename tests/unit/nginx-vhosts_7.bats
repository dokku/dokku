#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
}

teardown() {
  destroy_app 0 $TEST_APP
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) nginx:build-config (wildcard SSL and custom nginx template)" {
  setup_test_tls wildcard
  add_domain "wildcard1.dokku.me"
  add_domain "wildcard2.dokku.me"
  deploy_app nodejs-express dokku@dokku.me:$TEST_APP custom_ssl_nginx_template

  assert_ssl_domain "wildcard1.dokku.me"
  assert_ssl_domain "wildcard2.dokku.me"
  assert_http_redirect "http://${CUSTOM_TEMPLATE_SSL_DOMAIN}" "https://${CUSTOM_TEMPLATE_SSL_DOMAIN}:443/"
  assert_http_success "https://${CUSTOM_TEMPLATE_SSL_DOMAIN}"
}

@test "(nginx-vhosts) nginx:build-config (custom nginx template - no ssl)" {
  add_domain "www.test.app.dokku.me"
  run deploy_app nodejs-express dokku@dokku.me:$TEST_APP custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP worker=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_nonssl_domain "www.test.app.dokku.me"
  assert_http_success "customtemplate.dokku.me"

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "${TEST_APP}-worker-5000"
}

@test "(nginx-vhosts) nginx:build-config (disable custom nginx template - no ssl)" {
  run /bin/bash -c "dokku nginx:set $TEST_APP  disable-custom-config true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  add_domain "www.test.app.dokku.me"
  run deploy_app nodejs-express dokku@dokku.me:$TEST_APP custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP worker=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_nonssl_domain "www.test.app.dokku.me"

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "${TEST_APP}-worker-5000" 0
}

@test "(nginx-vhosts) nginx:build-config (failed validate_nginx)" {
  run deploy_app nodejs-express dokku@dokku.me:$TEST_APP bad_custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
