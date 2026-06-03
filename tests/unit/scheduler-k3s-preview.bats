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

@test "(scheduler-k3s:preview) clean redeploy shows no diff" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:preview $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""
}

@test "(scheduler-k3s:preview) resource limit change appears in diff" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku resource:limit --process-type web --memory 256 $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:preview $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "has changed"
  assert_output_contains "256Mi"
}

@test "(scheduler-k3s:preview) undeployed app shows additive diff" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:preview $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "has been added"
}
