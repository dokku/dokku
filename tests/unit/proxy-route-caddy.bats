#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku caddy:start
  create_app
  dokku proxy:set "$TEST_APP" caddy
}

teardown() {
  dokku proxy:route:clear "$TEST_APP" >/dev/null 2>&1 || true
  destroy_app
  dokku caddy:stop
  dokku nginx:start
  global_teardown
}

@test "(proxy-route:caddy) basic routing - /api/v0 hits api, / hits web" {
  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP" add_api_process_callback
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 api=1"
  assert_success

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Without --strip-prefix, /api/v0/Procfile reaches api as /api/v0/Procfile
  # which python's http.server returns 404 for. A 404 here proves the route
  # reaches api - web's catch-all would have returned 200.
  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "" "404"
}

@test "(proxy-route:caddy) --strip-prefix toggles upstream path visibility" {
  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP" add_api_process_callback
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 api=1"
  assert_success

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001"
  assert_success
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "" "404"

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001 --strip-prefix"
  assert_success
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "python3 -m http.server"
}

@test "(proxy-route:caddy) removing a route falls back to web" {
  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP" add_api_process_callback
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 api=1"
  assert_success

  # Set without --strip-prefix first; 404 proves the labels attached.
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001"
  assert_success
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "" "404"

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001 --strip-prefix"
  assert_success
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "python3 -m http.server"

  run /bin/bash -c "dokku proxy:route:remove $TEST_APP /api/v0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "(none)"

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  local body=""
  for attempt in $(seq 1 30); do
    run curl --connect-to "${TEST_APP}.${DOKKU_DOMAIN}:80:localhost:80" -kSs "http://${TEST_APP}.${DOKKU_DOMAIN}:80/api/v0/Procfile"
    body="$output"
    if ! grep -q 'python3 -m http.server' <<<"$body"; then
      break
    fi
    sleep 1
  done
  echo "final body attempt $attempt: $body"
  run /bin/bash -c "grep -c 'python3 -m http.server' <<<\"$body\" || true"
  assert_output "0"
}

add_api_process_callback() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  echo "api: python3 -m http.server 5001" >>"$APP_REPO_DIR/Procfile"
}
