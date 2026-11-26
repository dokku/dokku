#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku haproxy:set --global letsencrypt-server https://acme-staging-v02.api.letsencrypt.org/directory
  dokku haproxy:set --global letsencrypt-email
  dokku haproxy:set --global refresh-conf 2
  dokku haproxy:start
  create_app
  dokku proxy:set "$TEST_APP" haproxy
}

teardown() {
  dokku proxy:redirect:clear "$TEST_APP" >/dev/null 2>&1 || true
  global_teardown
  destroy_app
  dokku haproxy:stop
  dokku haproxy:set --global refresh-conf
  dokku nginx:start
}

@test "(proxy-redirect:haproxy) redirect label keeps the source host for certificates" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP" convert_to_dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN --code 302"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "always responds to redirects with a 301"
  assert_output_contains "dokku ps:rebuild $TEST_APP"

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"haproxy.$TEST_APP-web.redirect\" }}' | jq -r '.\"www.$TEST_APP.$DOKKU_DOMAIN\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "http://$TEST_APP.$DOKKU_DOMAIN"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"haproxy.$TEST_APP-web.host\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "www.$TEST_APP.$DOKKU_DOMAIN"

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/some/path?query=1" "301" "http://$TEST_APP.$DOKKU_DOMAIN/some/path?query=1"
  assert_http_localhost_response_contains "http" "$TEST_APP.$DOKKU_DOMAIN" "80" "/" "python/http.server"
}
