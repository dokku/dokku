#!/usr/bin/env bats

load test_helper
source "$PLUGIN_CORE_AVAILABLE_PATH/config/functions"

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

assert_access_log() {
  local prefix=$1
  run [ -a /var/log/nginx/$prefix-access.log ]
  assert_success
}

assert_error_log() {
  local prefix=$1
  run [ -a /var/log/nginx/$prefix-error.log ]
  assert_success
}

@test "(nginx-vhosts) nginx (no server tokens)" {
  deploy_app
  run /bin/bash -c "curl -s -D - $(dokku url $TEST_APP) -o /dev/null | egrep '^Server' | egrep '[0-9]+'"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(nginx-vhosts) logging" {
  deploy_app
  assert_access_log ${TEST_APP}
  assert_error_log ${TEST_APP}
}

@test "(nginx-vhosts) nginx:build-config (with SSL and unrelated domain)" {
  setup_test_tls
  add_domain "node-js-app.dokku.me"
  add_domain "test.dokku.me"
  deploy_app
  assert_ssl_domain "node-js-app.dokku.me"
  assert_http_redirect "http://test.dokku.me" "https://test.dokku.me:443/"
}

@test "(nginx-vhosts) nginx:build-config (wildcard SSL)" {
  setup_test_tls wildcard
  add_domain "wildcard1.dokku.me"
  add_domain "wildcard2.dokku.me"
  deploy_app
  cat /home/dokku/${TEST_APP}/nginx.conf
  assert_ssl_domain "wildcard1.dokku.me"
  assert_ssl_domain "wildcard2.dokku.me"
}

@test "(nginx-vhosts) nginx:build-config (wildcard SSL and unrelated domain)" {
  destroy_app
  TEST_APP="${TEST_APP}.example.com"
  setup_test_tls wildcard
  deploy_app nodejs-express dokku@dokku.me:$TEST_APP
  run /bin/bash -c "egrep '*.dokku.me' $DOKKU_ROOT/${TEST_APP}/nginx.conf | wc -l"
  assert_output "0"
}

@test "(nginx-vhosts) nginx:build-config (with SSL and Multiple SANs)" {
  setup_test_tls sans
  add_domain "test.dokku.me"
  add_domain "www.test.dokku.me"
  add_domain "www.test.app.dokku.me"
  deploy_app
  assert_ssl_domain "test.dokku.me"
  assert_ssl_domain "www.test.dokku.me"
  assert_ssl_domain "www.test.app.dokku.me"
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
  deploy_app nodejs-express dokku@dokku.me:$TEST_APP custom_nginx_template
  assert_nonssl_domain "www.test.app.dokku.me"
  assert_http_success "customtemplate.dokku.me"
}

@test "(nginx-vhosts) nginx:build-config (failed validate_nginx)" {
  run deploy_app nodejs-express dokku@dokku.me:$TEST_APP bad_custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(nginx-vhosts) nginx:validate" {
  deploy_app
  run /bin/bash -c "dokku nginx:validate"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:validate $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  echo "invalid config" > "/home/dokku/${TEST_APP}/nginx.conf"

  run /bin/bash -c "dokku nginx:validate"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku nginx:validate $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku nginx:validate --clean"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:validate"
  echo "output: $output"
  echo "status: $status"
  assert_success

  echo "invalid config" > "/home/dokku/${TEST_APP}/nginx.conf"

  run /bin/bash -c "dokku nginx:validate $TEST_APP --clean"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:validate"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
