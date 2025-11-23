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

@test "(scheduler-k3s) app.json defined service" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 worker=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP inject_app_json
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get services $TEST_APP-web -o json | jq -r '.spec.ports[0].port'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "80"

  run /bin/bash -c "kubectl get services $TEST_APP-web -o json | jq -r '.spec.ports[0].targetPort'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "5000"
}

inject_app_json() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "formation": {
    "worker": {
      "service": {
        "exposed": true
      }
    }
  }
}
EOF
}
