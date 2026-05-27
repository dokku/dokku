#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  dokku builder:set --global skip-cleanup >/dev/null 2>/dev/null || true
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

@test "(builder:report) --global --format json" {
  run /bin/bash -c "dokku builder:set --global selected dockerfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:report --global --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:report --global --format json | jq -r '.\"builder-global-selected\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dockerfile"

  run /bin/bash -c "dokku builder:report --global --format json | jq -r '.\"global-selected\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dockerfile"

  run /bin/bash -c "dokku builder:report --global --format json | jq -r 'has(\"builder-selected\")'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku builder:report --global --format json | jq -r 'has(\"selected\") and has(\"builder-global-selected\")'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku builder:report --global --format json | jq -r 'has(\"computed-skip-cleanup\") and has(\"builder-computed-skip-cleanup\")'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku builder:report --global"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set --global selected"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(builder:report) selected raw vs computed vs global" {
  run /bin/bash -c "dokku builder:set --global selected"
  assert_success

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-selected"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-global-selected"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-computed-selected"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku builder:set --global selected dockerfile"
  assert_success

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-global-selected"
  assert_success
  assert_output "dockerfile"

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-computed-selected"
  assert_success
  assert_output "dockerfile"

  run /bin/bash -c "dokku builder:set $TEST_APP selected herokuish"
  assert_success

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-selected"
  assert_success
  assert_output "herokuish"

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-global-selected"
  assert_success
  assert_output "dockerfile"

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-computed-selected"
  assert_success
  assert_output "herokuish"

  run /bin/bash -c "dokku builder:set $TEST_APP selected"
  assert_success

  run /bin/bash -c "dokku builder:set --global selected"
  assert_success
}

@test "(builder:report) build-dir raw vs computed vs global" {
  run /bin/bash -c "dokku builder:set --global build-dir"
  assert_success

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-build-dir"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-global-build-dir"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-computed-build-dir"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku builder:set --global build-dir global/subdir"
  assert_success

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-global-build-dir"
  assert_success
  assert_output "global/subdir"

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-computed-build-dir"
  assert_success
  assert_output "global/subdir"

  run /bin/bash -c "dokku builder:set $TEST_APP build-dir app/subdir"
  assert_success

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-build-dir"
  assert_success
  assert_output "app/subdir"

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-global-build-dir"
  assert_success
  assert_output "global/subdir"

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-computed-build-dir"
  assert_success
  assert_output "app/subdir"

  run /bin/bash -c "dokku builder:set $TEST_APP build-dir"
  assert_success

  run /bin/bash -c "dokku builder:set --global build-dir"
  assert_success
}

@test "(builder:report) skip-cleanup raw vs computed vs global" {
  run /bin/bash -c "dokku builder:set --global skip-cleanup"
  assert_success

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-skip-cleanup"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-global-skip-cleanup"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-computed-skip-cleanup"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku builder:set --global skip-cleanup true"
  assert_success

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-global-skip-cleanup"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-computed-skip-cleanup"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku builder:set $TEST_APP skip-cleanup false"
  assert_success

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-skip-cleanup"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-global-skip-cleanup"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku builder:report $TEST_APP --builder-computed-skip-cleanup"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku builder:set $TEST_APP skip-cleanup"
  assert_success

  run /bin/bash -c "dokku builder:set --global skip-cleanup"
  assert_success
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
