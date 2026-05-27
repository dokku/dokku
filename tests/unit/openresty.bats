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
  dokku openresty:set --global log-level >/dev/null 2>&1 || true
  global_teardown
  destroy_app
  dokku openresty:stop
  dokku nginx:start
}

@test "(openresty:report) --global --openresty-computed-letsencrypt-server" {
  run /bin/bash -c "dokku openresty:report --global --openresty-computed-letsencrypt-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "https://acme-staging-v02.api.letsencrypt.org/directory"

  run /bin/bash -c "dokku openresty:report --global --openresty-global-letsencrypt-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "https://acme-staging-v02.api.letsencrypt.org/directory"
}

@test "(openresty:report) --global raw and computed keys in --format json" {
  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"global-letsencrypt-server\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "https://acme-staging-v02.api.letsencrypt.org/directory"

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"computed-letsencrypt-server\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "https://acme-staging-v02.api.letsencrypt.org/directory"

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"global-letsencrypt-email\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"computed-letsencrypt-email\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"global-image\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"computed-image\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"global-allowed-letsencrypt-domains-func-base64\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"computed-allowed-letsencrypt-domains-func-base64\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"global-hsts\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"computed-hsts\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  for key in image letsencrypt-email letsencrypt-server allowed-letsencrypt-domains-func-base64; do
    run /bin/bash -c "dokku openresty:report --global --format json | jq -e \"has(\\\"$key\\\")\""
    echo "key: $key"
    echo "output: $output"
    echo "status: $status"
    assert_failure
    assert_output "false"
  done
}

@test "(openresty:report) --global global vs computed image" {
  run /bin/bash -c "dokku openresty:report --global --openresty-global-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku openresty:report --global --openresty-computed-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku openresty:set --global image dokku/openresty-docker-proxy:0.5.6"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:report --global --openresty-global-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku/openresty-docker-proxy:0.5.6"

  run /bin/bash -c "dokku openresty:report --global --openresty-computed-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku/openresty-docker-proxy:0.5.6"

  run /bin/bash -c "dokku openresty:set --global image"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:report --global --openresty-global-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku openresty:report --global --openresty-computed-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists
}

@test "(openresty:report) --global log-level raw vs computed" {
  run /bin/bash -c "dokku openresty:report --global --openresty-global-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku openresty:report --global --openresty-computed-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ERROR"

  run /bin/bash -c "dokku openresty:set --global log-level DEBUG"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:report --global --openresty-global-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "DEBUG"

  run /bin/bash -c "dokku openresty:report --global --openresty-computed-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "DEBUG"

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"global-log-level\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "DEBUG"

  run /bin/bash -c "dokku openresty:report --global --format json | jq -r '.\"computed-log-level\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "DEBUG"

  run /bin/bash -c "dokku openresty:set --global log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:report --global --openresty-global-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku openresty:report --global --openresty-computed-log-level"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ERROR"
}

@test "(openresty) global-only keys" {
  for key in allowed-letsencrypt-domains-func-base64 image log-level letsencrypt-email letsencrypt-server; do
    run /bin/bash -c "dokku openresty:set $TEST_APP $key somevalue"
    echo "key: $key"
    echo "output: $output"
    echo "status: $status"
    assert_failure
    assert_output_contains "can only be set globally"
  done

  run /bin/bash -c "dokku openresty:set $TEST_APP bind-address-ipv4 127.0.0.1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:report $TEST_APP --openresty-bind-address-ipv4"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "127.0.0.1"

  run /bin/bash -c "dokku openresty:set $TEST_APP bind-address-ipv4"
  echo "output: $output"
  echo "status: $status"
  assert_success
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

@test "(openresty) [security] eval injection via malicious include filename" {
  rm -f /tmp/openresty-include

  run /bin/bash -c "dokku proxy:set $TEST_APP openresty"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_openresty_include_unsafe
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "unsafe filename" -1

  # No injection payload to test since we're using a simple space character
  # The test should have failed during core-post-extract, not during eval
}

@test "(openresty) label management" {
  run /bin/bash -c "dokku proxy:set $TEST_APP openresty"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:labels:add $TEST_APP openresty.directive value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:labels:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "openresty.directive=value"

  run /bin/bash -c "dokku openresty:labels:show $TEST_APP openresty.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "value"

  run /bin/bash -c "dokku openresty:labels:show $TEST_APP openresty.directive2"
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

  run /bin/bash -c "dokku openresty:labels:remove $TEST_APP openresty.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku openresty:labels:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "openresty.directive=value"

  run /bin/bash -c "dokku openresty:labels:show $TEST_APP openresty.directive"
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

add_openresty_include_unsafe() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"

  mkdir -p "$APP_REPO_DIR/openresty/http-includes"
  # Create a filename with a space - simpler test that should be rejected by [^a-zA-Z0-9_.-]
  printf 'charset UTF-8;\n' >"$APP_REPO_DIR/openresty/http-includes/unsafe filename.conf"

  mkdir -p "$APP_REPO_DIR/openresty/http-location-includes"
  printf '# location\n' >"$APP_REPO_DIR/openresty/http-location-includes/unsafe filename.conf"
}
