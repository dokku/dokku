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

@test "(scheduler-k3s:secrets) helm rollback of app chart leaves config and pull secret releases intact" {
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

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sleep 30"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "helm rollback $TEST_APP 1 -n default"
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
}

@test "(scheduler-k3s:secrets) user image-pull-secrets override skips dokku-managed pull secret release" {
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

  run /bin/bash -c "dokku scheduler-k3s:set $TEST_APP image-pull-secrets my-custom-pull-secret"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get secret pull-secret-$TEST_APP -n default"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "kubectl get deployment ${TEST_APP}-web -n default -o jsonpath='{.spec.template.spec.imagePullSecrets[*].name}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "my-custom-pull-secret"
}

@test "(scheduler-k3s:secrets) leaked imagePullSecrets entries get pruned on next deploy" {
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

  run /bin/bash -c "kubectl patch deployment ${TEST_APP}-web -n default --type=json -p='[{\"op\":\"add\",\"path\":\"/spec/template/spec/imagePullSecrets/-\",\"value\":{\"name\":\"ims-${TEST_APP}.111\"}},{\"op\":\"add\",\"path\":\"/spec/template/spec/imagePullSecrets/-\",\"value\":{\"name\":\"ims-${TEST_APP}.222\"}}]'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get deployment ${TEST_APP}-web -n default -o jsonpath='{.spec.template.spec.imagePullSecrets[*].name}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "ims-${TEST_APP}.111"

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sleep 30"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get deployment ${TEST_APP}-web -n default -o jsonpath='{.spec.template.spec.imagePullSecrets[*].name}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "pull-secret-$TEST_APP"
}
