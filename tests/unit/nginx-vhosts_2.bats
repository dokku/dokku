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
  run /bin/bash -c "dokku domains:add $TEST_APP node-js-app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP test.dokku.me"
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

  assert_ssl_domain "node-js-app.dokku.me"
  assert_http_redirect "http://test.dokku.me" "https://test.dokku.me:443/"
}

@test "(nginx-vhosts) proxy:build-config (wildcard SSL)" {
  setup_test_tls wildcard
  run /bin/bash -c "dokku domains:add $TEST_APP wildcard1.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP wildcard2.dokku.me"
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

  assert_ssl_domain "wildcard1.dokku.me"
  assert_ssl_domain "wildcard2.dokku.me"
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

  run deploy_app nodejs-express dokku@dokku.me:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP | grep -e '*.dokku.me' | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"
}

@test "(nginx-vhosts) proxy:build-config (with SSL and Multiple SANs)" {
  setup_test_tls sans
  run /bin/bash -c "dokku domains:add $TEST_APP test.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP www.test.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_ssl_domain "test.dokku.me"
  assert_ssl_domain "www.test.dokku.me"
  assert_ssl_domain "www.test.app.dokku.me"
}
