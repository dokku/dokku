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

@test "(scheduler-k3s) [ingress] traefik redirects http to https when tls is enabled" {
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

  run /bin/bash -c "kubectl get ingressroute ${TEST_APP}-web-http-80-5000 -n default -o jsonpath='{.spec.routes[0].middlewares[0].name}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "${TEST_APP}-web-redirect-to-https"

  run /bin/bash -c "kubectl get ingressroute ${TEST_APP}-web-http-80-5000 -n default -o jsonpath='{.spec.tls.secretName}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "kubectl get ingressroute ${TEST_APP}-web-http-80-5000-websecure -n default -o jsonpath='{.spec.tls.secretName}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "tls-${TEST_APP}"

  assert_http_redirect "http://$TEST_APP.dokku.me" "https://$TEST_APP.dokku.me/"
  assert_http_localhost_response "https" "$TEST_APP.dokku.me" "443" "" "python/http.server"
}
