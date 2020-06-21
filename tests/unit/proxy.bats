#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  [[ -f "$DOKKU_ROOT/HOSTNAME" ]] && cp -fp "$DOKKU_ROOT/HOSTNAME" "$DOKKU_ROOT/HOSTNAME.bak"
  create_app
}

teardown() {
  destroy_app 0 $TEST_APP
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  [[ -f "$DOKKU_ROOT/HOSTNAME.bak" ]] && mv "$DOKKU_ROOT/HOSTNAME.bak" "$DOKKU_ROOT/HOSTNAME" && chown dokku:dokku "$DOKKU_ROOT/HOSTNAME"
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

@test "(proxy) proxy:enable/disable" {
  deploy_app
  assert_nonssl_domain "${TEST_APP}.dokku.me"

  run /bin/bash -c "dokku proxy:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    assert_not_external_port $(< $CID_FILE)
  done

  run /bin/bash -c "dokku proxy:enable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success "${TEST_APP}.dokku.me"

  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    assert_not_external_port $(< $CID_FILE)
  done
}

@test "(proxy) proxy:ports (list/add/set/remove/clear)" {
  run /bin/bash -c "dokku proxy:ports-set $TEST_APP http:1234:5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet proxy:ports $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "http 1234 5001"

  run /bin/bash -c "dokku proxy:ports-add $TEST_APP http:8080:5002 https:8443:5003"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet proxy:ports $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "http 1234 5001 http 8080 5002 https 8443 5003"

  run /bin/bash -c "dokku proxy:ports-set $TEST_APP http:8080:5000 https:8443:5000 http:1234:5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet proxy:ports $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "http 1234 5001 http 8080 5000 https 8443 5000"

  run /bin/bash -c "dokku proxy:ports-remove $TEST_APP 8080"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet proxy:ports $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "http 1234 5001 https 8443 5000"

  run /bin/bash -c "dokku proxy:ports-remove $TEST_APP http:1234:5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet proxy:ports $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "https 8443 5000"

  run /bin/bash -c "dokku proxy:ports-clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet proxy:ports $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "http 80 5000"
}

@test "(proxy) proxy:ports (post-deploy add)" {
  deploy_app
  run /bin/bash -c "dokku proxy:ports-add $TEST_APP http:8080:5000 http:8081:5000"
  echo "output: $output"
  echo "status: $status"
  assert_success

  URLS="$(dokku --quiet urls "$TEST_APP")"
  for URL in $URLS; do
    assert_http_success $URL
  done
  assert_http_success "http://$TEST_APP.dokku.me:8080"
  assert_http_success "http://$TEST_APP.dokku.me:8081"
}
