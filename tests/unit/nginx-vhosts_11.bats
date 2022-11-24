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

@test "(nginx-vhosts) nginx:set proxy-buffer-size" {
  deploy_app

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-buffer-size 2k"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "proxy_buffer_size 2k;"

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-buffer-size"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "proxy_buffer_size 2k;" 0
}

@test "(nginx-vhosts) nginx:set proxy-buffering" {
  deploy_app

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-buffering off"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "proxy_buffering off;"

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-buffering"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "proxy_buffering off;" 0
}

@test "(nginx-vhosts) nginx:set proxy-buffers" {
  deploy_app

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-buffers \"64 4k\""
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "proxy_buffers 64 4k;"

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-buffers"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "proxy_buffers 64 4k;" 0
}

@test "(nginx-vhosts) nginx:set proxy-busy-buffers-size" {
  deploy_app

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-busy-buffers-size 10k"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "proxy_busy_buffers_size 10k;"

  run /bin/bash -c "dokku nginx:set $TEST_APP proxy-busy-buffers-size"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "proxy_busy_buffers_size 10k;" 0
}


@test "(nginx-vhosts) nginx:set nginx.conf.sigil" {
  local TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM
  export CUSTOM_TMP="$TMP"

  run deploy_app python
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Overriding default nginx.conf with detected nginx.conf.sigil" 0

  pushd $TMP
  custom_nginx_template "$TEST_APP" "$CUSTOM_TMP"

  run /bin/bash -c "git add nginx.conf.sigil"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "git commit -m 'Add custom nginx.conf.sigil'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "git push target master:master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Overriding default nginx.conf with detected nginx.conf.sigil"

  mv "$CUSTOM_TMP/nginx.conf.sigil" "$CUSTOM_TMP/nginx.conf.sigil-2"

  run /bin/bash -c "git add nginx.conf.sigil nginx.conf.sigil-2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "git commit -m 'Moving nginx.conf.sigil'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "git push target master:master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Overriding default nginx.conf with detected nginx.conf.sigil" 0

  run /bin/bash -c "dokku nginx:set $TEST_APP nginx-conf-sigil-path nginx.conf.sigil-2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Overriding default nginx.conf with detected nginx.conf.sigil"

  popd
  rm -rf "$TMP"
}
