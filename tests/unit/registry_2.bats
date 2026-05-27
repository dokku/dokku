#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  dokku checks:set $TEST_APP wait-to-retire 30
}

teardown() {
  destroy_app
  global_teardown
}

setup_push_on_release_app() {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  run /bin/bash -c "dokku builder:set $TEST_APP selected herokuish"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:set $TEST_APP nginx"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:login $TEST_APP docker.io $DOCKERHUB_USERNAME $DOCKERHUB_TOKEN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set $TEST_APP image-repo $DOCKERHUB_USERNAME/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set $TEST_APP push-on-release true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image rm $DOCKERHUB_USERNAME/$TEST_APP:1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image inspect $DOCKERHUB_USERNAME/$TEST_APP:1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(registry) [recover] ps:restart pulls missing image" {
  setup_push_on_release_app

  run /bin/bash -c "dokku ps:restart $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "not found locally, pulling from registry"

  run /bin/bash -c "docker image inspect $DOCKERHUB_USERNAME/$TEST_APP:1"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(registry) [recover] dokku run pulls missing image" {
  setup_push_on_release_app

  run /bin/bash -c "dokku run $TEST_APP echo recovery-ok"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "not found locally, pulling from registry"
  assert_output_contains "recovery-ok"
}

@test "(registry) [recover] domains:add pulls missing image" {
  setup_push_on_release_app

  run /bin/bash -c "dokku domains:add $TEST_APP recovery.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image inspect $DOCKERHUB_USERNAME/$TEST_APP:1"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
