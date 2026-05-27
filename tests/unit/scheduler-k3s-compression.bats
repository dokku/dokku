#!/usr/bin/env bats

load test_helper

TEST_APP="rdmtestapp"

setup_local_tls() {
  TLS=$BATS_TMPDIR/tls
  mkdir -p $TLS
  tar xf $BATS_TEST_DIRNAME/server_ssl.tar -C $TLS
  sudo chown -R dokku:dokku $TLS
}

teardown_local_tls() {
  TLS=$BATS_TMPDIR/tls
  rm -R $TLS
}

setup() {
  uninstall_k3s || true
  global_setup
  dokku nginx:stop
  export KUBECONFIG="/etc/rancher/k3s/k3s.yaml"
  setup_local_tls
}

teardown() {
  global_teardown
  dokku nginx:start
  uninstall_k3s || true
  teardown_local_tls
}

assert_http_localhost_header() {
  local scheme="$1" domain="$2" port="${3:-80}" path="${4:-/}" accept_encoding="$5" expected_header="$6"
  local retries="${HTTP_ASSERT_RETRIES:-30}" attempt=1

  while [[ "$attempt" -lt "$retries" ]]; do
    run /bin/bash -c "curl --connect-to '$domain:$port:localhost:$port' -kSsD - -o /dev/null -H 'Accept-Encoding: $accept_encoding' '$scheme://$domain:$port$path' | tr -d '\r' | grep -i '^$expected_header$'"
    [[ "$status" -eq 0 ]] && break
    sleep 1
    attempt=$((attempt + 1))
  done

  echo "output: $output"
  echo "status: $status"
  echo "attempts: $attempt"
  assert_success
  assert_output_contains "$expected_header"
}

@test "(scheduler-k3s) [ingress] traefik compression middleware adds gzip response header" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=traefik install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/server.crt $BATS_TMPDIR/tls/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sleep 30"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get ingressroutes.traefik.io ${TEST_APP}-web-http-80-5000 -n default -o jsonpath='{.spec.routes[0].middlewares[0].name}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "${TEST_APP}-web-compression"

  assert_http_localhost_response "https" "$TEST_APP.dokku.me" "443" "" "python/http.server"
  assert_http_localhost_header "https" "$TEST_APP.dokku.me" "443" "/" "gzip" "content-encoding: gzip"
}
