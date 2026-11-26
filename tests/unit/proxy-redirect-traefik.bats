#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku traefik:start
  create_app
  dokku proxy:set "$TEST_APP" traefik
}

teardown() {
  dokku proxy:redirect:clear "$TEST_APP" >/dev/null 2>&1 || true
  destroy_app
  dokku traefik:stop
  dokku nginx:start
  global_teardown
}

@test "(proxy-redirect:traefik) redirect router replaces the source in the app router rule" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku ps:rebuild $TEST_APP"

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.routers.$TEST_APP-web-http.rule\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "Host(\`$TEST_APP.$DOKKU_DOMAIN\`)"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.routers.$TEST_APP-redirect-0-http.rule\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "Host(\`www.$TEST_APP.$DOKKU_DOMAIN\`)"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.routers.$TEST_APP-redirect-0-http.middlewares\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP-redirect-0-http"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.middlewares.$TEST_APP-redirect-0-http.redirectregex.replacement\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "http://$TEST_APP.$DOKKU_DOMAIN\${1}"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.middlewares.$TEST_APP-redirect-0-http.redirectregex.permanent\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/some/path?query=1" "301" "http://$TEST_APP.$DOKKU_DOMAIN/some/path?query=1"
  assert_http_localhost_response_contains "http" "$TEST_APP.$DOKKU_DOMAIN" "80" "/" "python/http.server"
}

@test "(proxy-redirect:traefik) temporary codes render a non-permanent redirect" {
  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN --code 307"
  assert_success
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.middlewares.$TEST_APP-redirect-0-http.redirectregex.permanent\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/" "302" "http://$TEST_APP.$DOKKU_DOMAIN/"
}
