#!/usr/bin/env bats

load test_helper

TEST_APP="rdmtestapp"

setup() {
  uninstall_k3s || true
  global_setup
  dokku nginx:stop
  export KUBECONFIG="/etc/rancher/k3s/k3s.yaml"
}

teardown_() {
  global_teardown
  dokku nginx:start
  uninstall_k3s || true
}

@test "(scheduler-k3s) deploy traefik [resource] [autoscaling1]" {
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

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:set $TEST_APP memory --metadata some-key=1234567890 --metadata some-value=asdfghjkl"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response "http" "$TEST_APP.dokku.me" "80" "" "python/http.server"

  # include autoscaling tests
  run /bin/bash -c "kubectl get scaledobjects.keda.sh $TEST_APP-web -o=jsonpath='{.spec.triggers[0].authenticationRef.name}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP-memory"

  run /bin/bash -c "kubectl get triggerauthentications.keda.sh $TEST_APP-memory -o=jsonpath='{.spec.secretTargetRef[0].key}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "some-key"

  run /bin/bash -c "kubectl get triggerauthentications.keda.sh $TEST_APP-memory -o=jsonpath='{.spec.secretTargetRef[0].name}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "kta-$TEST_APP-memory"

  run /bin/bash -c "kubectl get triggerauthentications.keda.sh $TEST_APP-memory -o=jsonpath='{.spec.secretTargetRef[0].parameter}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "some-key"

  run /bin/bash -c "kubectl get secret kta-$TEST_APP-memory -o=jsonpath='{.data.some-key}' | base64 --decode"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1234567890"

  run /bin/bash -c "kubectl get secret kta-$TEST_APP-memory -o=jsonpath='{.data.some-value}' | base64 --decode"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "asdfghjkl"

  # include resource tests
  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*].spec.containers[*].resources.requests.cpu}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "100m"

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*].spec.containers[*].resources.requests.memory}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "128Mi"

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*].spec.containers[*].resources.limits.cpu}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*].spec.containers[*].resources.limits.memory}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku resource:reserve $TEST_APP --memory 300 --cpu 0m --process-type web"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku resource:limit $TEST_APP --memory 512 --process-type web"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response "http" "$TEST_APP.dokku.me" "80" "" "python/http.server"

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*].spec.containers[*].resources.requests.cpu}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*].spec.containers[*].resources.requests.memory}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "300Mi"

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*].spec.containers[*].resources.limits.cpu}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*].spec.containers[*].resources.limits.memory}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "512Mi"

  # include run tests
  run /bin/bash -c "dokku run $TEST_APP ls -lah"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # include enter tests
  run /bin/bash -c "dokku enter $TEST_APP web ls -lah"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(scheduler-k3s) deploy nginx" {
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

  run /bin/bash -c "dokku scheduler-k3s:set $TEST_APP shm-size 64Mi"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response "http" "$TEST_APP.dokku.me" "80" "" "python/http.server"

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.spec.template.spec.volumes[0].emptyDir.sizeLimit'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "64Mi"

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.spec.template.spec.volumes[0].emptyDir.medium'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "Memory"
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

  run /bin/bash -c "dokku scheduler-k3s:annotations:set $TEST_APP --resource-type job test.dokku.com/resource-type job"
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
  assert_output "job"

  run /bin/bash -c "kubectl get secret -o json | jq -r '.items[0].metadata.annotations.\"test.dokku.com/resource-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "secret"
}
