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

@test "(ports) ports:help" {
  run /bin/bash -c "dokku ports"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage ports for an app"
  help_output="$output"

  run /bin/bash -c "dokku ports:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage ports for an app"
  assert_output "$help_output"
}

@test "(ports) list/add/set/remove/clear" {
  run /bin/bash -c "dokku ports:set $TEST_APP http:1234:5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "http 1234 5001"

  run /bin/bash -c "dokku ports:add $TEST_APP http:8080:5002 https:8443:5003"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "http 1234 5001 http 8080 5002 https 8443 5003"

  run /bin/bash -c "dokku ports:set $TEST_APP http:8080:5000 https:8443:5000 http:1234:5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "http 1234 5001 http 8080 5000 https 8443 5000"

  run /bin/bash -c "dokku ports:remove $TEST_APP 8080"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "http 1234 5001 https 8443 5000"

  run /bin/bash -c "dokku ports:remove $TEST_APP http:1234:5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "https 8443 5000"

  run /bin/bash -c "dokku ports:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "http 80 5000"
}

@test "(ports) ports:add (post-deploy add)" {
  deploy_app
  run /bin/bash -c "dokku ports:add $TEST_APP http:8080:5000 http:8081:5000"
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
