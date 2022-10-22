#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
}

teardown() {
  destroy_app 0 $TEST_APP
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) nginx:set client-max-body-size" {
  deploy_app

  run /bin/bash -c "dokku nginx:set $TEST_APP client-max-body-size"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "client_max_body_size" 0

  run /bin/bash -c "dokku nginx:set $TEST_APP client-max-body-size 1m"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "client_max_body_size 1m;" 1
}

@test "(nginx-vhosts) nginx:set proxy-read-timeout" {
  deploy_app

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-read-timeout 45s"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "45s;"

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-read-timeout"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "45s;" 0
}

@test "(nginx-vhosts) nginx:set proxy-read-timeout (with SSL)" {
  setup_test_tls
  deploy_app

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-read-timeout 45s"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "45s;"

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-read-timeout"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "45s;" 0
}

@test "(nginx-vhosts) nginx:build-config ignore bad https mapping" {
  setup_test_tls
  run deploy_app "dockerfile-noexpose"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Ignoring detected https port mapping without an accompanying ssl certificate" 0

  teardown_test_tls
  run /bin/bash -c "dokku proxy:report $TEST_APP --proxy-port-map"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:5000 https:443:5000"

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Ignoring detected https port mapping without an accompanying ssl certificate" 1

  run /bin/bash -c "dokku proxy:report $TEST_APP --proxy-port-map"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:5000 https:443:5000"
}
