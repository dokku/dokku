#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  dokku checks:set $TEST_APP wait-to-retire=30
}

teardown() {
  destroy_app
  global_teardown
}

@test "(registry) registry:help" {
  run /bin/bash -c "dokku registry"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage registry settings for an app"
  help_output="$output"

  run /bin/bash -c "dokku registry:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage registry settings for an app"
  assert_output "$help_output"
}

@test "(registry) registry:login" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  run /bin/bash -c "dokku registry:login docker.io $DOCKERHUB_USERNAME $DOCKERHUB_TOKEN"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Login Succeeded"

  run /bin/bash -c "echo $DOCKERHUB_TOKEN | dokku registry:login docker.io --password-stdin $DOCKERHUB_USERNAME"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Login Succeeded"
}

@test "(registry) registry:set server" {
  run /bin/bash -c "dokku registry:set --global server ghcr.io"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-global-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ghcr.io"

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-computed-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ghcr.io/"

  run /bin/bash -c "dokku registry:set --global server docker.io"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-global-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker.io"

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-computed-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku registry:set $TEST_APP server docker.io"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-computed-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-global-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker.io"

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker.io"

  run /bin/bash -c "dokku registry:set $TEST_APP server ghcr.io"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-computed-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ghcr.io/"

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-global-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker.io"

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ghcr.io"
}

@test "(registry) registry:set image-repo" {
  run /bin/bash -c "docker images"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set $TEST_APP image-repo heroku/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect heroku/$TEST_APP:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker images"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(registry) registry:set push-on-release" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  run /bin/bash -c "dokku registry:set $TEST_APP push-on-release true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set $TEST_APP image-repo dokku/test-app"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 60

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container ls -a"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image ls"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 60

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container ls -a"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image ls"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set $TEST_APP key=VALUE"
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 60

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container ls -a"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image ls"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 60

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container ls -a"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image ls"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image inspect dokku/test-app:1"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "docker image inspect dokku/test-app:2"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "docker image inspect dokku/test-app:3"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "docker image inspect dokku/test-app:4"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
