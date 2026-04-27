#!/usr/bin/env bats

# Tests for the nginx:reload command and the catch-all default site shipped
# at /etc/nginx/conf.d/00-default-vhost.conf. Postinst-driven behavior
# (debconf flag, fresh-install detection, sites-enabled/default rename) is
# not exercised here because the bats harness does not reinstall the apt
# package; the rename logic is exercised manually by simulating the
# conflicting upstream vhost.

load test_helper

NGINX_DEFAULT_VHOST_PATH="/etc/nginx/conf.d/00-default-vhost.conf"
NGINX_DEFAULT_VHOST_SOURCE="/var/lib/dokku/core-plugins/available/nginx-vhosts/templates/default-site.conf"

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  [[ -f "$NGINX_DEFAULT_VHOST_PATH" ]] && cp -fp "$NGINX_DEFAULT_VHOST_PATH" "${NGINX_DEFAULT_VHOST_PATH}.bats-bak"
  rm -f "$NGINX_DEFAULT_VHOST_PATH"
}

teardown() {
  rm -f "$NGINX_DEFAULT_VHOST_PATH"
  if [[ -f "${NGINX_DEFAULT_VHOST_PATH}.bats-bak" ]]; then
    mv "${NGINX_DEFAULT_VHOST_PATH}.bats-bak" "$NGINX_DEFAULT_VHOST_PATH"
  fi
  rm -f /etc/nginx/sites-enabled/default.bats-stub
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  if sudo nginx -t &>/dev/null; then
    sudo systemctl reload nginx || true
  fi
  global_teardown
}

@test "(nginx-vhosts:reload) reloads with valid config" {
  run /bin/bash -c "dokku nginx:reload"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(nginx-vhosts:reload) fails with invalid config without restarting" {
  local invalid_conf="/etc/nginx/conf.d/dokku-bats-invalid.conf"
  echo "this is not valid nginx config" | sudo tee "$invalid_conf" >/dev/null

  run /bin/bash -c "dokku nginx:reload"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  sudo rm -f "$invalid_conf"
}

@test "(nginx-vhosts) [default-site] shipped template installs and validates" {
  [[ -f "$NGINX_DEFAULT_VHOST_SOURCE" ]] || skip "default-site template not installed; run 'sudo make copyfiles'"

  sudo install -m 0644 -o root -g root "$NGINX_DEFAULT_VHOST_SOURCE" "$NGINX_DEFAULT_VHOST_PATH"

  run /bin/bash -c "sudo grep -F 'ssl_reject_handshake on' '$NGINX_DEFAULT_VHOST_PATH'"
  assert_success
  run /bin/bash -c "sudo grep -F 'return 444' '$NGINX_DEFAULT_VHOST_PATH'"
  assert_success
  run /bin/bash -c "sudo grep -F 'default_server' '$NGINX_DEFAULT_VHOST_PATH'"
  assert_success

  run /bin/bash -c "sudo nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(nginx-vhosts) [default-site] coexists with stock sites-enabled/default" {
  [[ -f "$NGINX_DEFAULT_VHOST_SOURCE" ]] || skip "default-site template not installed; run 'sudo make copyfiles'"

  sudo mkdir -p /etc/nginx/sites-enabled
  sudo tee /etc/nginx/sites-enabled/default.bats-stub >/dev/null <<'STOCK'
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name _;
    return 200 "stock default";
}
STOCK

  sudo install -m 0644 -o root -g root "$NGINX_DEFAULT_VHOST_SOURCE" "$NGINX_DEFAULT_VHOST_PATH"

  run /bin/bash -c "sudo nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  sudo rm -f /etc/nginx/sites-enabled/default.bats-stub

  run /bin/bash -c "sudo nginx -t"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
