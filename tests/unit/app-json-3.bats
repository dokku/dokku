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

@test "(app-json) persist scale" {
  local CUSTOM_TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM

  CUSTOM_TMP="$CUSTOM_TMP" run deploy_app python dokku@dokku.me:$TEST_APP persist_scale_callback_a
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ps:scale $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "web:  1"
  assert_output_contains "worker: 1"
  assert_output_contains "cron: 0" 0

  run persist_scale_callback_b "$TEST_APP" "$CUSTOM_TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$CUSTOM_TMP" add app.json Procfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$CUSTOM_TMP" commit -m 'Update scaling parameters'
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$CUSTOM_TMP" push target master:master
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ps:scale $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "web:  1"
  assert_output_contains "worker: 1" 0
  assert_output_contains "worker: 0" 0
  assert_output_contains "cron: 1" 1
}

persist_scale_callback_a() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"

  rm "$APP_REPO_DIR/Procfile"
  touch "$APP_REPO_DIR/Procfile"
  echo "web: python3 web.py" >>"$APP_REPO_DIR/Procfile"
  echo "worker: python3 worker.py" >>"$APP_REPO_DIR/Procfile"
  mv "$APP_REPO_DIR/app-5205a.json" "$APP_REPO_DIR/app.json"
}

persist_scale_callback_b() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"

  rm "$APP_REPO_DIR/Procfile"
  touch "$APP_REPO_DIR/Procfile"
  echo "web: python3 web.py" >>"$APP_REPO_DIR/Procfile"
  echo "cron: python3 worker.py" >>"$APP_REPO_DIR/Procfile"
  mv "$APP_REPO_DIR/app-5205b.json" "$APP_REPO_DIR/app.json"
}
