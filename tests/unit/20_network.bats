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
  docker network rm create-network || true
  docker network rm deploy-network || true
  global_teardown
}

assert_nonssl_domain() {
  local domain=$1
  assert_app_domain "${domain}"
  assert_http_success "http://${domain}"
}

assert_app_domain() {
  local domain=$1
  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null | grep -xF ${domain}"
  echo "output: $output"
  echo "status: $status"
  assert_output "${domain}"
}

assert_external_port() {
  local CID="$1"; local exit_status="$2"
  local EXTERNAL_PORT_COUNT=$(docker port $CID | wc -l)
  run /bin/bash -c "[[ $EXTERNAL_PORT_COUNT -gt 0 ]]"
  if [[ "$exit_status" == "success" ]]; then
    assert_success
  else
    assert_failure
  fi
}

@test "(network) network:help" {
  run /bin/bash -c "dokku network"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage network settings for an app"
  help_output="$output"

  run /bin/bash -c "dokku network:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage network settings for an app"
  assert_output "$help_output"
}

@test "(network) network:set bind-all-interfaces" {
  deploy_app
  assert_nonssl_domain "${TEST_APP}.dokku.me"

  run /bin/bash -c "dokku network:set $TEST_APP bind-all-interfaces true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success "${TEST_APP}.dokku.me"

  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    assert_external_port $(< $CID_FILE) success
  done

  run /bin/bash -c "dokku network:set $TEST_APP bind-all-interfaces false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success "${TEST_APP}.dokku.me"

  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.web.*; do
    assert_external_port $(< $CID_FILE) failure
  done
}

@test "(network) network host-mode" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy \"--network=host\""
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "curl --silent --write-out '%{http_code}\n' `dokku url $TEST_APP` | grep 200"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(network) network management" {
  run /bin/bash -c "dokku network:list | grep test-network"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku network:exists test-network"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku network:create test-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:list | grep test-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:create test-network"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku network:exists test-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:exists nonexistent-network"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku --force network:destroy nonexistent-network"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku --force network:destroy test-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force network:destroy test-network"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku network:exists test-network"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku network:list | grep test-network"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(network) network:set attach" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:create create-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:create deploy-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:set $TEST_APP attach-post-create nonexistent-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:set $TEST_APP attach-post-deploy nonexistent-network"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_http_success "${TEST_APP}.dokku.me"

  run /bin/bash -c "dokku network:set $TEST_APP attach-post-create create-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:set $TEST_APP attach-post-deploy create-network"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku network:set $TEST_APP attach-post-deploy deploy-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success "${TEST_APP}.dokku.me"

  run /bin/bash -c "dokku --force network:destroy create-network"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # necessary in order to remove networks in use by "dead" containers
  docker container prune --force

  run /bin/bash -c "dokku --force network:destroy create-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force network:destroy deploy-network"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
