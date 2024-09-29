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

@test "(app-json) app.json failing dokku.predeploy" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_failing_dokku_predeploy
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Execution of predeploy task failed"
  assert_failure
}

@test "(app-json) app.json failing dokku.postdeploy" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_failing_dokku_postdeploy
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Execution of postdeploy task failed"
  assert_failure
}

@test "(app-json) app.json failing postdeploy" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_failing_postdeploy
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Execution of first deploy postdeploy task failed"
  assert_failure
}

@test "(app-json) app.json failing release" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_failing_release
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Execution of release task failed"
  assert_failure
}

add_failing_dokku_predeploy() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"

  cat >"$APP_REPO_DIR/app.json" <<EOF
  {
    "scripts": {
      "dokku": {
        "predeploy": "exit 1"
      }
    }
  }
EOF
}

add_failing_dokku_postdeploy() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"

  cat >"$APP_REPO_DIR/app.json" <<EOF
  {
    "scripts": {
      "dokku": {
        "postdeploy": "exit 1"
      }
    }
  }
EOF
}

add_failing_postdeploy() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"

  cat >"$APP_REPO_DIR/app.json" <<EOF
  {
    "scripts": {
      "postdeploy": "exit 1"
    }
  }
EOF
}

add_failing_release() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"

  cat >"$APP_REPO_DIR/Procfile" <<EOF
  web: python3 -u web.py
  release: exit 1
EOF
}
