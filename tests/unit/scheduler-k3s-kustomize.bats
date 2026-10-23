#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(scheduler-k3s) [kustomize] deploy removes a stale non-empty kustomization directory" {
  local DATA_DIR="/var/lib/dokku/data/scheduler-k3s/$TEST_APP"
  mkdir -p "$DATA_DIR/kustomization.99999999"
  echo "kind: Kustomization" >"$DATA_DIR/kustomization.99999999/kustomization.yaml"
  chown -R dokku:dokku "$DATA_DIR"

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "find $DATA_DIR -maxdepth 1 -name 'kustomization.*'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
}

@test "(scheduler-k3s) [kustomize] deploy removes the stored kustomization when the repo no longer has one" {
  local DATA_DIR="/var/lib/dokku/data/scheduler-k3s/$TEST_APP"
  local CUSTOM_TMP=$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$CUSTOM_TMP"' INT TERM

  CUSTOM_TMP="$CUSTOM_TMP" run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP inject_kustomization_yaml
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f $DATA_DIR/kustomization/kustomization.yaml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$CUSTOM_TMP" rm -r config/kustomize
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$CUSTOM_TMP" commit -m 'Remove kustomization'
  echo "output: $output"
  echo "status: $status"
  assert_success

  run git -C "$CUSTOM_TMP" push target master:master
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -e $DATA_DIR/kustomization"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "find $DATA_DIR -maxdepth 1 -name 'kustomization.*'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  rm -rf "$CUSTOM_TMP"
}

inject_kustomization_yaml() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mkdir -p "$APP_REPO_DIR/config/kustomize"
  cat <<EOF >"$APP_REPO_DIR/config/kustomize/kustomization.yaml"
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- rendered.yaml
EOF
}
