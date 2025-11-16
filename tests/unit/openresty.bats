#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku openresty:set --global letsencrypt-server https://acme-staging-v02.api.letsencrypt.org/directory
  dokku openresty:set --global letsencrypt-email
  dokku openresty:start
  create_app
}

teardown() {
  global_teardown
  destroy_app
  dokku openresty:stop
  dokku nginx:start
}

@test "(openresty) openresty:help" {
  run /bin/bash -c "dokku openresty"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the openresty proxy integration"
  help_output="$output"

  run /bin/bash -c "dokku openresty:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the openresty proxy integration"
  assert_output "$help_output"
}

@test "(openresty) single domain" {
  run /bin/bash -c "dokku proxy:set $TEST_APP openresty"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP convert_to_dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response "http" "$TEST_APP.dokku.me" "80" "" "python/http.server"
}

@test "(openresty) multiple domains" {
  run /bin/bash -c "dokku proxy:set $TEST_APP openresty"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP $TEST_APP.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP $TEST_APP-2.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP convert_to_dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response "http" "$TEST_APP.dokku.me" "80" "" "python/http.server"
  assert_http_localhost_response "http" "$TEST_APP-2.dokku.me" "80" "" "python/http.server"
}

@test "(openresty) ssl" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:set $TEST_APP openresty"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_localhost_response "http" "$TEST_APP.dokku.me" "80" "" "python/http.server"

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "http:80:5000"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"openresty.letsencrypt\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku openresty:set --global letsencrypt-email test@example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:stop"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:start"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:inspect $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"openresty.letsencrypt\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "http:80:5000 https:443:5000"
}

@test "(openresty) allowed-domains" {
  run /bin/bash -c "dokku proxy:set $TEST_APP openresty"
  echo "output: $output"
  echo "status: $status"
  assert_success

  value="$(echo 'return true' | base64 -w 0)"
  run /bin/bash -c "dokku openresty:set --global allowed-letsencrypt-domains-func-base64 $value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:start"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker exec openresty-openresty-1 /usr/local/openresty/nginx/sbin/nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_success

  body='allowed_domains = {"domain.com", "extra-domain.com"}
  for index, value in ipairs(allowed_domains) do
    if value == domain then
      return true
    end
  end
  return false
  '
  value="$(echo "$body" | base64 -w 0)"
  run /bin/bash -c "dokku openresty:set --global allowed-letsencrypt-domains-func-base64 $value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:stop"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:start"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker exec openresty-openresty-1 /usr/local/openresty/nginx/sbin/nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:set --global allowed-letsencrypt-domains-func-base64"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:stop"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:start"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(openresty) includes" {
  run /bin/bash -c "dokku proxy:set $TEST_APP openresty"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_openresty_include
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:inspect $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"openresty.include-http-example.conf\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "IyBmb3JjZSB0aGUgY2hhcmFjdGVyIHNldCB0byB1dGYtOApjaGFyc2V0IFVURi04Owo="

  run /bin/bash -c "docker logs openresty-openresty-1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker exec openresty-openresty-1 /usr/local/openresty/nginx/sbin/nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker exec openresty-openresty-1 cat /etc/nginx/sites-enabled/sites.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "force the character set to utf-8"
  assert_output_contains "charset UTF-8;"
}

@test "(openresty) label management" {
  run /bin/bash -c "dokku proxy:set $TEST_APP openresty"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:label:add $TEST_APP openresty.directive value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:label:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "openresty.directive=value"

  run /bin/bash -c "dokku openresty:label:show $TEST_APP openresty.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "value"

  run /bin/bash -c "dokku openresty:label:show $TEST_APP openresty.directive2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"openresty.directive\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "value"

  run /bin/bash -c "dokku openresty:label:remove $TEST_APP openresty.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:label:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "openresty.directive=value"

  run /bin/bash -c "dokku openresty:label:show $TEST_APP openresty.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"openresty.directive\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
}

add_openresty_include() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mkdir -p "$APP_REPO_DIR/openresty/http-includes"
  touch "$APP_REPO_DIR/openresty/http-includes/example.conf"
  echo "# force the character set to utf-8" >>"$APP_REPO_DIR/openresty/http-includes/example.conf"
  echo "charset UTF-8;" >>"$APP_REPO_DIR/openresty/http-includes/example.conf"

  mkdir -p "$APP_REPO_DIR/openresty/http-location-includes"
  touch "$APP_REPO_DIR/openresty/http-location-includes/example.conf"
  echo "# location-block" >>"$APP_REPO_DIR/openresty/http-location-includes/example.conf"
}
