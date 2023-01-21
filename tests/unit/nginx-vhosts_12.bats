#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
}

teardown() {
  destroy_app 0 $TEST_APP
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) git:from-image nginx.conf.sigil" {
  local CUSTOM_TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$CUSTOM_TMP"' INT TERM
  rmdir "$CUSTOM_TMP" && cp -r "${BATS_TEST_DIRNAME}/../../tests/apps/python" "$CUSTOM_TMP"

  pushd $CUSTOM_TMP
  custom_nginx_template "$TEST_APP" "$CUSTOM_TMP"

  run /bin/bash -c "dokku nginx:set --global nginx-conf-sigil-path"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "git config --global init.defaultBranch master"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "git init"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "git add nginx.conf.sigil"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "git commit -m 'Add custom nginx.conf.sigil'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image build -t dokku-test/$TEST_APP:latest -f $CUSTOM_TMP/alt.Dockerfile $CUSTOM_TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP dokku-test/$TEST_APP:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Overriding default nginx.conf with detected nginx.conf.sigil"

  run /bin/bash -c "dokku nginx:set --global nginx-conf-sigil-path dokku/nginx.conf.sigil"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Overriding default nginx.conf with detected nginx.conf.sigil" 0

  run /bin/bash -c "mkdir -p $CUSTOM_TMP/dokku"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "mv $CUSTOM_TMP/nginx.conf.sigil $CUSTOM_TMP/dokku/nginx.conf.sigil"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "git add ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "git commit -m 'Move the nginx.conf.sigil'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image build -t dokku-test/$TEST_APP:v2 -f $CUSTOM_TMP/alt.Dockerfile $CUSTOM_TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP dokku-test/$TEST_APP:v2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Overriding default nginx.conf with detected nginx.conf.sigil"

  run /bin/bash -c "dokku nginx:set --global nginx-conf-sigil-path"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Overriding default nginx.conf with detected nginx.conf.sigil" 0

  run /bin/bash -c "dokku nginx:set $TEST_APP nginx-conf-sigil-path dokku/nginx.conf.sigil"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Overriding default nginx.conf with detected nginx.conf.sigil"

  run /bin/bash -c "docker image rm dokku-test/$TEST_APP:latest dokku-test/$TEST_APP:v2"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
