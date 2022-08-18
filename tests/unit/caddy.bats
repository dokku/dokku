#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku caddy:start
  create_app
}

teardown() {
  global_teardown
  destroy_app
  dokku caddy:stop
  dokku nginx:start
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

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "curl $(dokku url $TEST_APP)"
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

  run /bin/bash -c "dokku domains:add $TEST_APP $TEST_APP.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP $TEST_APP-2.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "curl $TEST_APP.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "python/http.server"

  run /bin/bash -c "curl $TEST_APP-2.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "python/http.server"
}
