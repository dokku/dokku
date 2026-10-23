#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku openresty:set --global letsencrypt-server https://acme-staging-v02.api.letsencrypt.org/directory
  dokku openresty:set --global letsencrypt-email
  dokku openresty:start
  create_app
}

teardown() {
  global_teardown
  destroy_app
  dokku openresty:stop
  dokku nginx:start
}

@test "(openresty) includes persist across ps:restart" {
  run /bin/bash -c "dokku proxy:set $TEST_APP openresty"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_openresty_include
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"openresty.include-http-example.conf\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "IyBmb3JjZSB0aGUgY2hhcmFjdGVyIHNldCB0byB1dGYtOApjaGFyc2V0IFVURi04Owo="

  run /bin/bash -c "dokku ps:restart $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"openresty.include-http-example.conf\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "IyBmb3JjZSB0aGUgY2hhcmFjdGVyIHNldCB0byB1dGYtOApjaGFyc2V0IFVURi04Owo="

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"openresty.include-location-http-example.conf\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "IyBsb2NhdGlvbi1ibG9jawo="
}

@test "(openresty) includes are removed when deleted from the repo" {
  local DATA_DIR="/var/lib/dokku/data/openresty-vhosts/app-$TEST_APP"
  local CUSTOM_TMP=$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$CUSTOM_TMP"' INT TERM

  run /bin/bash -c "dokku proxy:set $TEST_APP openresty"
  echo "output: $output"
  echo "status: $status"
  assert_success

  CUSTOM_TMP="$CUSTOM_TMP" run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_openresty_include
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$CUSTOM_TMP" rm -r openresty
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$CUSTOM_TMP" commit -m 'Remove openresty includes'
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$CUSTOM_TMP" push target master:master
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"openresty.include-http-example.conf\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "find $DATA_DIR -maxdepth 1 -name 'openresty-*-includes*'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  rm -rf "$CUSTOM_TMP"
}

@test "(openresty) deploy removes stale include markers and directories" {
  local DATA_DIR="/var/lib/dokku/data/openresty-vhosts/app-$TEST_APP"
  mkdir -p "$DATA_DIR/openresty-http-includes.99999999" "$DATA_DIR/openresty-location-includes.99999999"
  echo "charset UTF-8;" >"$DATA_DIR/openresty-http-includes.99999999/example.conf"
  echo "# location-block" >"$DATA_DIR/openresty-location-includes.99999999/example.conf"
  touch "$DATA_DIR/openresty-http-includes.99999998.missing" "$DATA_DIR/openresty-location-includes.99999998.missing"
  chown -R dokku:dokku "$DATA_DIR"

  run /bin/bash -c "dokku proxy:set $TEST_APP openresty"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "find $DATA_DIR -maxdepth 1 -name 'openresty-*-includes*'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
}

add_openresty_include() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mkdir -p "$APP_REPO_DIR/openresty/http-includes"
  echo "# force the character set to utf-8" >>"$APP_REPO_DIR/openresty/http-includes/example.conf"
  echo "charset UTF-8;" >>"$APP_REPO_DIR/openresty/http-includes/example.conf"

  mkdir -p "$APP_REPO_DIR/openresty/http-location-includes"
  echo "# location-block" >>"$APP_REPO_DIR/openresty/http-location-includes/example.conf"
}
