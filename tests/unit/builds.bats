#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  rm -rf "$DOKKU_LIB_ROOT/data/builds/$TEST_APP" || true
  destroy_app
  global_teardown
}

write_finished_record() {
  local app="$1" id="$2" status="$3" source="${4:-git-hook}" pid="${5:-99999}" kind="${6:-build}" exit_code="${7:-0}"
  local dir="$DOKKU_LIB_ROOT/data/builds/$app"
  mkdir -p "$dir"
  local now
  now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  cat >"$dir/$id.json" <<EOF
{
  "id": "$id",
  "app": "$app",
  "kind": "$kind",
  "pid": $pid,
  "started_at": "$now",
  "finished_at": "$now",
  "status": "$status",
  "source": "$source",
  "exit_code": $exit_code
}
EOF
  : >"$dir/$id.log"
  chown -R dokku:dokku "$DOKKU_LIB_ROOT/data/builds"
}

write_running_record() {
  local app="$1" id="$2" pid="$3" source="${4:-git-hook}" kind="${5:-build}"
  local dir="$DOKKU_LIB_ROOT/data/builds/$app"
  mkdir -p "$dir"
  local now
  now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  cat >"$dir/$id.json" <<EOF
{
  "id": "$id",
  "app": "$app",
  "kind": "$kind",
  "pid": $pid,
  "started_at": "$now",
  "status": "running",
  "source": "$source"
}
EOF
  : >"$dir/$id.log"
  chown -R dokku:dokku "$DOKKU_LIB_ROOT/data/builds"
}

@test "(builds) builds:help" {
  run /bin/bash -c "dokku builds"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage running and historical builds"

  run /bin/bash -c "dokku builds:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage running and historical builds"
}

@test "(builds:list) returns empty list when no builds have run" {
  run /bin/bash -c "dokku builds:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "No builds recorded"
}

@test "(builds:list) shows a finished record after a successful deploy" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builds:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "succeeded"
  assert_output_contains "git-hook"
}

@test "(builds:list) shows an abandoned record when the PID is gone but the record is still running" {
  write_running_record "$TEST_APP" "ghost001" 99999 "git-hook"

  run /bin/bash -c "dokku builds:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "abandoned"
}

@test "(builds:list --kind=build) filters to build-kind records" {
  write_finished_record "$TEST_APP" "b1" "succeeded" "git-hook" 1001 "build"
  write_finished_record "$TEST_APP" "d1" "succeeded" "ps:restart" 1002 "deploy"

  run /bin/bash -c "dokku builds:list $TEST_APP --kind build"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "b1"
  assert_output_not_contains "d1"
}

@test "(builds:list --kind=deploy) filters to deploy-kind records" {
  write_finished_record "$TEST_APP" "b1" "succeeded" "git-hook" 1001 "build"
  write_finished_record "$TEST_APP" "d1" "succeeded" "ps:restart" 1002 "deploy"

  run /bin/bash -c "dokku builds:list $TEST_APP --kind deploy"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "d1"
  assert_output_not_contains "b1"
}

@test "(builds:list --kind=invalid) fails with a usage error" {
  run /bin/bash -c "dokku builds:list $TEST_APP --kind=garbage"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid --kind"
}

@test "(builds:set) writes a per-app retention" {
  run /bin/bash -c "dokku builds:set $TEST_APP retention 5"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builds:report $TEST_APP --builds-retention"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "5"
}

@test "(builds:set --global) writes the global retention" {
  run /bin/bash -c "dokku builds:set --global retention 7"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builds:report $TEST_APP --builds-computed-retention"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "7"

  dokku builds:set --global retention "" || true
}

@test "(builds:set) clears the override when called with no value" {
  dokku builds:set "$TEST_APP" retention 5

  run /bin/bash -c "dokku builds:set $TEST_APP retention"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builds:report $TEST_APP --builds-retention"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""
}

@test "(builds:set) rejects non-positive-integer values" {
  run /bin/bash -c "dokku builds:set $TEST_APP retention 0"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku builds:set $TEST_APP retention abc"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(builds:info) returns non-zero for a missing build-id" {
  run /bin/bash -c "dokku builds:info $TEST_APP nonexistent"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(builds:info) returns the recorded fields including --format json" {
  write_finished_record "$TEST_APP" "info001" "succeeded" "git-hook" 1234 "build"

  run /bin/bash -c "dokku builds:info $TEST_APP info001"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "info001" -1
  assert_output_contains "succeeded"
  assert_output_contains "Log:"

  run /bin/bash -c "dokku builds:info $TEST_APP info001 --format json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "\"id\":\"info001\""
  assert_output_contains "\"log_path\":"
}

@test "(builds:cancel) returns 0 with a friendly message when no lock is present" {
  run /bin/bash -c "dokku builds:cancel $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "App not currently deploying"
}

@test "(builds:cancel) refuses to cancel a record whose status is not running" {
  write_finished_record "$TEST_APP" "done01" "succeeded" "git-hook" 1234 "build"
  mkdir -p "$DOKKU_LIB_ROOT/data/apps/$TEST_APP"
  echo "done01" >"$DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock"
  chown dokku:dokku "$DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock"

  run /bin/bash -c "dokku builds:cancel $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "no longer running"

  rm -f "$DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock"
}

@test "(builds:cancel) marks an abandoned record as failed instead of canceled" {
  write_running_record "$TEST_APP" "abnd01" 99999 "git-hook"
  mkdir -p "$DOKKU_LIB_ROOT/data/apps/$TEST_APP"
  echo "abnd01" >"$DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock"
  chown dokku:dokku "$DOKKU_LIB_ROOT/data/apps/$TEST_APP/.deploy.lock"

  run /bin/bash -c "dokku builds:cancel $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "already terminated"

  run /bin/bash -c "dokku builds:info $TEST_APP abnd01"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "failed"
}

@test "(builds:output) cats the log file for a finished build" {
  write_finished_record "$TEST_APP" "out001" "succeeded" "git-hook" 1234 "build"
  echo "build log line" >"$DOKKU_LIB_ROOT/data/builds/$TEST_APP/out001.log"
  chown dokku:dokku "$DOKKU_LIB_ROOT/data/builds/$TEST_APP/out001.log"

  run /bin/bash -c "dokku builds:output $TEST_APP out001"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "build log line"
}

@test "(builds:prune) prunes an app's records to retention" {
  dokku builds:set "$TEST_APP" retention 2

  for i in 1 2 3 4 5; do
    write_finished_record "$TEST_APP" "p$i" "succeeded" "git-hook" $((1000 + i)) "build"
    sleep 1
  done

  run /bin/bash -c "dokku builds:prune $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  count="$(ls $DOKKU_LIB_ROOT/data/builds/$TEST_APP/*.json 2>/dev/null | wc -l)"
  echo "remaining: $count"
  [[ "$count" -le 2 ]] || (echo "expected <=2 records, got $count" && false)
}

@test "(builds:prune) reaps an abandoned record and finalizes it as failed" {
  write_running_record "$TEST_APP" "ghost02" 99999 "git-hook"

  run /bin/bash -c "dokku builds:prune $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builds:info $TEST_APP ghost02"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "failed"
}

@test "(builds:prune) skips live-running records" {
  local pid=$$
  write_running_record "$TEST_APP" "alive02" "$pid" "git-hook"

  run /bin/bash -c "dokku builds:prune $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builds:info $TEST_APP alive02"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "running"
}

@test "(builds) [storage] directory is removed on apps:destroy" {
  write_finished_record "$TEST_APP" "del01" "succeeded" "git-hook" 1234 "build"
  [[ -d "$DOKKU_LIB_ROOT/data/builds/$TEST_APP" ]]

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  [[ ! -d "$DOKKU_LIB_ROOT/data/builds/$TEST_APP" ]]

  # Recreate so teardown can run cleanly.
  create_app
}
