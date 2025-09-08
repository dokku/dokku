#!/usr/bin/env bats

load test_helper

setup_file() {
  install_nixpacks
}

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(builder-nixpacks:set)" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected nixpacks"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP inject_requirements_txt
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'load build definition from Dockerfile'
  assert_output_contains "SECRET_KEY: fjdkslafjdk"

  run /bin/bash -c "dokku builder-nixpacks:set $TEST_APP nixpackstoml-path nonexistent.toml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'load build definition from Dockerfile'
}

inject_requirements_txt() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "flask" >>"$APP_REPO_DIR/requirements.txt"
}
