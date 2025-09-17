#!/usr/bin/env bats

load test_helper

setup_file() {
  global_setup
  install_railpack
  docker run --rm --privileged -d --name buildkit moby/buildkit:latest
  touch /etc/default/dokku
  sudo tee -a /etc/default/dokku <<<"export BUILDKIT_HOST='docker-container://buildkit'"
}

teardown_file() {
  docker container stop buildkit || true
  docker container rm buildkit || true
}

setup() {
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(builder-railpack:set)" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected railpack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP inject_requirements_txt
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'Successfully built image in'
  assert_output_contains "SECRET_KEY: fjdkslafjdk"

  run /bin/bash -c "dokku builder-railpack:set $TEST_APP railpackjson-path nonexistent.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'Successfully built image in'
}

inject_requirements_txt() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "flask" >>"$APP_REPO_DIR/requirements.txt"
}
