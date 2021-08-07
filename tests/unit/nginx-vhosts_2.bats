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

@test "(nginx-vhosts) nginx:build-config (with SSL and unrelated domain)" {
  setup_test_tls
  add_domain "node-js-app.dokku.me"
  add_domain "test.dokku.me"
  deploy_app
  dokku nginx:show-config $TEST_APP
  assert_ssl_domain "node-js-app.dokku.me"
  assert_http_redirect "http://test.dokku.me" "https://test.dokku.me:443/"
}

@test "(nginx-vhosts) nginx:build-config (wildcard SSL)" {
  setup_test_tls wildcard
  add_domain "wildcard1.dokku.me"
  add_domain "wildcard2.dokku.me"
  deploy_app
  dokku nginx:show-config $TEST_APP
  assert_ssl_domain "wildcard1.dokku.me"
  assert_ssl_domain "wildcard2.dokku.me"
}

@test "(nginx-vhosts) nginx:build-config (wildcard SSL and unrelated domain) 1" {
  destroy_app
  TEST_APP="${TEST_APP}.example.com"
  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  setup_test_tls wildcard
  deploy_app nodejs-express dokku@dokku.me:$TEST_APP
  run /bin/bash -c "dokku nginx:show-config $TEST_APP | grep -e '*.dokku.me' | wc -l"
  dokku nginx:show-config $TEST_APP
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
