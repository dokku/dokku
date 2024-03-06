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

  INGRESS_CLASS="${INGRESS_CLASS:-traefik}"
  if [[ "$TAINT_SCHEDULING" == "true" ]]; then
    run /bin/bash -c "dokku scheduler-k3s:initialize --ingress-class $INGRESS_CLASS --taint-scheduling"
    echo "output: $output"
    echo "status: $status"
    assert_success
  else
    run /bin/bash -c "dokku scheduler-k3s:initialize --ingress-class $INGRESS_CLASS"
    echo "output: $output"
    echo "status: $status"
    assert_success
  fi
}

uninstall_k3s() {
  run /bin/bash -c "dokku scheduler-k3s:uninstall"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(scheduler-k3s) install traefik with taint" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  TAINT_SCHEDULING=true install_k3s
}

@test "(scheduler-k3s) install nginx with taint" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx TAINT_SCHEDULING=true install_k3s
}

@test "(scheduler-k3s) deploy traefik [resource] [autoscaling]" {
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

  run /bin/bash -c "dokku scheduler-k3s:autoscaling-auth:set node-js-app memory --metadata some-key=1234567890 --metadata some-value=asdfghjkl"
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
  assert_output "kta-test-memory"

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
  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*]..resources.requests.cpu}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "100m"

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*]..resources.requests.memory}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "128Mi"

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*]..resources.limits.cpu}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*]..resources.limits.memory}'"
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

  run /bin/bash -c "sleep 20"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response "http" "$TEST_APP.dokku.me" "80" "" "python/http.server"

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*]..resources.requests.cpu}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*]..resources.requests.memory}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "300Mi"

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*]..resources.limits.cpu}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "kubectl get pods -o=jsonpath='{.items[*]..resources.limits.memory}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "512Mi"
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
  assert_output "job"

  run /bin/bash -c "kubectl get secret -o json | jq -r '.items[0].metadata.annotations.\"test.dokku.com/resource-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "secret"
}
