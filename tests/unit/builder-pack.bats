#!/usr/bin/env bats

load test_helper

setup_file() {
  install_pack
}

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(builder-pack:set)" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected pack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku buildpacks:set-property $TEST_APP stack heroku/builder:24"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_requirements_txt_cnb
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'Building with buildpack 1' 0
  assert_output_contains 'Installing dependencies using pip'

  run /bin/bash -c "dokku builder-pack:set $TEST_APP projecttoml-path nonexistent.toml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'Building with buildpack 1' 0
  assert_output_contains 'Installing dependencies using pip'

  run /bin/bash -c "dokku builder-pack:set $TEST_APP projecttoml-path project2.toml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'Building with buildpack 1'
  assert_output_contains 'Installing dependencies using pip' 0

  run /bin/bash -c "dokku builder-pack:set $TEST_APP projecttoml-path"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'Building with buildpack 1' 0
  assert_output_contains 'Installing dependencies using pip'
}

@test "(builder-pack) run" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected pack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_requirements_txt_cnb
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'from cnb stack'
  assert_output_contains 'Building with buildpack 1' 0
  assert_output_contains 'Installing dependencies using pip'

  run /bin/bash -c "dokku run $TEST_APP python task.py test"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "['task.py', 'test']"

  run /bin/bash -c "dokku run $TEST_APP task"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "['task.py', 'test']"

  run /bin/bash -c "dokku run $TEST_APP env"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "SECRET_KEY=fjdkslafjdk"
}

@test "(builder-pack) git:from-image without a Procfile" {
  run /bin/bash -c "dokku git:from-image $TEST_APP dokku/smoke-test-gradle-app:1"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(ps) cnb env vars" {
  run /bin/bash -c "dokku config:set $TEST_APP APP_RESPECTS_ENV_VARS=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected pack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_requirements_txt_cnb
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "curl $(dokku url $TEST_APP)/env"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains '"APP_RESPECTS_ENV_VARS": "1"'
}
