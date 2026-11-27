#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku openresty:start
  create_app
  dokku proxy:set "$TEST_APP" openresty
}

teardown() {
  dokku proxy:redirect:clear "$TEST_APP" >/dev/null 2>&1 || true
  destroy_app
  dokku openresty:stop
  dokku nginx:start
  global_teardown
}

@test "(proxy-redirect:openresty) redirect labels replace the source in the app domains" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN --code 308"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku ps:rebuild $TEST_APP"

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"openresty.domains\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP.$DOKKU_DOMAIN"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"openresty.redirect.0.source\" }} {{ index .Config.Labels \"openresty.redirect.0.destination\" }} {{ index .Config.Labels \"openresty.redirect.0.code\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN 308"

  sleep 5

  run /bin/bash -c "docker exec openresty-openresty-1 cat /etc/nginx/sites-enabled/sites.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "server_name                 www.$TEST_APP.$DOKKU_DOMAIN;"
  assert_output_contains "return 308 \$scheme://$TEST_APP.$DOKKU_DOMAIN\$request_uri;"

  run /bin/bash -c "docker exec openresty-openresty-1 /usr/local/openresty/nginx/sbin/nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "conflicting server name"

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/some/path?query=1" "308" "http://$TEST_APP.$DOKKU_DOMAIN/some/path?query=1"
  assert_http_localhost_response_contains "http" "$TEST_APP.$DOKKU_DOMAIN" "80" "/" "python/http.server"
}
