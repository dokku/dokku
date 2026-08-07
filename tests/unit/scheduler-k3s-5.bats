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

@test "(scheduler-k3s) docker-options sysctl" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.spec.template.spec.securityContext.sysctls'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "null"

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy '--sysctl net.ipv4.ip_unprivileged_port_start=1024'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:restart $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.spec.template.spec.securityContext.sysctls[0].name'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "net.ipv4.ip_unprivileged_port_start"

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.spec.template.spec.securityContext.sysctls[0].value'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1024"

  run /bin/bash -c "dokku docker-options:remove $TEST_APP deploy '--sysctl net.ipv4.ip_unprivileged_port_start=1024'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy '--sysctl vm.max_map_count=262144'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:restart $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "is not namespaced" -1
}
