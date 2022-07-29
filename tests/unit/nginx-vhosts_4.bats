#!/usr/bin/env bats

load test_helper
source "$PLUGIN_CORE_AVAILABLE_PATH/config/functions"

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) nginx:build-config (without global VHOST)" {
  rm "$DOKKU_ROOT/VHOST"
  deploy_app

  run /bin/bash -c "dokku --quiet urls $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  HOSTNAME="$(hostname -f)"
  check_urls http://${HOSTNAME}:[0-9]+

  URLS=$(dokku --quiet urls "$TEST_APP")
  for URL in $URLS; do
    assert_http_success $URL
  done
}

@test "(nginx-vhosts) nginx:build-config (without global VHOST and IPv4 address set as HOSTNAME)" {
  rm "$DOKKU_ROOT/VHOST"
  echo "127.0.0.1" >"$DOKKU_ROOT/VHOST"
  deploy_app

  HOSTNAME=$(<"$DOKKU_ROOT/VHOST")
  check_urls http://${HOSTNAME}:[0-9]+

  URLS=$(dokku --quiet urls "$TEST_APP")
  for URL in $URLS; do
    assert_http_success $URL
  done
}

@test "(nginx-vhosts) nginx:build-config (without global VHOST and IPv6 address set as HOSTNAME)" {
  rm "$DOKKU_ROOT/VHOST"
  echo "fda5:c7db:a520:bb6d::aabb:ccdd:eeff" >"$DOKKU_ROOT/VHOST"
  deploy_app

  HOSTNAME=$(<"$DOKKU_ROOT/VHOST")
  check_urls http://${HOSTNAME}:[0-9]+
}

@test "(nginx-vhosts) nginx:build-config (without global VHOST and domains:add pre deploy)" {
  rm "$DOKKU_ROOT/VHOST"
  create_app
  add_domain "www.test.app.dokku.me"
  deploy_app
  assert_nonssl_domain "www.test.app.dokku.me"
}

@test "(nginx-vhosts) nginx:build-config (without global VHOST and domains:add post deploy)" {
  rm "$DOKKU_ROOT/VHOST"
  deploy_app
  add_domain "www.test.app.dokku.me"
  check_urls http://www.test.app.dokku.me
  assert_http_success http://www.test.app.dokku.me
}
