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
  docker network rm create-network || true
  docker network rm deploy-network || true
  docker network rm initial-network || true
  global_teardown
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
    assert_external_port $(<$CID_FILE)
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
    assert_not_external_port $(<$CID_FILE)
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

  run /bin/bash -c "curl --silent --write-out '%{http_code}\n' $(dokku url $TEST_APP) | grep 200"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(network) network management" {
  run /bin/bash -c "dokku network:list | grep $TEST_NETWORK"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku network:exists $TEST_NETWORK"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku network:create $TEST_NETWORK"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:list | grep $TEST_NETWORK"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:create $TEST_NETWORK"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku network:exists $TEST_NETWORK"
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

  run /bin/bash -c "dokku --force network:destroy $TEST_NETWORK"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force network:destroy $TEST_NETWORK"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku network:exists $TEST_NETWORK"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku network:list | grep $TEST_NETWORK"
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

  run /bin/bash -c "dokku network:create initial-network"
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

  run /bin/bash -c "dokku network:set $TEST_APP initial-network initial-network"
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

  run /bin/bash -c "dokku --force network:destroy initial-network"
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

  run /bin/bash -c "dokku --force network:destroy initial-network"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(network) dont re-attach to network" {
  run /bin/bash -c "dokku network:create deploy-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku network:set $TEST_APP attach-post-deploy deploy-network"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app dockerfile-procfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP worker=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=3"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container inspect $TEST_APP.web.1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container inspect $TEST_APP.web.2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container inspect $TEST_APP.web.3"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container inspect $TEST_APP.worker.1"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(network) handle single-udp app" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP -p 1194:1194"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku checks:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # The app _should_ fail to deploy because it requires some config file
  # but we really only care about getting to pre-flight checks so that is fine.
  # It does not because of some yet unknown bug in deploying when zero-downtime is disable...
  run /bin/bash -c "dokku git:from-image $TEST_APP kylemanna/openvpn:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Attempting pre-flight checks" 0
  assert_output_contains "Application deployed"
}
