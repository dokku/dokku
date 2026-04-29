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

@test "(scheduler-k3s:secrets) deploy creates stable config and pull secrets" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP HELLO=world"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get secret config-$TEST_APP -n default"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get secret pull-secret-$TEST_APP -n default"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get deployment ${TEST_APP}-web -n default -o jsonpath='{.spec.template.spec.containers[0].envFrom[0].secretRef.name}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "config-$TEST_APP"

  run /bin/bash -c "kubectl get deployment ${TEST_APP}-web -n default -o jsonpath='{.spec.template.spec.imagePullSecrets[*].name}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "pull-secret-$TEST_APP"

  run /bin/bash -c "kubectl get secret config-$TEST_APP -n default -o jsonpath='{.data.HELLO}' | base64 --decode"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "world"
}

@test "(scheduler-k3s:secrets) config:set updates secret in place" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP HELLO=world"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set $TEST_APP HELLO=mars"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sleep 10"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get secret config-$TEST_APP -n default -o jsonpath='{.data.HELLO}' | base64 --decode"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "mars"
}

@test "(scheduler-k3s:secrets) apps:destroy cleans up secret releases" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
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

  run /bin/bash -c "kubectl get secret config-$TEST_APP -n default"
  assert_success

  run /bin/bash -c "kubectl get secret pull-secret-$TEST_APP -n default"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get secret config-$TEST_APP -n default"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "kubectl get secret pull-secret-$TEST_APP -n default"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
