#!/usr/bin/env bats

load test_helper

REDIRECT_SOURCE="www.test.${DOKKU_DOMAIN}"
REDIRECT_DESTINATION="www.test.app.${DOKKU_DOMAIN}"
ACME_TOKEN="proxy-redirect-acme-token"

setup() {
  global_setup
  create_app
}

teardown() {
  dokku proxy:redirect:clear "$TEST_APP" >/dev/null 2>&1 || true
  rm -rf "${DOKKU_LIB_ROOT:?}/data/letsencrypt/$TEST_APP"
  destroy_app
  global_teardown
}

@test "(proxy-redirect:nginx-letsencrypt) redirect sources stay in the app domain list" {
  run /bin/bash -c "dokku domains:set $TEST_APP $REDIRECT_DESTINATION $REDIRECT_SOURCE"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP $REDIRECT_SOURCE $REDIRECT_DESTINATION"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-vhosts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$REDIRECT_SOURCE"
  assert_output_contains "$REDIRECT_DESTINATION"

  assert_http_localhost_redirect "http" "$REDIRECT_SOURCE" "80" "/" "301" "http://$REDIRECT_DESTINATION/"
}

@test "(proxy-redirect:nginx-letsencrypt) redirect sources serve the http-01 challenge" {
  run /bin/bash -c "dokku domains:set $TEST_APP $REDIRECT_DESTINATION $REDIRECT_SOURCE"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP $REDIRECT_SOURCE $REDIRECT_DESTINATION"
  echo "output: $output"
  echo "status: $status"
  assert_success

  enable_acme_challenge "$TEST_APP"

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response "http" "$REDIRECT_SOURCE" "80" "/.well-known/acme-challenge/$ACME_TOKEN" "$ACME_TOKEN"
  assert_http_localhost_redirect "http" "$REDIRECT_SOURCE" "80" "/some/path" "301" "http://$REDIRECT_DESTINATION/some/path"
}

@test "(proxy-redirect:nginx-letsencrypt) https redirects render only when the certificate covers the source" {
  run /bin/bash -c "dokku domains:set $TEST_APP $REDIRECT_DESTINATION $REDIRECT_SOURCE"
  echo "output: $output"
  echo "status: $status"
  assert_success

  setup_test_tls
  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP $REDIRECT_SOURCE $REDIRECT_DESTINATION"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "grep -c 'server_name $REDIRECT_SOURCE;' $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  enable_acme_challenge "$TEST_APP"
  teardown_test_tls
  setup_test_tls sans
  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "conflicting server name"

  run /bin/bash -c "nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "conflicting server name"

  run /bin/bash -c "grep -c 'server_name $REDIRECT_SOURCE;' $DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  assert_http_localhost_redirect "https" "$REDIRECT_SOURCE" "443" "/some/path?query=1" "301" "https://$REDIRECT_DESTINATION/some/path?query=1"
  assert_http_localhost_redirect "http" "$REDIRECT_SOURCE" "80" "/" "301" "http://$REDIRECT_DESTINATION/"
  assert_http_localhost_response "http" "$REDIRECT_SOURCE" "80" "/.well-known/acme-challenge/$ACME_TOKEN" "$ACME_TOKEN"
}

enable_acme_challenge() {
  declare desc="writes the nginx include and challenge token that dokku-letsencrypt uses for the http-01 challenge"
  declare APP="$1"
  local challenge_root="$DOKKU_LIB_ROOT/data/letsencrypt/$APP"

  mkdir -p "$challenge_root/.well-known/acme-challenge" "$DOKKU_ROOT/$APP/nginx.conf.d"
  echo -n "$ACME_TOKEN" >"$challenge_root/.well-known/acme-challenge/$ACME_TOKEN"
  chmod -R a+rX "$DOKKU_LIB_ROOT/data/letsencrypt"
  cat >"$DOKKU_ROOT/$APP/nginx.conf.d/letsencrypt.conf" <<EOF
location /.well-known/acme-challenge {
  # allow every ip address
  allow all;

  # disable http-auth for /.well-known/acme-challenge
  auth_basic off;
  root $challenge_root;
}
EOF
  chown -R dokku:dokku "$DOKKU_ROOT/$APP/nginx.conf.d"
}
