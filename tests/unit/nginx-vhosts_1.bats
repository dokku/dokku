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

@test "(nginx) nginx:help" {
  run /bin/bash -c "dokku nginx"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the nginx proxy"
  help_output="$output"

  run /bin/bash -c "dokku nginx:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the nginx proxy"
  assert_output "$help_output"
}

@test "(nginx-vhosts) nginx:build-config (domains:disable/enable)" {
  deploy_app
  dokku domains:disable $TEST_APP

  HOSTNAME=$(<"$DOKKU_ROOT/VHOST")
  check_urls http://${HOSTNAME}:[0-9]+

  URLS=$(dokku --quiet urls "$TEST_APP")
  for URL in $URLS; do
    assert_http_success $URL
  done

  dokku domains:enable $TEST_APP
  check_urls http://${TEST_APP}.dokku.me
  assert_http_success http://${TEST_APP}.dokku.me
}

@test "(nginx-vhosts) nginx:build-config (domains:add pre deploy)" {
  create_app
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  deploy_app
  sleep 5 # wait for nginx to reload

  check_urls http://www.test.app.dokku.me
  assert_http_success http://www.test.app.dokku.me
}

@test "(nginx-vhosts) nginx:build-config (with global VHOST)" {
  echo "dokku.me" >"$DOKKU_ROOT/VHOST"
  deploy_app

  check_urls http://${TEST_APP}.dokku.me
  assert_http_success http://${TEST_APP}.dokku.me
}
