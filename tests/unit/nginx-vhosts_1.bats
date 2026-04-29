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

@test "(nginx:report) --format json" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set $TEST_APP bind-address-ipv4 127.0.0.1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:report $TEST_APP --format json | jq -r '.\"bind-address-ipv4\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "127.0.0.1"

  run /bin/bash -c "dokku nginx:report $TEST_APP --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:report --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:report $TEST_APP --format json --nginx-bind-address-ipv4"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "--format flag cannot be specified when specifying an info flag"
}

@test "(nginx:report) --global --format json" {
  run /bin/bash -c "dokku nginx:report --global --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:report --global --format json | jq -r 'keys[]' | grep -v 'global-' | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku nginx:report --global"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "global nginx information"
}

@test "(nginx-vhosts) proxy:build-config (domains:disable/enable)" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  HOSTNAME=$(<"$DOKKU_ROOT/VHOST")
  check_urls http://${HOSTNAME}:[0-9]+

  URLS=$(dokku --quiet urls "$TEST_APP")
  for URL in $URLS; do
    assert_http_success $URL
  done

  run /bin/bash -c "dokku domains:enable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  check_urls "http://${TEST_APP}.${DOKKU_DOMAIN}"
  assert_http_success "http://${TEST_APP}.${DOKKU_DOMAIN}"
}

@test "(nginx-vhosts) proxy:build-config (domains:add pre deploy)" {
  run create_app
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
  sleep 5 # wait for nginx to reload

  check_urls "http://www.test.app.${DOKKU_DOMAIN}"
  assert_http_success "http://www.test.app.${DOKKU_DOMAIN}"
}

@test "(nginx-vhosts) proxy:build-config (with global VHOST)" {
  echo "${DOKKU_DOMAIN}" >"$DOKKU_ROOT/VHOST"
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  check_urls "http://${TEST_APP}.${DOKKU_DOMAIN}"
  assert_http_success "http://${TEST_APP}.${DOKKU_DOMAIN}"
}
