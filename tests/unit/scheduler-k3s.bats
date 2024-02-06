#!/usr/bin/env bats

load test_helper

TEST_APP="rdmtestapp"

setup() {
  uninstall_k3s || true
  global_setup
  dokku nginx:stop
}

teardown_() {
  global_teardown
  dokku nginx:start
  uninstall_k3s || true
}

install_k3s() {
  run /bin/bash -c "dokku proxy:set --global k3s"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set --global server hub.docker.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set --global image-repo-template '$DOCKERHUB_USERNAME/{{ .AppName }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set --global push-on-release true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler:set --global selected k3s"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:login docker.io $DOCKERHUB_USERNAME $DOCKERHUB_TOKEN"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Login Succeeded"

  run /bin/bash -c "dokku scheduler-k3s:initialize"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

uninstall_k3s() {
  run /bin/bash -c "dokku scheduler-k3s:uninstall"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(scheduler-k3s) install" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  install_k3s
}

@test "(scheduler-k3s) deploy" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sleep 20"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response "http" "$TEST_APP.dokku.me" "80" "" "python/http.server"
}

@test "(scheduler-k3s) deploy annotations" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type secret test.dokku.com/resource-type secret"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type pod test.dokku.com/resource-type pod"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type deployment test.dokku.com/resource-type deployment"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --process-type web --resource-type deployment test.dokku.com/resource-type deployment-web"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sleep 20"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response "http" "$TEST_APP.dokku.me" "80" "" "python/http.server"

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.metadata.annotations.\"test.dokku.com/resource-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "deployment-web"

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.spec.template.metadata.annotations.\"test.dokku.com/resource-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "pod"

  run /bin/bash -c "kubectl get cronjob -o json | jq -r '.items[0].metadata.annotations.\"test.dokku.com/resource-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "null"

  run /bin/bash -c "kubectl get cronjob -o json | jq -r '.items[0].spec.jobTemplate.metadata.annotations.\"test.dokku.com/resource-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "pod"

  run /bin/bash -c "kubectl get secret -o json | jq -r '.items[0].metadata.annotations.\"test.dokku.com/resource-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "secret"
}
