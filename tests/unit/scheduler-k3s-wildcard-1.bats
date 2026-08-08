#!/usr/bin/env bats

load test_helper

TEST_APP="rdmtestapp"
EXACT_APP="rdmtestapp2"

setup() {
  uninstall_k3s || true
  global_setup
  dokku nginx:stop
  export KUBECONFIG="/etc/rancher/k3s/k3s.yaml"
}

teardown() {
  global_teardown
  dokku nginx:start
  uninstall_k3s || true
}

@test "(scheduler-k3s) [ingress] traefik routes wildcard domains and prefers exact domains" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=traefik install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP '*.dokku.me'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set $TEST_APP HELLO=wildcard"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:create $EXACT_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $EXACT_APP exact.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set $EXACT_APP HELLO=exact"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$EXACT_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sleep 30"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get ingressroutes.traefik.io ${TEST_APP}-web-http-80-5000 -n default -o jsonpath='{.spec.routes[0].match}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output 'HostRegexp(`{subdomain:[^.]+}.dokku.me`)'

  run /bin/bash -c "kubectl get ingressroutes.traefik.io ${TEST_APP}-web-http-80-5000 -n default -o jsonpath='{.spec.routes[0].priority}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "kubectl get ingressroutes.traefik.io ${EXACT_APP}-web-http-80-5000 -n default -o jsonpath='{.spec.routes[0].priority}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  assert_http_localhost_response "http" "anything.dokku.me" "80" "/hello" "wildcard"
  assert_http_localhost_response "http" "exact.dokku.me" "80" "/hello" "exact"
}
