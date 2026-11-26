#!/usr/bin/env bats

load test_helper

NGINX_HTTP3_BACKUP_DIR="/tmp/dokku-nginx-http3-backup"
HTTP3_LISTENERS_FILE="$DOKKU_LIB_ROOT/data/nginx-vhosts/http3/listeners.conf"
SECOND_APP="${TEST_APP}-second"

setup_file() {
  install_mainline_nginx
}

teardown_file() {
  restore_distro_nginx
}

setup() {
  global_setup
  create_app
}

teardown() {
  dokku nginx:set --global http3 || true
  global_teardown
}

@test "(nginx-vhosts) [http3] nginx:set http3 renders quic listeners" {
  setup_test_tls wildcard

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set $TEST_APP http3 true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "skipping quic listeners" 0

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "quic;" 2
  assert_output_contains "http3 on;"
  assert_output_contains "add_header Alt-Svc 'h3=\":443\"; ma=86400' always;"
  assert_output_contains "reuseport" 0

  run /bin/bash -c "grep -c 'quic reuseport;' $HTTP3_LISTENERS_FILE"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  run /bin/bash -c "nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set $TEST_APP http3"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "quic" 0

  run /bin/bash -c "grep -c listen $HTTP3_LISTENERS_FILE"
  echo "output: $output"
  echo "status: $status"
  assert_output "0"
}

@test "(nginx-vhosts) [http3] apps sharing a quic listener get a single reuseport" {
  setup_test_tls wildcard

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set $TEST_APP http3 true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  TEST_APP="$SECOND_APP" setup_test_tls wildcard

  run /bin/bash -c "dokku ports:set $SECOND_APP http:80:5000 https:443:5000"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set $SECOND_APP http3 true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $SECOND_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $SECOND_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "quic;" 2

  run /bin/bash -c "grep -c 'quic reuseport;' $HTTP3_LISTENERS_FILE"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  run /bin/bash -c "grep -h 'reuseport' $DOKKU_ROOT/*/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "grep -c 'quic reuseport;' $HTTP3_LISTENERS_FILE"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  run /bin/bash -c "dokku nginx:set $SECOND_APP http3 false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $SECOND_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "grep -c listen $HTTP3_LISTENERS_FILE"
  echo "output: $output"
  echo "status: $status"
  assert_output "0"

  run /bin/bash -c "nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(nginx-vhosts) [http3] nginx:set --global http3" {
  setup_test_tls wildcard

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set --global http3 true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "quic;" 2

  run /bin/bash -c "dokku nginx:set $TEST_APP http3 false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "quic" 0

  run /bin/bash -c "grep -c listen $HTTP3_LISTENERS_FILE"
  echo "output: $output"
  echo "status: $status"
  assert_output "0"
}

install_mainline_nginx() {
  declare desc="replaces the distro nginx with an nginx.org mainline build that supports http3"
  local codename

  codename="$(. /etc/os-release && echo "$VERSION_CODENAME")"
  rm -rf "$NGINX_HTTP3_BACKUP_DIR"
  cp -a /etc/nginx "$NGINX_HTTP3_BACKUP_DIR"

  curl -fsSL https://nginx.org/keys/nginx_signing.key | gpg --dearmor --yes -o /usr/share/keyrings/nginx-archive-keyring.gpg
  echo "deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] http://nginx.org/packages/mainline/ubuntu $codename nginx" >/etc/apt/sources.list.d/nginx-mainline.list
  printf 'Package: *\nPin: origin nginx.org\nPin-Priority: 900\n' >/etc/apt/preferences.d/99nginx-mainline

  dokku nginx:stop
  install_nginx_packages nginx
  start_nginx
}

restore_distro_nginx() {
  declare desc="restores the distro nginx that was installed before install_mainline_nginx"
  rm -f /etc/apt/sources.list.d/nginx-mainline.list /etc/apt/preferences.d/99nginx-mainline /usr/share/keyrings/nginx-archive-keyring.gpg

  dokku nginx:stop
  printf 'Package: nginx nginx-common\nPin: release o=Ubuntu\nPin-Priority: 1001\n' >/etc/apt/preferences.d/99nginx-distro
  install_nginx_packages nginx nginx-common
  rm -f /etc/apt/preferences.d/99nginx-distro
  start_nginx
}

start_nginx() {
  declare desc="starts nginx, retrying while the service supervisor is not yet ready"
  local attempt

  for attempt in $(seq 1 15); do
    if dokku nginx:start; then
      return 0
    fi
    sleep 1
  done

  return 1
}

install_nginx_packages() {
  declare desc="installs nginx packages without starting services, then restores the saved nginx config"
  local created_policy=false

  if [[ ! -f /usr/sbin/policy-rc.d ]]; then
    printf '#!/bin/sh\nexit 101\n' >/usr/sbin/policy-rc.d
    chmod +x /usr/sbin/policy-rc.d
    created_policy=true
  fi

  apt-get -qq update
  DEBIAN_FRONTEND=noninteractive apt-get -qq -y --allow-downgrades -o Dpkg::Options::=--force-confold install "$@"

  if [[ "$created_policy" == "true" ]]; then
    rm -f /usr/sbin/policy-rc.d
  fi

  rm -rf /etc/nginx
  cp -a "$NGINX_HTTP3_BACKUP_DIR" /etc/nginx
}
