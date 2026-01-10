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

@test "(scheduler-k3s:certs) app destruction deletes k8s TLS secret" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler:set $TEST_APP selected k3s"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku certs:add $TEST_APP $BATS_TMPDIR/tls/server.crt $BATS_TMPDIR/tls/server.key"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Installing TLS certificate for $TEST_APP"

  run /bin/bash -c "kubectl get secret tls-$TEST_APP -n default"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:destroy $TEST_APP --force"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get secret tls-$TEST_APP -n default"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
