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

@test "(git:report) info-flag works before deploy" {
  run /bin/bash -c "dokku git:report $TEST_APP --git-deploy-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku git:report $TEST_APP --git-computed-deploy-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "master"

  run /bin/bash -c "dokku git:set $TEST_APP deploy-branch main"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-deploy-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "main"

  run /bin/bash -c "dokku git:report $TEST_APP --git-computed-deploy-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "main"

  run /bin/bash -c "dokku git:set $TEST_APP deploy-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-deploy-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku git:report $TEST_APP --git-computed-deploy-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "master"

  run /bin/bash -c "dokku git:report $TEST_APP --git-sha"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-invalid-flag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid flag passed"
}

@test "(git:report) keep-git-dir raw vs computed vs global" {
  run /bin/bash -c "dokku git:set --global keep-git-dir"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-keep-git-dir"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku git:report $TEST_APP --git-global-keep-git-dir"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku git:report $TEST_APP --git-computed-keep-git-dir"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku git:set --global keep-git-dir true"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-global-keep-git-dir"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku git:report $TEST_APP --git-computed-keep-git-dir"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku git:set $TEST_APP keep-git-dir false"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-keep-git-dir"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku git:report $TEST_APP --git-global-keep-git-dir"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku git:report $TEST_APP --git-computed-keep-git-dir"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku git:set $TEST_APP keep-git-dir"
  assert_success

  run /bin/bash -c "dokku git:set --global keep-git-dir"
  assert_success
}

@test "(git:report) rev-env-var raw" {
  # Unset the per-app property file if a previous test left it written-empty
  # (git:set <app> rev-env-var "" intentionally writes an empty file as the
  # documented "unset" behavior, which masks the report's default fallback).
  run /bin/bash -c "dokku git:set $TEST_APP rev-env-var COMMIT_SHA"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-rev-env-var"
  assert_success
  assert_output "COMMIT_SHA"

  run /bin/bash -c "dokku git:set $TEST_APP rev-env-var"
  assert_success
}

@test "(git:report) source-image raw" {
  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku git:set $TEST_APP source-image alpine:3.20"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  assert_success
  assert_output "alpine:3.20"

  run /bin/bash -c "dokku git:set $TEST_APP source-image"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  assert_success
  assert_output_not_exists
}

@test "(git:report) raw and computed keys in --format json" {
  run /bin/bash -c "dokku git:report $TEST_APP --format json | jq -r '.\"deploy-branch\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku git:report $TEST_APP --format json | jq -r '.\"computed-deploy-branch\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "master"

  run /bin/bash -c "dokku git:report $TEST_APP --format json | jq -r '.\"keep-git-dir\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku git:report $TEST_APP --format json | jq -r '.\"computed-keep-git-dir\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"
}

@test "(git:report) --global raw and computed keys" {
  run /bin/bash -c "dokku git:report --global --format json | jq -r '.\"global-deploy-branch\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku git:report --global --format json | jq -r '.\"computed-deploy-branch\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "master"

  run /bin/bash -c "dokku git:report --global --format json | jq -r '.\"global-archive-max-files\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku git:report --global --format json | jq -r '.\"computed-archive-max-files\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "10000"

  run /bin/bash -c "dokku git:set --global deploy-branch main"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:report --global --format json | jq -r '.\"global-deploy-branch\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "main"

  run /bin/bash -c "dokku git:report --global --format json | jq -r '.\"computed-deploy-branch\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "main"

  run /bin/bash -c "dokku git:set --global deploy-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:report --global --format json | jq -r '.\"global-deploy-branch\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""
}

@test "(git) git:help" {
  run /bin/bash -c "dokku git"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage app deploys via git"
  help_output="$output"

  run /bin/bash -c "dokku git:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage app deploys via git"
  assert_output "$help_output"
}

@test "(git) ensure GIT_REV env var is set" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP GIT_REV"
  echo "output: $output"
  echo "status: $status"
  assert_output_exists
}

@test "(git) disable GIT_REV" {
  run /bin/bash -c "dokku git:set $TEST_APP rev-env-var"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP GIT_REV"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists
}

@test "(git) customize the GIT_REV environment variable" {
  run /bin/bash -c "dokku git:set $TEST_APP rev-env-var GIT_REV_ALT"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP GIT_REV_ALT"
  echo "output: $output"
  echo "status: $status"
  assert_output_exists
}

@test "(git) keep-git-dir" {
  run /bin/bash -c "dokku git:set $TEST_APP keep-git-dir true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku enter $TEST_APP web git status"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku enter $TEST_APP web ls .git"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "config"
  assert_output_contains "description"
  assert_output_contains "HEAD"
  assert_output_contains "hooks"
  assert_output_contains "index"
  assert_output_contains "info"
  assert_output_contains "logs"
  assert_output_contains "objects"
  assert_output_contains "refs"

  run /bin/bash -c "dokku enter $TEST_APP web test -d .git"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
