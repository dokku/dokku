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

@test "(app-json) app.json env simple value" {
  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path app-env.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Processing app.json env vars"
  assert_output_contains "Setting 4 env var(s) from app.json"

  run /bin/bash -c "dokku config:get $TEST_APP SIMPLE_VAR"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "simple_value"

  run /bin/bash -c "dokku config:get $TEST_APP OBJECT_VAR"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "object_default"
}

@test "(app-json) app.json env secret generator" {
  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path app-env.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP SECRET_VAR"
  echo "output: $output"
  echo "status: $status"
  assert_success
  # Secret should be 64 characters (hex encoded 32 bytes)
  [[ ${#output} -eq 64 ]]
}

@test "(app-json) app.json env does not overwrite on redeploy" {
  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path app-env.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Get the original secret
  run /bin/bash -c "dokku config:get $TEST_APP SECRET_VAR"
  echo "output: $output"
  echo "status: $status"
  assert_success
  local original_secret="$output"

  # Change SIMPLE_VAR manually
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP SIMPLE_VAR=changed_value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Rebuild the app
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Verify SIMPLE_VAR was NOT overwritten
  run /bin/bash -c "dokku config:get $TEST_APP SIMPLE_VAR"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "changed_value"

  # Verify SECRET_VAR was NOT regenerated
  run /bin/bash -c "dokku config:get $TEST_APP SECRET_VAR"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$original_secret"
}

@test "(app-json) app.json env sync overwrites on redeploy" {
  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path app-env.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Verify SYNC_VAR is set
  run /bin/bash -c "dokku config:get $TEST_APP SYNC_VAR"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "sync_value"

  # Change SYNC_VAR manually
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP SYNC_VAR=manual_value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Rebuild the app
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Verify SYNC_VAR WAS overwritten back to sync_value
  run /bin/bash -c "dokku config:get $TEST_APP SYNC_VAR"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "sync_value"
}

@test "(app-json) app.json env required without value fails" {
  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path app-env-required.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "required env var REQUIRED_VAR has no value"
}

@test "(app-json) app.json env skips optional without value" {
  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path app-env.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  # OPTIONAL_VAR should not be set since it has no default and is optional
  # config:get returns exit code 1 when a variable is not set
  run /bin/bash -c "dokku config:get $TEST_APP OPTIONAL_VAR"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output ""
}

@test "(app-json) app.json env respects pre-set values on first deploy" {
  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path app-env.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Pre-set SIMPLE_VAR before first deploy
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP SIMPLE_VAR=preset_value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Verify SIMPLE_VAR was NOT overwritten
  run /bin/bash -c "dokku config:get $TEST_APP SIMPLE_VAR"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "preset_value"
}
