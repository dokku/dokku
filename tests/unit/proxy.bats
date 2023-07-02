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
  deploy_app
  assert_nonssl_domain "${TEST_APP}.dokku.me"

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
  assert_http_success "${TEST_APP}.dokku.me"

  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    assert_not_external_port $(<$CID_FILE)
  done
}
