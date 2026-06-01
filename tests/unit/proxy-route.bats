#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  dokku proxy:route:clear "$TEST_APP" >/dev/null 2>&1 || true
  destroy_app
  global_teardown
}

@test "(proxy:route) help text mentions route commands" {
  run /bin/bash -c "dokku proxy:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "proxy:route:set"
  assert_output_contains "proxy:route:remove"
  assert_output_contains "proxy:route:clear"
  assert_output_contains "proxy:route:report"
}

@test "(proxy:route:set) stores a route under the default port and is idempotent" {
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "/api/v0 -> api:5000"

  # second call with identical args produces identical state
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "/api/v0 -> api:5000"
}

@test "(proxy:route:set) --port overrides the default port" {
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "/api/v0 -> api:5001"
}

@test "(proxy:route:set) --strip-prefix marks the route as stripping" {
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --strip-prefix"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "/api/v0 -> api:5000 (strip)"
}

@test "(proxy:route:set) re-setting without --port resets port to default" {
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "/api/v0 -> api:5000"
}

@test "(proxy:route:set) re-setting without --strip-prefix resets strip to false" {
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --strip-prefix"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_contains "(strip)"
}

@test "(proxy:route:set) refuses web as a target" {
  run /bin/bash -c "dokku proxy:route:set $TEST_APP web /api/v0"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "web cannot be a route target"
}

@test "(proxy:route:set) refuses root path" {
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "reserved for the web process"
}

@test "(proxy:route:set) refuses path without leading slash" {
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api api/v0"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "must start with /"
}

@test "(proxy:route:set) refuses path with trailing slash" {
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0/"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "must not end with /"
}

@test "(proxy:route:set) refuses ports out of range" {
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 70000"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "out of range"
}

@test "(proxy:route:set) warns when target process is scaled to zero" {
  run /bin/bash -c "dokku ps:scale $TEST_APP api=0"
  echo "output: $output"
  echo "status: $status"

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "scaled to 0"
}

@test "(proxy:route:remove) deletes a route" {
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0"
  assert_success

  run /bin/bash -c "dokku proxy:route:remove $TEST_APP /api/v0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_contains "/api/v0"
}

@test "(proxy:route:remove) is idempotent on a nonexistent path" {
  run /bin/bash -c "dokku proxy:route:remove $TEST_APP /does/not/exist"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(proxy:route:clear) removes every route" {
  dokku proxy:route:set "$TEST_APP" api /api/v0 >/dev/null
  dokku proxy:route:set "$TEST_APP" api /api/v0/admin >/dev/null
  dokku proxy:route:set "$TEST_APP" ws /ws --port 8080 >/dev/null

  run /bin/bash -c "dokku proxy:route:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "(none)"
}

@test "(proxy:route:clear) is idempotent on an empty route set" {
  run /bin/bash -c "dokku proxy:route:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(proxy:route:report) supports json output" {
  dokku proxy:route:set "$TEST_APP" api /api/v0 --port 5001 >/dev/null
  dokku proxy:route:set "$TEST_APP" ws /ws --port 8080 --strip-prefix >/dev/null

  run /bin/bash -c "dokku proxy:route:report $TEST_APP --format json | jq -e '.routes | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  run /bin/bash -c "dokku proxy:route:report $TEST_APP --format json | jq -r '.routes[] | select(.path == \"/ws\") | .strip_prefix'"
  echo "output: $output"
  echo "status: $status"
  assert_output "true"
}

@test "(proxy:route:set) refuses under openresty backend" {
  dokku proxy:set "$TEST_APP" openresty >/dev/null 2>&1 || skip "openresty proxy not available"

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "not yet supported"

  run /bin/bash -c "dokku proxy:route:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "(none)"
}

@test "(proxy:route:set) refuses under haproxy backend" {
  dokku proxy:set "$TEST_APP" haproxy >/dev/null 2>&1 || skip "haproxy proxy not available"

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "not yet supported"

  run /bin/bash -c "dokku proxy:route:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "(none)"
}
