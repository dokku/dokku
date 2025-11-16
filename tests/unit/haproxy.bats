#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku haproxy:set --global letsencrypt-server https://acme-staging-v02.api.letsencrypt.org/directory
  dokku haproxy:set --global letsencrypt-email
  dokku haproxy:start
  create_app
}

teardown() {
  global_teardown
  destroy_app
  dokku haproxy:stop
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

  run /bin/bash -c "curl --silent $(dokku url $TEST_APP)"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python/http.server"
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

  sleep 5

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

  run /bin/bash -c "dokku haproxy:label:add $TEST_APP haproxy.directive value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:label:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "haproxy.directive=value"

  run /bin/bash -c "dokku haproxy:label:show $TEST_APP haproxy.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "value"

  run /bin/bash -c "dokku haproxy:label:show $TEST_APP haproxy.directive2"
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

  run /bin/bash -c "dokku haproxy:label:remove $TEST_APP haproxy.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku haproxy:label:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "haproxy.directive=value"

  run /bin/bash -c "dokku haproxy:label:show $TEST_APP haproxy.directive"
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
