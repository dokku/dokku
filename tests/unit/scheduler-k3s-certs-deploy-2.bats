#!/usr/bin/env bats

load test_helper

TEST_APP="rdmtestapp"

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

@test "(scheduler-k3s:certs) app-level letsencrypt email renders a per-app namespaced Issuer" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:set $TEST_APP letsencrypt-server staging"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:set $TEST_APP letsencrypt-email-stag app@dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.dokku.me"
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

  run /bin/bash -c "kubectl get certificate ${TEST_APP}-web -n default -o jsonpath='{.spec.issuerRef.kind}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "Issuer"

  run /bin/bash -c "kubectl get certificate ${TEST_APP}-web -n default -o jsonpath='{.spec.issuerRef.name}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "${TEST_APP}-letsencrypt-stag"

  run /bin/bash -c "kubectl get issuer ${TEST_APP}-letsencrypt-stag -n default -o jsonpath='{.spec.acme.email}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "app@dokku.me"
}
