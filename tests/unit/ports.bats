#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app

  dokku --force postgres:destroy $TEST_APP 2>/dev/null || true
  dokku --force redis:destroy $TEST_APP 2>/dev/null || true
  dokku --force postgres:destroy $TEST_APP-other 2>/dev/null || true
  dokku --force redis:destroy $TEST_APP-other 2>/dev/null || true
}

teardown() {
  destroy_app

  dokku --force postgres:destroy $TEST_APP 2>/dev/null || true
  dokku --force redis:destroy $TEST_APP 2>/dev/null || true
  dokku --force postgres:destroy $TEST_APP-other 2>/dev/null || true
  dokku --force redis:destroy $TEST_APP-other 2>/dev/null || true

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

@test "(ports) ports subcommands (list/add/set/remove/clear)" {
  run /bin/bash -c "dokku ports:set $TEST_APP http:1234:5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:1234:5001"

  run /bin/bash -c "dokku ports:add $TEST_APP http:8080:5002 https:8443:5003"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:1234:5001 http:8080:5002 https:8443:5003"

  run /bin/bash -c "dokku ports:set $TEST_APP http:8080:5000 https:8443:5000 http:1234:5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ports:set $TEST_APP http:8080:5000 https:8443:5000 http:1234:5001 http:1234:5002"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ports:add $TEST_APP http:12345:5002 http:12345:5004"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ports:add $TEST_APP http:1234:5003 http:12345:5004"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ports:add $TEST_APP http:1234:5001"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:1234:5001 http:8080:5000 https:8443:5000"

  run /bin/bash -c "dokku ports:remove $TEST_APP 8080"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:1234:5001 https:8443:5000"

  run /bin/bash -c "dokku ports:remove $TEST_APP http:1234:5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output "https:8443:5000"

  run /bin/bash -c "dokku ports:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:5000"
}

@test "(ports:add) post-deploy add" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ports:add $TEST_APP http:8080:5000 http:8081:5000"
  echo "output: $output"
  echo "status: $status"
  assert_success

  URLS="$(dokku --quiet urls "$TEST_APP")"
  for URL in $URLS; do
    assert_http_success $URL
  done
  assert_http_success "http://$TEST_APP.${DOKKU_DOMAIN}:8080"
  assert_http_success "http://$TEST_APP.${DOKKU_DOMAIN}:8081"
}

@test "(ports:report) herokuish tls" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:5000"

  run setup_test_tls
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:5000 https:443:5000"
}

@test "(ports:report) dockerfile tls" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP move_expose_dockerfile_into_place
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:3000:3000 http:3003:3003 udp:3001:3001"

  run setup_test_tls
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_output "https:3000:3000 https:3003:3003 udp:3001:3001"
}

@test "(ports:rename) renamed app persists ports" {
  if ! dokku plugin:installed postgres; then
    run /bin/bash -c "dokku plugin:install https://github.com/dokku/dokku-postgres.git"
    echo "output: $output"
    echo "status: $status"
    assert_success
  fi

  if ! dokku plugin:installed redis; then
    run /bin/bash -c "dokku plugin:install https://github.com/dokku/dokku-redis.git"
    echo "output: $output"
    echo "status: $status"
    assert_success
  fi

  run /bin/bash -c "dokku ports:set $TEST_APP http:80:3000"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set $TEST_APP APP_SECRET=REPLACE_WITH_LONG_SECRET_REPLACE_WITH_LONG_SECRET"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku postgres:create $TEST_APP $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku postgres:link $TEST_APP $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku redis:create $TEST_APP $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku redis:link $TEST_APP $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku git:from-image $TEST_APP docmost/docmost:latest
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "http:80:3000"

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "http:3000:3000"

  run /bin/bash -c "dokku apps:rename $TEST_APP $TEST_APP-other"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run /bin/bash -c "dokku ports:report $TEST_APP-other --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "http:80:3000"

  run /bin/bash -c "dokku ports:report $TEST_APP-other --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "http:3000:3000"
}
