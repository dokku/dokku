#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(builder) builder-detect [cnb]" {
  local TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM

  # test project.toml
  run touch "$TMP/project.toml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  chown -R dokku:dokku "$TMP"
  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP | tail -n1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "cnb"

  sudo rm -rf $TMP/*
  echo "output: $output"
  echo "status: $status"
  assert_success

  # test DOKKU_CNB_EXPERIMENTAL env var
  run /bin/bash -c "dokku config:set $TEST_APP DOKKU_CNB_EXPERIMENTAL=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  chown -R dokku:dokku "$TMP"
  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP | tail -n1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "cnb"
}

@test "(builder) builder-detect [dockerfile]" {
  local TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM

  run touch "$TMP/Dockerfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  chown -R dokku:dokku "$TMP"
  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP | tail -n1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dockerfile"
}

@test "(builder) builder-detect [herokuish]" {
  local TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM

  touch "$TMP/Dockerfile"

  # test buildpacks
  chown -R dokku:dokku "$TMP"
  run /bin/bash -c "touch $TMP/.buildpacks"
  echo "output: $output"
  echo "status: $status"
  assert_success

  chown -R dokku:dokku "$TMP"
  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP | tail -n1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "herokuish"

  sudo rm -rf $TMP/*
  echo "output: $output"
  echo "status: $status"
  assert_success

  # test .env
  run /bin/bash -c "echo BUILDPACK_URL=null > $TMP/.env"
  echo "output: $output"
  echo "status: $status"
  assert_success

  chown -R dokku:dokku "$TMP"
  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP | tail -n1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "herokuish"

  sudo rm -rf $TMP/*
  echo "output: $output"
  echo "status: $status"
  assert_success

  # test BUILDPACK_URL env var
  run /bin/bash -c "dokku config:set $TEST_APP BUILDPACK_URL=null"
  echo "output: $output"
  echo "status: $status"
  assert_success

  chown -R dokku:dokku "$TMP"
  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP | tail -n1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "herokuish"
}
