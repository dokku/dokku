#!/usr/bin/env bats

load test_helper

TEST_APP="rdmtestapp"

setup_file() {
  install_pack
}

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

@test "(scheduler-k3s) cnb deployment sets command launcher for non-web process" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected pack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1 worker=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_requirements_txt_cnb
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'from cnb stack'

  run /bin/bash -c "kubectl get deployment $TEST_APP-worker -o=jsonpath='{.spec.template.spec.containers[0].command[0]}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "launcher"

  run /bin/bash -c "kubectl get deployment $TEST_APP-worker -o=jsonpath='{.spec.template.spec.containers[0].args[0]}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "python3"
}

@test "(scheduler-k3s) cnb cronjob manifest sets command launcher" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected pack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP cron_run_wrapper
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'from cnb stack'

  run /bin/bash -c "kubectl get cronjob -o=jsonpath='{.items[0].spec.jobTemplate.spec.template.spec.containers[0].command[0]}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "launcher"

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"
  run /bin/bash -c "echo $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  cron_hash="$(printf '%s' "$cron_id" | sha1sum | awk '{print $1}')"

  run /bin/bash -c "kubectl get cronjob -o json | jq -r '.items[0].metadata.annotations.\"dokku.com/cron-id\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$cron_id"

  run /bin/bash -c "kubectl get cronjob -o json | jq -r '.items[0].metadata.labels.\"dokku.com/cron-hash\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$cron_hash"

  run /bin/bash -c "kubectl get cronjob -o json | jq -r '.items[0].metadata.annotations.\"dokku.com/cron-hash\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$cron_hash"

  run /bin/bash -c "dokku --quiet cron:run $TEST_APP $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "['task.py', 'some', 'cron', 'task']"
}

@test "(scheduler-k3s) cnb dokku run uses launcher entrypoint" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected pack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_requirements_txt_cnb
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'from cnb stack'

  run /bin/bash -c "dokku --quiet run $TEST_APP python3 task.py test"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "['task.py', 'test']"

  run /bin/bash -c "dokku --quiet run $TEST_APP env"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "SECRET_KEY=fjdkslafjdk"
}

@test "(scheduler-k3s) cnb dokku run resolves Procfile key" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected pack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP cron_run_procfile_wrapper
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'from cnb stack'

  run /bin/bash -c "dokku --quiet run $TEST_APP task"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "['task.py', 'test']"

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"
  run /bin/bash -c "echo $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku --quiet cron:run $TEST_APP $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "['task.py', 'test']"
}

@test "(scheduler-k3s) cnb dokku run:detached returns pod name with launcher entrypoint" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected pack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_requirements_txt_cnb
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'from cnb stack'

  run /bin/bash -c "dokku --quiet run:detached $TEST_APP sleep 300"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_exists
  pod_name="$output"

  run /bin/bash -c "kubectl get pod $pod_name -o=jsonpath='{.metadata.labels.app\.kubernetes\.io/part-of}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP"

  run /bin/bash -c "kubectl get pod $pod_name -o=jsonpath='{.metadata.annotations.dokku\.com/builder-type}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "pack"

  run /bin/bash -c "kubectl get pod $pod_name -o=jsonpath='{.spec.containers[0].command[0]}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "launcher"

  run /bin/bash -c "kubectl get pod $pod_name -o=jsonpath='{.spec.containers[0].args[0]}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "sleep"
}

cron_run_wrapper() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  APP_REPO_DIR="$(realpath "$APP_REPO_DIR")"

  add_requirements_txt "$APP" "$APP_REPO_DIR"
  mv -f "$APP_REPO_DIR/app-cnb-cron.json" "$APP_REPO_DIR/app.json"
}

cron_run_procfile_wrapper() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  APP_REPO_DIR="$(realpath "$APP_REPO_DIR")"

  add_requirements_txt "$APP" "$APP_REPO_DIR"
  mv -f "$APP_REPO_DIR/app-cnb-cron-procfile.json" "$APP_REPO_DIR/app.json"
}
