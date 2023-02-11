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

  run deploy_app python dokku@dokku.me:$TEST_APP convert_to_dockerfile
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

  run /bin/bash -c "dokku domains:add $TEST_APP $TEST_APP.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP $TEST_APP-2.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@dokku.me:$TEST_APP convert_to_dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "curl --silent $TEST_APP.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "python/http.server"

  run /bin/bash -c "curl --silent $TEST_APP-2.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "python/http.server"
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

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"haproxy\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP.dokku.me:80"

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

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"haproxy\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP.dokku.me"

  run /bin/bash -c "dokku proxy:report $TEST_APP --proxy-port-map"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:5000"
}
