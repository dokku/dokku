#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku haproxy:set --global letsencrypt-server https://acme-staging-v02.api.letsencrypt.org/directory
  dokku haproxy:set --global letsencrypt-email
  dokku haproxy:set --global refresh-conf 2
  dokku haproxy:start
  create_app
}

teardown() {
  global_teardown
  destroy_app
  dokku haproxy:stop
  dokku haproxy:set --global refresh-conf
  dokku nginx:start
}

@test "(haproxy) haproxy:help" {
  run /bin/bash -c "dokku haproxy"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the haproxy proxy integration"
  help_output="$output"

  run /bin/bash -c "dokku haproxy:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the haproxy proxy integration"
  assert_output "$help_output"
}

@test "(haproxy) global-only keys" {
  for key in image log-level letsencrypt-email letsencrypt-server refresh-conf; do
    run /bin/bash -c "dokku haproxy:set $TEST_APP $key somevalue"
    echo "key: $key"
    echo "output: $output"
    echo "status: $status"
    assert_failure
    assert_output_contains "can only be set globally"
  done
}

@test "(haproxy) refresh-conf" {
  run /bin/bash -c "dokku haproxy:set $TEST_APP refresh-conf 2"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "can only be set globally"

  run /bin/bash -c "dokku haproxy:set --global refresh-conf 5"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:report --global --haproxy-global-refresh-conf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "5"

  run /bin/bash -c "dokku haproxy:report --global --haproxy-computed-refresh-conf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "5"

  run /bin/bash -c "dokku haproxy:show-config"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "EASYHAPROXY_REFRESH_CONF=5"

  run /bin/bash -c "dokku haproxy:set --global refresh-conf"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:report --global --haproxy-global-refresh-conf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku haproxy:report --global --haproxy-computed-refresh-conf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "10"
}

@test "(haproxy:report) --global raw and computed keys in --format json" {
  run /bin/bash -c "dokku haproxy:set --global refresh-conf"
  assert_success

  run /bin/bash -c "dokku --quiet haproxy:report --global --format json | jq -r '.\"global-image\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku --quiet haproxy:report --global --format json | jq -r '.\"computed-image\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  refute_output ""

  run /bin/bash -c "dokku --quiet haproxy:report --global --format json | jq -r '.\"global-refresh-conf\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku --quiet haproxy:report --global --format json | jq -r '.\"computed-refresh-conf\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "10"
}

@test "(haproxy) log-level" {
  run /bin/bash -c "dokku haproxy:set --global log-level DEBUG"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:stop"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:start"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(haproxy) single domain" {
  run /bin/bash -c "dokku proxy:set $TEST_APP haproxy"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP convert_to_dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response_contains "http" "$TEST_APP.$DOKKU_DOMAIN" "80" "/" "python/http.server"
}

@test "(haproxy) multiple domains" {
  run /bin/bash -c "dokku proxy:set $TEST_APP haproxy"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP $TEST_APP.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP $TEST_APP-2.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP convert_to_dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:inspect $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker logs haproxy-haproxy-1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response "http" "$TEST_APP.$DOKKU_DOMAIN" "80" "/" "python/http.server"
  assert_http_localhost_response "http" "$TEST_APP-2.$DOKKU_DOMAIN" "80" "/" "python/http.server"
}

@test "(haproxy) ssl" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:set $TEST_APP haproxy"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"haproxy.$TEST_APP-web.redirect_ssl\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku haproxy:set --global letsencrypt-email test@example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:stop"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:start"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:inspect $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"haproxy.$TEST_APP-web.redirect_ssl\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:5000 https:443:5000"
}

@test "(haproxy) label management" {
  run /bin/bash -c "dokku proxy:set $TEST_APP haproxy"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:labels:add $TEST_APP haproxy.directive value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:labels:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "haproxy.directive=value"

  run /bin/bash -c "dokku haproxy:labels:show $TEST_APP haproxy.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "value"

  run /bin/bash -c "dokku haproxy:labels:show $TEST_APP haproxy.directive2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"haproxy.directive\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "value"

  run /bin/bash -c "dokku haproxy:labels:remove $TEST_APP haproxy.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:labels:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "haproxy.directive=value"

  run /bin/bash -c "dokku haproxy:labels:show $TEST_APP haproxy.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"haproxy.directive\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
}
