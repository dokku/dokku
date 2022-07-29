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

@test "(app-json) app.json dockerfile cmd" {
  run deploy_app dockerfile-procfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run docker inspect "dokku/${TEST_APP}:latest" --format "{{json .Config.Cmd}}"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output '["/bin/sh","-c","npm start"]'

  run docker inspect "dokku/${TEST_APP}:latest" --format "{{json .Config.Entrypoint}}"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output 'null'
}

@test "(app-json) app.json dockerfile release" {
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP SECRET_KEY=fjdkslafjdk ENVIRONMENT=dev DATABASE_URL=sqlite:///db.sqlite3"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app dockerfile-release
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Executing release task from Procfile"
  assert_output_contains "SECRET_KEY: fjdkslafjdk"
  assert_success
}

@test "(app-json) app.json dockerfile entrypoint release" {
  run deploy_app dockerfile-entrypoint dokku@dokku.me:$TEST_APP add_release_command
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "touch /app/release.test" 2
}

@test "(app-json) app.json dockerfile entrypoint predeploy" {
  run deploy_app dockerfile-entrypoint
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Executing predeploy task from app.json"
  assert_output_contains "entrypoint script started with arguments touch /app/predeploy.test"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP ls /app/predeploy.test"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
