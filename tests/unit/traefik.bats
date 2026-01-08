#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  dokku nginx:stop
  dokku traefik:set --global letsencrypt-server https://acme-staging-v02.api.letsencrypt.org/directory
  dokku traefik:set --global letsencrypt-email
  dokku traefik:set --global api-enabled
  dokku traefik:start
  create_app
}

teardown() {
  global_teardown
  destroy_app
  dokku traefik:stop
  dokku nginx:start
}

@test "(traefik) traefik:help" {
  run /bin/bash -c "dokku traefik"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the traefik proxy integration"
  help_output="$output"

  run /bin/bash -c "dokku traefik:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the traefik proxy integration"
  assert_output "$help_output"
}

@test "(traefik) single domain" {
  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP convert_to_dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "curl --silent $(dokku url $TEST_APP)"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "python/http.server"
}

@test "(traefik) multiple domains" {
  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP $TEST_APP.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP $TEST_APP-2.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP convert_to_dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "curl --silent $TEST_APP.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "python/http.server"

  run /bin/bash -c "curl --silent $TEST_APP-2.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "python/http.server"
}

@test "(traefik) traefik:set api" {
  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect traefik-traefik-1 --format '{{ index .Config.Labels \"traefik.http.routers.api.rule\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists
  assert_success

  run /bin/bash -c "dokku traefik:set --global api-enabled false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:stop"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:start"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect traefik-traefik-1 --format '{{ index .Config.Labels \"traefik.http.routers.api.rule\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists
  assert_success

  run /bin/bash -c "dokku traefik:set --global api-enabled true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:stop"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:start"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect traefik-traefik-1 --format '{{ index .Config.Labels \"traefik.http.routers.api.rule\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_output "Host(\`traefik.$DOKKU_DOMAIN\`)"
  assert_success
}

@test "(traefik) ssl" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-http.loadbalancer.server.port\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "5000"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-https.loadbalancer.server.port\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku traefik:set --global letsencrypt-email test@example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:stop"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:start"
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

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-http.loadbalancer.server.port\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "5000"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-https.loadbalancer.server.port\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "5000"

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists

  run /bin/bash -c "dokku --quiet ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_output "http:80:5000 https:443:5000"
}

@test "(traefik) show-config without auth set" {
  run /bin/bash -c "dokku traefik:set --global basic-auth-username \"\""
  run /bin/bash -c "dokku traefik:set --global basic-auth-password \"\""

  run /bin/bash -c "dokku traefik:show-config"
  echo "output: $output"
  echo "status: $status"
  assert_success
  refute_line '      - "traefik.http.routers.api.middlewares=auth"'
}

@test "(traefik) show-config with auth set" {
  run /bin/bash -c "dokku traefik:set --global basic-auth-username test-username"
  run /bin/bash -c "dokku traefik:set --global basic-auth-password test-password"

  run /bin/bash -c "dokku traefik:show-config"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line '      - "traefik.http.routers.api.middlewares=auth"'
}

@test "(traefik) change traefik entry point http" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:set --global http-entry-point web"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.routers.$TEST_APP-web-http.entrypoints\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "web"

}

@test "(traefik) change traefik entry point https" {
  run /bin/bash -c "dokku builder-herokuish:set $TEST_APP allowed true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:set --global letsencrypt-email test@example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:set --global https-entry-point websecure"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.routers.$TEST_APP-web-https.entrypoints\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "websecure"
}

@test "(traefik) label management" {
  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:labels:add $TEST_APP traefik.directive value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:labels:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "traefik.directive=value"

  run /bin/bash -c "dokku traefik:labels:show $TEST_APP traefik.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "value"

  run /bin/bash -c "dokku traefik:labels:show $TEST_APP traefik.directive2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.directive\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "value"

  run /bin/bash -c "dokku traefik:labels:remove $TEST_APP traefik.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku traefik:labels:show $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "traefik.directive=value"

  run /bin/bash -c "dokku traefik:labels:show $TEST_APP traefik.directive"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.directive\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
}

@test "(traefik) healthcheck labels from app.json" {
  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP setup_traefik_healthcheck
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-http.loadbalancer.healthcheck.path\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-http.loadbalancer.healthcheck.timeout\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "5s"

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-http.loadbalancer.healthcheck.interval\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2s"
}

@test "(traefik) no healthcheck labels without readiness check" {
  run /bin/bash -c "dokku proxy:set $TEST_APP traefik"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect $TEST_APP.web.1 --format '{{ index .Config.Labels \"traefik.http.services.$TEST_APP-web-http.loadbalancer.healthcheck.path\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
}

setup_traefik_healthcheck() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  APP_REPO_DIR="$(realpath "$APP_REPO_DIR")"

  mv "$APP_REPO_DIR/app-traefik.json" "$APP_REPO_DIR/app.json"
}
