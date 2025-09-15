#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(builder) builder-detect [set]" {
  local TMP=$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM

  # test project.toml
  run touch "$TMP/project.toml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  chown -R dokku:dokku "$TMP"

  run /bin/bash -c "dokku builder:set --global selected pack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 0 "pack"

  run /bin/bash -c "dokku builder:set $TEST_APP selected pack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 0 "pack"

  run /bin/bash -c "dokku builder:set --global selected other"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 0 "pack"

  run /bin/bash -c "dokku builder:set $TEST_APP selected"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 0 "other"

  run /bin/bash -c "dokku builder:set --global selected"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 0 "pack"
}

@test "(builder) builder-detect [pack]" {
  local TMP=$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM

  # test project.toml
  run touch "$TMP/project.toml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  chown -R dokku:dokku "$TMP"
  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 0 "pack"

  sudo rm -rf $TMP/*
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(builder) builder-detect [dockerfile]" {
  local TMP=$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM

  run touch "$TMP/Dockerfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  chown -R dokku:dokku "$TMP"
  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 0 "dockerfile"
}

@test "(builder) builder-detect [herokuish]" {
  local TMP=$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM

  touch "$TMP/Dockerfile"

  # test buildpacks
  chown -R dokku:dokku "$TMP"
  run /bin/bash -c "touch $TMP/.buildpacks"
  echo "output: $output"
  echo "status: $status"
  assert_success

  chown -R dokku:dokku "$TMP"
  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 0 "herokuish"

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
  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 0 "herokuish"

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
  run /bin/bash -c "dokku plugin:trigger builder-detect $TEST_APP $TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 0 "herokuish"
}

@test "(builder:set)" {
  run deploy_app python
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP build-dir nonexistent-app"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku builder:set $TEST_APP build-dir sub-app"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'SECRET_KEYS:'

  run /bin/bash -c "dokku builder:set $TEST_APP build-dir"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'SECRET_KEY:'
}
