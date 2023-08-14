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

@test "(nginx-vhosts) proxy:build-config (with SSL and unrelated domain)" {
  setup_test_tls
  run /bin/bash -c "dokku domains:add $TEST_APP node-js-app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP test.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_ssl_domain "node-js-app.${DOKKU_DOMAIN}"
  assert_http_redirect "http://test.${DOKKU_DOMAIN}" "https://test.${DOKKU_DOMAIN}:443/"
}

@test "(nginx-vhosts) proxy:build-config (wildcard SSL)" {
  setup_test_tls wildcard
  run /bin/bash -c "dokku domains:add $TEST_APP wildcard1.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP wildcard2.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_ssl_domain "wildcard1.${DOKKU_DOMAIN}"
  assert_ssl_domain "wildcard2.${DOKKU_DOMAIN}"
}

@test "(nginx-vhosts) proxy:build-config (wildcard SSL and unrelated domain) 1" {
  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  TEST_APP="${TEST_APP}.example.com"
  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  setup_test_tls wildcard

  run deploy_app nodejs-express dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP | grep -e '*.${DOKKU_DOMAIN}' | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"
}

@test "(nginx-vhosts) proxy:build-config (with SSL and Multiple SANs)" {
  setup_test_tls sans
  run /bin/bash -c "dokku domains:add $TEST_APP test.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP www.test.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_ssl_domain "test.${DOKKU_DOMAIN}"
  assert_ssl_domain "www.test.${DOKKU_DOMAIN}"
  assert_ssl_domain "www.test.app.${DOKKU_DOMAIN}"
}
