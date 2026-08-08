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
  dokku scheduler-k3s:set --global cert-issuer-name >/dev/null 2>/dev/null || true
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

@test "(scheduler-k3s:certs) a manually managed cert issuer issues the app certificate" {
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

  # no letsencrypt email is ever set, so tls is enabled purely by cert-issuer-name
  run /bin/bash -c "dokku scheduler-k3s:set $TEST_APP cert-issuer-name dokku-test-selfsigned"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:preview $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "ClusterIssuer dokku-test-selfsigned not found"

  create_selfsigned_cluster_issuer

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get certificate ${TEST_APP}-web -n default -o jsonpath='{.spec.issuerRef.kind}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ClusterIssuer"

  run /bin/bash -c "kubectl get certificate ${TEST_APP}-web -n default -o jsonpath='{.spec.issuerRef.name}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku-test-selfsigned"

  # a selfSigned issuer issues immediately, so the secret proves end-to-end issuance
  run /bin/bash -c "kubectl wait --for=condition=Ready certificate/${TEST_APP}-web -n default --timeout=120s"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get secret tls-${TEST_APP}-web -n default -o jsonpath='{.type}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "kubernetes.io/tls"

  run /bin/bash -c "kubectl get ingress ${TEST_APP}-web-${TEST_APP}-dokku-me -n default -o jsonpath='{.spec.tls[0].secretName}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "tls-${TEST_APP}-web"
}

@test "(scheduler-k3s:certs) letsencrypt-server false disables a global cert issuer" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  create_selfsigned_cluster_issuer

  run /bin/bash -c "dokku scheduler-k3s:set --global cert-issuer-name dokku-test-selfsigned"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:set $TEST_APP letsencrypt-server false"
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

  run /bin/bash -c "kubectl get certificate ${TEST_APP}-web -n default"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "kubectl get ingress ${TEST_APP}-web-${TEST_APP}-dokku-me -n default -o jsonpath='{.spec.tls}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""
}

create_selfsigned_cluster_issuer() {
  declare desc="creates a selfSigned ClusterIssuer that issues certificates without acme"

  local manifest="${BATS_TMPDIR:-/tmp}/dokku-test-selfsigned.yaml"
  cat >"$manifest" <<'EOF'
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: dokku-test-selfsigned
spec:
  selfSigned: {}
EOF

  run /bin/bash -c "kubectl apply -f $manifest"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
