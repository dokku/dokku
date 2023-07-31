#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(init) buildpack" {
  source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
  local APP="zombies-buildpack"
  run deploy_app "$APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  local CIDS=$(get_app_container_ids "$APP")

  run "$DOCKER_BIN" container top "$CIDS"
  echo "output: $output"
  assert_output_contains "<defunct>" "0"
}

@test "(init) dockerfile no tini" {
  source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
  local APP="zombies-dockerfile-no-tini"
  run deploy_app "$APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  local CIDS=$(get_app_container_ids "$APP")

  run "$DOCKER_BIN" container top "$CIDS"
  echo "output: $output"
  assert_output_contains "<defunct>" "0"
}

@test "(init) dockerfile with tini" {
  source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
  local APP="zombies-dockerfile-tini"
  run deploy_app "$APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  local CIDS=$(get_app_container_ids "$APP")

  run "$DOCKER_BIN" container top "$CIDS"
  echo "output: $output"
  assert_output_contains "<defunct>" "0"
}
