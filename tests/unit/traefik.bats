#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku traefik:set --global letsencrypt-server https://acme-staging-v02.api.letsencrypt.org/directory
  dokku traefik:set --global letsencrypt-email
  dokku traefik:start
  create_app
}

teardown() {
  global_teardown
  destroy_app
  dokku traefik:stop
  dokku nginx:start
}

@test "(traefik) traefik:help" {
  run /bin/bash -c "dokku traefik"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the traefik proxy integration"
  help_output="$output"

  run /bin/bash -c "dokku traefik:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the traefik proxy integration"
  assert_output "$help_output"
}

@test "(traefik) single domain" {
  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
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

@test "(traefik) multiple domains" {
  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
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

@test "(traefik) traefik:set priority" {
  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:set $TEST_APP priority 12345"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-http.loadbalancer.server.port\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_output "5000"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.routers.$TEST_APP-web-http.priority\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "12345"
}

@test "(traefik) ssl" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-http.loadbalancer.server.port\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "5000"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-https.loadbalancer.server.port\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku traefik:set --global letsencrypt-email test@example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:stop"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:start"
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

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-http.loadbalancer.server.port\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "5000"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-https.loadbalancer.server.port\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "5000"

  run /bin/bash -c "dokku proxy:report $TEST_APP --proxy-port-map"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:5000"
}
