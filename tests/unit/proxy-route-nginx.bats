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

@test "(proxy-route:nginx) basic routing - /api/v0 hits api, / hits web" {
  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP" add_api_process_callback
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 api=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Without --strip-prefix, /api/v0/Procfile reaches api as /api/v0/Procfile.
  # api (python's http.server) does not have that file on disk and returns
  # 404. A 404 here proves the route reaches api - web's catch-all would
  # have returned 200 instead.
  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "" "404"
}

@test "(proxy-route:nginx) --strip-prefix toggles upstream path visibility" {
  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP" add_api_process_callback
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 api=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Without --strip-prefix, /api/v0/Procfile is forwarded as-is to api,
  # which does not have an /api/v0/Procfile file on disk.
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001"
  assert_success
  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "" "404"

  # With --strip-prefix, /api/v0/Procfile becomes /Procfile upstream and api
  # returns the Procfile content.
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001 --strip-prefix"
  assert_success
  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "python3 -m http.server"
}

@test "(proxy-route:nginx) removing a route falls back to web" {
  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP" add_api_process_callback
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 api=1"
  assert_success

  # First set the route without --strip-prefix so a 404 from api uniquely
  # identifies that the route reached api (avoids racing with web's catch-all
  # 200 during the brief window before nginx fully reloads).
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001"
  assert_success
  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "" "404"

  # Now switch the route to --strip-prefix and assert api serves the Procfile.
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001 --strip-prefix"
  assert_success
  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "python3 -m http.server"

  run /bin/bash -c "dokku proxy:route:remove $TEST_APP /api/v0"
  assert_success

  # After removal, /api/v0/Procfile no longer reaches api - it falls through
  # to web's catch-all. Assert the response is no longer api's Procfile body.
  run curl --connect-to "${TEST_APP}.${DOKKU_DOMAIN}:80:localhost:80" -kSso /tmp/route-removed-body "http://${TEST_APP}.${DOKKU_DOMAIN}:80/api/v0/Procfile"
  run /bin/bash -c "grep -c 'python3 -m http.server' /tmp/route-removed-body || true"
  assert_output "0"
}

@test "(proxy-route:nginx) websocket upgrade headers are present in rendered config" {
  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP" add_api_process_callback
  assert_success
  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 api=1"
  assert_success
  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "location /api/v0"
  # These headers appear in both the route's location block AND the existing
  # location / block, so allow >= 1 occurrence (count = -1) rather than the
  # default exact-count-of-1.
  assert_output_contains "proxy_http_version 1.1" -1
  assert_output_contains "proxy_set_header Upgrade \$http_upgrade" -1
  assert_output_contains "proxy_set_header Connection \$http_connection" -1
}

add_api_process_callback() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  # api serves /app's working directory on port 5001. GET /Procfile returns
  # the literal Procfile bytes, which web.py does not serve - this gives the
  # HTTP assertions a clear marker for which process handled the request.
  echo "api: python3 -m http.server 5001" >>"$APP_REPO_DIR/Procfile"
}
