#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  LOCK_DIR="$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")"
  LOCK_FILE="$LOCK_DIR/test.lock"
  LOCK_PIDS=()
}

teardown() {
  local pid
  touch "$LOCK_DIR/release"
  for pid in "${LOCK_PIDS[@]}"; do
    kill "$pid" &>/dev/null || true
  done
  rm -rf "$LOCK_DIR"
  global_teardown
}

lock_script() {
  declare BUILD_ID="$1" SCRIPT="$2"
  echo "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; export DOKKU_BUILD_ID='$BUILD_ID'; $SCRIPT"
}

start_lock_holder() {
  declare BUILD_ID="$1"
  /bin/bash -c "$(lock_script "$BUILD_ID" "acquire_advisory_lock '$LOCK_FILE' exclusive 'waiting' 'lock held'; while [[ ! -f '$LOCK_DIR/release' ]]; do sleep 0.2; done; release_advisory_lock '$LOCK_FILE'")" >/dev/null 2>&1 3>&- &
  LOCK_PIDS+=("$!")
  wait_for "[[ \"\$(cat '$LOCK_FILE' 2>/dev/null)\" == '$BUILD_ID' ]]"
}

wait_for() {
  declare CONDITION="$1"
  local i
  for i in $(seq 1 50); do
    if eval "$CONDITION"; then
      return 0
    fi
    sleep 0.2
  done

  echo "timed out waiting for: $CONDITION"
  return 1
}

@test "(deploy-lock) a failed acquire does not truncate the lock holder's build id" {
  start_lock_holder "holder-a"

  run /bin/bash -c "$(lock_script "other" "acquire_advisory_lock '$LOCK_FILE' exclusive 'waiting' 'lock held'")"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "lock held"

  run /bin/bash -c "cat '$LOCK_FILE'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "holder-a"
}

@test "(deploy-lock) a waiter does not keep a stale lock after the holder removes the lock file" {
  start_lock_holder "holder-a"

  /bin/bash -c "$(lock_script "holder-b" "acquire_advisory_lock '$LOCK_FILE' waiting 'waiting' 'lock held'; touch '$LOCK_DIR/b.acquired'; exec sleep 60")" >"$LOCK_DIR/b.log" 2>&1 3>&- &
  local B_PID="$!"
  LOCK_PIDS+=("$B_PID")
  wait_for "grep -q waiting '$LOCK_DIR/b.log'"

  touch "$LOCK_DIR/release"
  wait_for "[[ -f '$LOCK_DIR/b.acquired' ]]"

  run /bin/bash -c "[[ /proc/$B_PID/fd/200 -ef '$LOCK_FILE' ]]"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat '$LOCK_FILE'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "holder-b"

  run /bin/bash -c "$(lock_script "holder-c" "acquire_advisory_lock '$LOCK_FILE' exclusive 'waiting' 'lock held'")"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "lock held"
}

@test "(deploy-lock) the lock file is removed when the holder exits without releasing it" {
  run /bin/bash -c "$(lock_script "holder-a" "acquire_advisory_lock '$LOCK_FILE' exclusive 'waiting' 'lock held'; exit 1")"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "[[ ! -e '$LOCK_FILE' ]]"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(deploy-lock) the lock file is removed when the holder is terminated" {
  start_lock_holder "holder-a"

  kill -TERM "${LOCK_PIDS[0]}"
  wait_for "! kill -0 ${LOCK_PIDS[0]} 2>/dev/null"

  run /bin/bash -c "[[ ! -e '$LOCK_FILE' ]]"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(deploy-lock) release only removes the lock file it holds" {
  start_lock_holder "holder-a"

  run /bin/bash -c "rm -f '$LOCK_FILE' && echo 'replacement' >'$LOCK_FILE'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  touch "$LOCK_DIR/release"
  wait_for "! kill -0 ${LOCK_PIDS[0]} 2>/dev/null"

  run /bin/bash -c "cat '$LOCK_FILE'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "replacement"
}
