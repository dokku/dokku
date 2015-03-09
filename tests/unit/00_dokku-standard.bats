#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  rm -rf /home/dokku/$TEST_APP/tls /home/dokku/tls
  destroy_app
  disable_tls_wildcard
}

assert_urls() {
  urls=$@
  run dokku urls $TEST_APP
  echo "output: "$output
  echo "status: "$status
  assert_output < <(tr ' ' '\n' <<< "${urls}")
}

build_nginx_config() {
  # simulate nginx post-deploy
  dokku domains:setup $TEST_APP
  dokku nginx:build-config $TEST_APP
}

@test "run (with tty)" {
  deploy_app
  run /bin/bash -c "dokku run $TEST_APP ls /app/package.json"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "run (without tty)" {
  deploy_app
  run /bin/bash -c ": |dokku run $TEST_APP ls /app/package.json"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "run (with --options)" {
  deploy_app
  run /bin/bash -c "dokku --force --quiet run $TEST_APP node --version"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "urls (non-ssl)" {
  assert_urls "http://dokku.me"
  build_nginx_config
  assert_urls "http://${TEST_APP}.dokku.me"
  add_domain "test.dokku.me"
  assert_urls "http://${TEST_APP}.dokku.me" "http://test.dokku.me"
}

@test "urls (app ssl)" {
  setup_test_tls
  assert_urls "https://dokku.me"
  build_nginx_config
  assert_urls "https://node-js-app.dokku.me" "http://${TEST_APP}.dokku.me"
  add_domain "test.dokku.me"
  assert_urls "https://node-js-app.dokku.me" "http://${TEST_APP}.dokku.me" "http://test.dokku.me"
}

@test "urls (wildcard ssl)" {
  setup_test_tls_wildcard
  assert_urls "https://dokku.me"
  build_nginx_config
  assert_urls "https://${TEST_APP}.dokku.me"
  add_domain "test.dokku.me"
  assert_urls "https://${TEST_APP}.dokku.me" "https://test.dokku.me"
  add_domain "dokku.example.com"
  assert_urls "https://${TEST_APP}.dokku.me" "https://test.dokku.me" "http://dokku.example.com"
}
