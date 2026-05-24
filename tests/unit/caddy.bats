#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku caddy:set --global letsencrypt-server https://acme-staging-v02.api.letsencrypt.org/directory
  dokku caddy:set --global letsencrypt-email
  dokku caddy:start
  create_app
}

teardown() {
  global_teardown
  destroy_app
  dokku caddy:stop
  dokku nginx:start
}

@test "(caddy:report) --global global vs computed log-level" {
  run /bin/bash -c "dokku caddy:report --global --caddy-global-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku caddy:report --global --caddy-computed-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ERROR"

  run /bin/bash -c "dokku caddy:set --global log-level DEBUG"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:report --global --caddy-global-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "DEBUG"

  run /bin/bash -c "dokku caddy:report --global --caddy-computed-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "DEBUG"

  run /bin/bash -c "dokku caddy:set --global log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:report --global --caddy-global-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku caddy:report --global --caddy-computed-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ERROR"
}

@test "(caddy:report) --global raw and computed keys in --format json" {
  run /bin/bash -c "dokku caddy:report --global --format json | jq -r '.\"global-image\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku caddy:report --global --format json | jq -r '.\"computed-image\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku caddy:report --global --format json | jq -r '.\"global-polling-interval\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku caddy:report --global --format json | jq -r '.\"computed-polling-interval\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "5s"
}

@test "(caddy) global-only keys" {
  for key in image letsencrypt-email letsencrypt-server log-level polling-interval; do
    run /bin/bash -c "dokku caddy:set $TEST_APP $key somevalue"
    echo "key: $key"
    echo "output: $output"
    echo "status: $status"
    assert_failure
    assert_output_contains "can only be set globally"
  done

  run /bin/bash -c "dokku caddy:set $TEST_APP tls-internal true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku caddy:set $TEST_APP tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(caddy:report) tls-internal raw vs computed vs global" {
  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-computed-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-global-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku caddy:set $TEST_APP tls-internal true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-computed-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku caddy:set $TEST_APP tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-computed-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"
}

@test "(caddy:set) --global tls-internal" {
  run /bin/bash -c "dokku caddy:set --global tls-internal true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:report --global --caddy-global-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-computed-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-global-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku caddy:set $TEST_APP tls-internal false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku caddy:report $TEST_APP --caddy-computed-tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku caddy:set $TEST_APP tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:set --global tls-internal"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(caddy) caddy:help" {
  run /bin/bash -c "dokku caddy"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the caddy proxy integration"
  help_output="$output"

  run /bin/bash -c "dokku caddy:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the caddy proxy integration"
  assert_output "$help_output"
}

@test "(caddy) single domain" {
  run /bin/bash -c "dokku proxy:set $TEST_APP caddy"
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

@test "(caddy) multiple domains" {
  run /bin/bash -c "dokku proxy:set $TEST_APP caddy"
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

  run /bin/bash -c "curl --silent $TEST_APP.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "python/http.server"

  run /bin/bash -c "curl --silent $TEST_APP-2.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "python/http.server"
}

@test "(caddy) ssl" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:set $TEST_APP caddy"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"caddy\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP.${DOKKU_DOMAIN}:80"

  run /bin/bash -c "dokku caddy:set --global letsencrypt-email test@example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:stop"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:start"
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

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"caddy\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP.${DOKKU_DOMAIN}"

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:5000 https:443:5000"
}

@test "(caddy) label management" {
  run /bin/bash -c "dokku proxy:set $TEST_APP caddy"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:labels:add $TEST_APP caddy.directive value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:labels:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "caddy.directive=value"

  run /bin/bash -c "dokku caddy:labels:show $TEST_APP caddy.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "value"

  run /bin/bash -c "dokku caddy:labels:show $TEST_APP caddy.directive2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"caddy.directive\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "value"

  run /bin/bash -c "dokku caddy:labels:remove $TEST_APP caddy.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku caddy:labels:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "caddy.directive=value"

  run /bin/bash -c "dokku caddy:labels:show $TEST_APP caddy.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"caddy.directive\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
}
