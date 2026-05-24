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

@test "(proxy) proxy:help" {
  run /bin/bash -c "dokku proxy"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the proxy integration for an app"
  help_output="$output"

  run /bin/bash -c "dokku proxy:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the proxy integration for an app"
  assert_output "$help_output"
}

@test "(proxy:report) --global raw vs computed type" {
  run /bin/bash -c "dokku proxy:report --global --format json | jq -r '.\"proxy-global-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku proxy:report --global --format json | jq -r '.\"proxy-computed-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "nginx"

  run /bin/bash -c "dokku proxy:set --global type caddy"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:report --global --format json | jq -r '.\"proxy-global-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "caddy"

  run /bin/bash -c "dokku proxy:report --global --format json | jq -r '.\"proxy-computed-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "caddy"

  run /bin/bash -c "dokku proxy:set --global type"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:report --global --format json | jq -r '.\"proxy-global-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku proxy:report --global --format json | jq -r '.\"proxy-computed-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "nginx"
}

@test "(proxy:report) type raw vs computed vs global" {
  run /bin/bash -c "dokku proxy:set --global type"
  assert_success

  run /bin/bash -c "dokku proxy:report $TEST_APP --format json | jq -r '.\"proxy-type\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku proxy:report $TEST_APP --format json | jq -r '.\"proxy-global-type\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku proxy:report $TEST_APP --format json | jq -r '.\"proxy-computed-type\"'"
  assert_success
  assert_output "nginx"

  run /bin/bash -c "dokku proxy:set --global type caddy"
  assert_success

  run /bin/bash -c "dokku proxy:report $TEST_APP --format json | jq -r '.\"proxy-global-type\"'"
  assert_success
  assert_output "caddy"

  run /bin/bash -c "dokku proxy:report $TEST_APP --format json | jq -r '.\"proxy-computed-type\"'"
  assert_success
  assert_output "caddy"

  run /bin/bash -c "dokku proxy:set $TEST_APP type traefik"
  assert_success

  run /bin/bash -c "dokku proxy:report $TEST_APP --format json | jq -r '.\"proxy-type\"'"
  assert_success
  assert_output "traefik"

  run /bin/bash -c "dokku proxy:report $TEST_APP --format json | jq -r '.\"proxy-global-type\"'"
  assert_success
  assert_output "caddy"

  run /bin/bash -c "dokku proxy:report $TEST_APP --format json | jq -r '.\"proxy-computed-type\"'"
  assert_success
  assert_output "traefik"

  run /bin/bash -c "dokku proxy:set $TEST_APP type"
  assert_success

  run /bin/bash -c "dokku proxy:set --global type"
  assert_success
}

@test "(proxy:set) invalid port mapping set" {
  run /bin/bash -c "dokku proxy:set $TEST_APP http:80:80"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Detected potential port mapping instead of proxy type"
}

@test "(proxy) proxy:build-config/clear-config" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "rm -f $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:clear-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:clear-config --all"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(proxy) proxy:enable/disable" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_nonssl_domain "${TEST_APP}.${DOKKU_DOMAIN}"

  run /bin/bash -c "dokku proxy:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    assert_not_external_port $(<$CID_FILE)
  done

  run /bin/bash -c "dokku proxy:enable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success "${TEST_APP}.${DOKKU_DOMAIN}"

  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    assert_not_external_port $(<$CID_FILE)
  done
}
