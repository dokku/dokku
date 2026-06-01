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

@test "(proxy-route:k3s-traefik) basic routing - /api/v0 hits api, / hits web" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=traefik install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP" add_api_process_callback
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 api=1"
  assert_success

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "python3 -m http.server"
}

@test "(proxy-route:k3s-traefik) --strip-prefix renders stripPrefix Middleware and strips upstream path" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=traefik install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP" add_api_process_callback
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 api=1"
  assert_success

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001 --strip-prefix"
  assert_success

  # The chart should emit a StripPrefix Middleware CRD when --strip-prefix
  # is set.
  run /bin/bash -c "kubectl get middleware.traefik.io -o json | jq -r '.items[].spec.stripPrefix.prefixes[]?' | head -1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "/api/v0"

  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "python3 -m http.server"
}

@test "(proxy-route:k3s-traefik) removing a route drops the stripPrefix Middleware" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=traefik install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP" add_api_process_callback
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 api=1"
  assert_success

  run /bin/bash -c "dokku proxy:route:set $TEST_APP api /api/v0 --port 5001 --strip-prefix"
  assert_success
  assert_http_localhost_response_contains "http" "${TEST_APP}.${DOKKU_DOMAIN}" "80" "/api/v0/Procfile" "python3 -m http.server"

  run /bin/bash -c "dokku proxy:route:remove $TEST_APP /api/v0"
  assert_success

  run /bin/bash -c "kubectl get middleware.traefik.io -o json | jq -r '[.items[] | select(.spec.stripPrefix != null)] | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"
}

add_api_process_callback() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  echo "api: python3 -m http.server 5001" >>"$APP_REPO_DIR/Procfile"
}
