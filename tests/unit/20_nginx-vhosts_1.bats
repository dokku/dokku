#!/usr/bin/env bats

load test_helper
source "$PLUGIN_CORE_AVAILABLE_PATH/config/functions"

setup() {
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  [[ -f "$DOKKU_ROOT/HOSTNAME" ]] && cp -fp "$DOKKU_ROOT/HOSTNAME" "$DOKKU_ROOT/HOSTNAME.bak"
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  [[ -f "$DOKKU_ROOT/HOSTNAME.bak" ]] && mv "$DOKKU_ROOT/HOSTNAME.bak" "$DOKKU_ROOT/HOSTNAME" && chown dokku:dokku "$DOKKU_ROOT/HOSTNAME"
}

@test "(nginx-vhosts) nginx:build-config (domains:disable/enable)" {
  deploy_app
  dokku domains:disable $TEST_APP

  HOSTNAME=$(< "$DOKKU_ROOT/HOSTNAME")
  check_urls http://${HOSTNAME}:[0-9]+

  NGINX_PORT="$(get_config_value $TEST_APP DOKKU_NGINX_PORT)"
  assert_http_success http://${HOSTNAME}:${NGINX_PORT}

  dokku domains:enable $TEST_APP
  check_urls http://${TEST_APP}.dokku.me
  assert_http_success http://${TEST_APP}.dokku.me
}

@test "(nginx-vhosts) nginx:build-config (domains:add pre deploy)" {
  create_app
  run dokku domains:add $TEST_APP www.test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success

  deploy_app
  sleep 5 # wait for nginx to reload

  check_urls http://www.test.app.dokku.me
  assert_http_success http://www.test.app.dokku.me
}

@test "(nginx-vhosts) nginx:build-config (with global VHOST)" {
  echo "dokku.me" > "$DOKKU_ROOT/VHOST"
  deploy_app

  check_urls http://${TEST_APP}.dokku.me
  assert_http_success http://${TEST_APP}.dokku.me
}

@test "(nginx-vhosts) nginx:build-config (without global VHOST but real HOSTNAME)" {
  rm "$DOKKU_ROOT/VHOST"
  echo "${TEST_APP}.dokku.me" > "$DOKKU_ROOT/HOSTNAME"
  deploy_app

  HOSTNAME=$(< "$DOKKU_ROOT/HOSTNAME")
  check_urls http://${HOSTNAME}:[0-9]+

  NGINX_PORT="$(get_config_value $TEST_APP DOKKU_NGINX_PORT)"
  assert_http_success http://${HOSTNAME}:${NGINX_PORT}
}

@test "(nginx-vhosts) nginx:build-config (without global VHOST and IPv4 address set as HOSTNAME)" {
  rm "$DOKKU_ROOT/VHOST"
  echo "127.0.0.1" > "$DOKKU_ROOT/HOSTNAME"
  deploy_app

  HOSTNAME=$(< "$DOKKU_ROOT/HOSTNAME")
  check_urls http://${HOSTNAME}:[0-9]+

  NGINX_PORT="$(get_config_value $TEST_APP DOKKU_NGINX_PORT)"
  assert_http_success http://${HOSTNAME}:${NGINX_PORT}
}

@test "(nginx-vhosts) nginx:build-config (without global VHOST and IPv6 address set as HOSTNAME)" {
  rm "$DOKKU_ROOT/VHOST"
  echo "fda5:c7db:a520:bb6d::aabb:ccdd:eeff" > "$DOKKU_ROOT/HOSTNAME"
  deploy_app

  HOSTNAME=$(< "$DOKKU_ROOT/HOSTNAME")
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

@test "(nginx-vhosts) nginx:build-config (xip.io style hostnames)" {
  echo "127.0.0.1.xip.io" > "$DOKKU_ROOT/VHOST"
  deploy_app

  check_urls http://${TEST_APP}.127.0.0.1.xip.io
  assert_http_success http://${TEST_APP}.127.0.0.1.xip.io
}

@test "(nginx-vhosts) nginx:build-config (dockerfile expose)" {
  deploy_app dockerfile

  add_domain "www.test.app.dokku.me"
  check_urls http://${TEST_APP}.dokku.me:3000
  check_urls http://${TEST_APP}.dokku.me:3003
  check_urls http://www.test.app.dokku.me:3000
  check_urls http://www.test.app.dokku.me:3003
  assert_http_success http://${TEST_APP}.dokku.me:3000
  assert_http_success http://${TEST_APP}.dokku.me:3003
  assert_http_success http://www.test.app.dokku.me:3000
  assert_http_success http://www.test.app.dokku.me:3003

}
