#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  dokku proxy:redirect:clear "$TEST_APP" >/dev/null 2>&1 || true
  destroy_app
  global_teardown
}

@test "(proxy-redirect:nginx) redirects a source domain to the destination and keeps the request uri" {
  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/some/path?query=1" "301" "http://$TEST_APP.$DOKKU_DOMAIN/some/path?query=1"
  assert_http_localhost_response_contains "http" "$TEST_APP.$DOKKU_DOMAIN" "80" "/" "python/http.server"

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN --code 302"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/" "302" "http://$TEST_APP.$DOKKU_DOMAIN/"
}

@test "(proxy-redirect:nginx) a source that is also an app domain is only served by the redirect" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response_contains "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/" "python/http.server"

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "conflicting server name"

  run /bin/bash -c "nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "conflicting server name"

  run /bin/bash -c "grep -c 'server_name .*www.$TEST_APP.$DOKKU_DOMAIN' $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "grep -cxF '  server_name www.$TEST_APP.$DOKKU_DOMAIN;' $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-vhosts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "www.$TEST_APP.$DOKKU_DOMAIN"

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/" "301" "http://$TEST_APP.$DOKKU_DOMAIN/"

  run /bin/bash -c "dokku proxy:redirect:remove $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response_contains "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/" "python/http.server"
}

@test "(proxy-redirect:nginx) redirects render for each http listener and survive a rebuild" {
  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ports:set $TEST_APP http:80:5000 http:8080:5000"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/" "301" "http://$TEST_APP.$DOKKU_DOMAIN/"
  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "8080" "/" "301" "http://$TEST_APP.$DOKKU_DOMAIN/"
}
