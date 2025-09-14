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

@test "(scheduler-k3s) security context" {
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

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.spec.template.spec.containers[0].securityContext.capabilities.add[0]'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "null"

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.spec.template.spec.containers[0].securityContext.privileged'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "null"

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy --cap-add=NET_ADMIN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy --privileged"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:restart $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.spec.template.spec.containers[0].securityContext.capabilities.add[0]'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "NET_ADMIN"

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.spec.template.spec.containers[0].securityContext.privileged'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"
}

@test "(scheduler-k3s) kustomize" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 worker=2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP inject_kustomization_yaml
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get deployment $TEST_APP-web -o json | jq -r '.spec.replicas'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "3"

  run /bin/bash -c "kubectl get deployment $TEST_APP-worker -o json | jq -r '.spec.replicas'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"
}

@test "(scheduler-k3s) deploy kustomize with vector sink" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  encoded="$(echo '{{ print "{{ pod }}" }}' | base64)"
  run /bin/bash -c "echo $encoded"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "e3sgcHJpbnQgInt7IHBvZCB9fSIgfX0K"

  run /bin/bash -c "dokku logs:set --global vector-sink 'http://?process=base64enc%3A${encoded}'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "kubectl get cm -n vector vector -o yaml"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "process: '{{ pod }}'"
}

inject_kustomization_yaml() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mkdir -p "$APP_REPO_DIR/config/kustomize"
  touch "$APP_REPO_DIR/config/kustomize/kustomization.yaml"
  cat <<EOF >"$APP_REPO_DIR/config/kustomize/kustomization.yaml"
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- rendered.yaml
patches:
- patch: |-
    - op: replace
      path: /spec/replicas
      value: 3
  target:
    group: apps
    version: v1
    kind: Deployment
    name: $APP-web
EOF
}
