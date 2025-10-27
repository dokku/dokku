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

@test "(app-json) app.json scripts" {
  run deploy_app python
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Executing predeploy task from app.json: touch /app/predeploy.test"
  assert_output_contains "Executing postdeploy task from app.json in ephemeral container: touch /app/postdeploy.test"
  assert_output_contains "Executing prebuild task from app.json in ephemeral container: touch /app/prebuild.test" 0

  run docker inspect "${TEST_APP}.web.1" --format "{{json .Config.Cmd}}"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output '["/start","web"]'

  run /bin/bash -c "dokku run $TEST_APP ls /app/prebuild.test"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku run $TEST_APP ls /app/predeploy.test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP ls /app/postdeploy.test"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(app-json) app.json scripts postdeploy" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_postdeploy_command
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "touch /app/heroku-postdeploy.test"
  assert_output_contains "python3 -u release.py"
}

@test "(app-json) app.json scripts missing" {
  run deploy_app nodejs-express-noappjson
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(app-json) app.json herokuish release" {
  run /bin/bash -c "dokku config:set --no-restart --global GLOBAL_SECRET=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_requirements_txt
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Executing release task from Procfile"
  assert_output_contains "SECRET_KEY: fjdkslafjdk"

  run /bin/bash -c "curl $(dokku url $TEST_APP)/env"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains '"GLOBAL_SECRET": "fjdkslafjdk"'
  assert_output_contains '"SECRET_KEY": "fjdkslafjdk"'
}

@test "(app-json) app.json cnb release" {
  run /bin/bash -c "dokku config:set --no-restart --global GLOBAL_SECRET=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP SECRET_KEY=fjdkslafjdk"
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
  assert_output_contains "Executing release task from Procfile"
  assert_output_contains "SECRET_KEY: fjdkslafjdk"

  run /bin/bash -c "curl $(dokku url $TEST_APP)/env"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains '"GLOBAL_SECRET": "fjdkslafjdk"'
  assert_output_contains '"SECRET_KEY": "fjdkslafjdk"'
}

@test "(app-json) tini test" {
  if ! dokku plugin:installed postgres; then
    run /bin/bash -c "dokku plugin:install https://github.com/dokku/dokku-postgres.git"
    echo "output: $output"
    echo "status: $status"
    assert_success
  fi

  if ! dokku plugin:installed redis; then
    run /bin/bash -c "dokku plugin:install https://github.com/dokku/dokku-redis.git"
    echo "output: $output"
    echo "status: $status"
    assert_success
  fi

  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY_BASE=derp OTP_SECRET=1234"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku postgres:create $TEST_APP $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku postgres:link $TEST_APP $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku redis:create $TEST_APP $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku redis:link $TEST_APP $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP tootsuite/mastodon:v3.3.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku postgres:unlink $TEST_APP $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku redis:unlink $TEST_APP $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force postgres:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force redis:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(app-json:set)" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP copy_app_json_to_sub_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path app-nonexistent.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "touch /app/predeploy.test" 0

  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path app2.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "touch /app/predeploy2.test"

  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "touch /app/predeploy.test"
  assert_output_contains "touch /app/predeploy2.test" 0

  run /bin/bash -c "dokku builder:set $TEST_APP build-dir sub-app"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "touch /app/predeploy.test" 0
  assert_output_contains "touch /app/predeploy2.test"

  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path app3.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "touch /app/predeploy.test" 0
  assert_output_contains "touch /app/predeploy2.test" 0
  assert_output_contains "touch /app/predeploy3.test"
}

copy_app_json_to_sub_app() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  cp "$APP_REPO_DIR/app2.json" "$APP_REPO_DIR/sub-app/app.json"
  cp "$APP_REPO_DIR/app3.json" "$APP_REPO_DIR/sub-app/app3.json"
}
