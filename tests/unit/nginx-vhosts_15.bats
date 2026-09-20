#!/usr/bin/env bats

load test_helper

NGINX_VALIDATE_PLUGIN_NAME="test-nginx-validate"
NGINX_VALIDATE_LEGACY_PLUGIN_NAME="test-nginx-validate-legacy"
NGINX_VALIDATE_ZONE_CONF="/etc/nginx/conf.d/00-dokku-test-limit-req.conf"
NGINX_VALIDATE_BROKEN_CONF="/etc/nginx/conf.d/00-dokku-test-broken.conf"
NGINX_VALIDATE_CACHE_DIR="/var/cache/nginx/dokku-test-cache"

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
}

teardown() {
  rm -f "$NGINX_VALIDATE_ZONE_CONF" "$NGINX_VALIDATE_BROKEN_CONF"
  rm -rf "$NGINX_VALIDATE_CACHE_DIR"
  rm -rf "${PLUGIN_ENABLED_PATH:?}/$NGINX_VALIDATE_PLUGIN_NAME" "${PLUGIN_AVAILABLE_PATH:?}/$NGINX_VALIDATE_PLUGIN_NAME"
  rm -rf "${PLUGIN_ENABLED_PATH:?}/$NGINX_VALIDATE_LEGACY_PLUGIN_NAME" "${PLUGIN_AVAILABLE_PATH:?}/$NGINX_VALIDATE_LEGACY_PLUGIN_NAME"
  rm -rf "$DOKKU_LIB_ROOT/data/nginx-vhosts/app-$TEST_APP" || true
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) pre-validate fails fast on broken nginx.conf.sigil" {
  run deploy_app nodejs-express dokku@$DOKKU_DOMAIN:$TEST_APP bad_custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Pre-validating custom nginx.conf.sigil"
  assert_output_contains "Custom nginx.conf.sigil failed nginx -t validation"
  assert_output_not_contains "Building $TEST_APP"
  assert_output_not_contains "Releasing $TEST_APP"
}

@test "(nginx-vhosts) pre-validate succeeds on first deploy with valid nginx.conf.sigil" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Pre-validating custom nginx.conf.sigil"
}

@test "(nginx-vhosts) pre-validate is skipped when disable-custom-config=true" {
  run /bin/bash -c "dokku nginx:set $TEST_APP disable-custom-config true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app nodejs-express dokku@$DOKKU_DOMAIN:$TEST_APP bad_custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "Pre-validating custom nginx.conf.sigil"
}

@test "(nginx-vhosts) pre-validate is skipped when proxy is not nginx" {
  run /bin/bash -c "dokku proxy:set $TEST_APP caddy"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app nodejs-express dokku@$DOKKU_DOMAIN:$TEST_APP bad_custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "Pre-validating custom nginx.conf.sigil"
}

@test "(nginx-vhosts) pre-validate fails when nginx.conf.sigil needs a directive absent from the wrapper" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP custom_nginx_template_with_limit_req
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Pre-validating custom nginx.conf.sigil"
  assert_output_contains "Custom nginx.conf.sigil failed nginx -t validation"
}

@test "(nginx-vhosts) pre-validate passes when nginx-app-template-source overrides validate-config" {
  echo 'limit_req_zone $binary_remote_addr zone=dokkuprevalidate:10m rate=100r/s;' >"$NGINX_VALIDATE_ZONE_CONF"
  setup_nginx_validate_plugin

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP custom_nginx_template_with_limit_req
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Pre-validating custom nginx.conf.sigil"
}

@test "(nginx-vhosts) [pre-validate] fn-nginx-vhosts-nginx-user parses the main nginx config" {
  source "$PLUGIN_AVAILABLE_PATH/nginx-vhosts/internal-functions"

  cat <<'EOF' >"$BATS_TEST_TMPDIR/nginx.conf"
#user nobody;
  user   www-data www-data; # workers run as www-data
userid on;
EOF

  run fn-nginx-vhosts-nginx-user "$BATS_TEST_TMPDIR/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_output "www-data"
}

@test "(nginx-vhosts) [pre-validate] fn-nginx-vhosts-nginx-user ignores configs declaring no user" {
  source "$PLUGIN_AVAILABLE_PATH/nginx-vhosts/internal-functions"

  cat <<'EOF' >"$BATS_TEST_TMPDIR/nginx.conf"
#user nobody;
userid on;
auth_basic_user_file conf/htpasswd;
EOF

  run fn-nginx-vhosts-nginx-user "$BATS_TEST_TMPDIR/nginx.conf"
  echo "output: $output"
  echo "status: $status"
  assert_output ""

  run fn-nginx-vhosts-nginx-user "$BATS_TEST_TMPDIR/missing.conf"
  echo "output: $output"
  echo "status: $status"
  assert_output ""
}

@test "(nginx-vhosts) [pre-validate] does not change proxy_cache_path directory ownership" {
  local NGINX_USER
  NGINX_USER="$(nginx_host_user)"
  [[ -n "$NGINX_USER" ]] || skip "host nginx config declares no user directive"

  setup_nginx_cache_dir "$NGINX_USER"
  nginx_cache_config "$BATS_TEST_TMPDIR/nginx.conf.sigil"

  run_plugin_script nginx-vhosts core-post-extract "$TEST_APP" "$BATS_TEST_TMPDIR"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Pre-validating custom nginx.conf.sigil"

  run /bin/bash -c "stat -c %U $NGINX_VALIDATE_CACHE_DIR"
  echo "output: $output"
  echo "status: $status"
  assert_output "$NGINX_USER"
}

@test "(nginx-vhosts) [pre-validate] preserves ownership with a validate-config template predating NGINX_USER" {
  local NGINX_USER
  NGINX_USER="$(nginx_host_user)"
  [[ -n "$NGINX_USER" ]] || skip "host nginx config declares no user directive"

  setup_nginx_validate_legacy_plugin
  setup_nginx_cache_dir "$NGINX_USER"
  nginx_cache_config "$BATS_TEST_TMPDIR/nginx.conf.sigil"

  run_plugin_script nginx-vhosts core-post-extract "$TEST_APP" "$BATS_TEST_TMPDIR"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Pre-validating custom nginx.conf.sigil"

  run /bin/bash -c "stat -c %U $NGINX_VALIDATE_CACHE_DIR"
  echo "output: $output"
  echo "status: $status"
  assert_output "$NGINX_USER"
}

@test "(nginx-vhosts:validate-config) does not change proxy_cache_path directory ownership" {
  local NGINX_USER
  NGINX_USER="$(nginx_host_user)"
  [[ -n "$NGINX_USER" ]] || skip "host nginx config declares no user directive"

  setup_nginx_cache_dir "$NGINX_USER"
  nginx_cache_config "$DOKKU_ROOT/$TEST_APP/nginx.conf"
  echo 'not_a_directive;' >"$NGINX_VALIDATE_BROKEN_CONF"

  run /bin/bash -c "dokku nginx:validate-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "stat -c %U $NGINX_VALIDATE_CACHE_DIR"
  echo "output: $output"
  echo "status: $status"
  assert_output "$NGINX_USER"
}

setup_nginx_validate_plugin() {
  declare desc="installs a plugin that overrides the validate-config nginx template"
  local PLUGIN_DIR="$PLUGIN_AVAILABLE_PATH/$NGINX_VALIDATE_PLUGIN_NAME"
  mkdir -p "$PLUGIN_DIR"

  cat <<EOF >"$PLUGIN_DIR/plugin.toml"
[plugin]
description = "test plugin overriding the validate-config nginx template"
version = "0.1.0"
[plugin.config]
EOF

  cat <<'EOF' >"$PLUGIN_DIR/nginx-app-template-source"
#!/usr/bin/env bash

set -eo pipefail
[[ $DOKKU_TRACE ]] && set -x

TEMPLATE_TYPE="$2"
if [[ "$TEMPLATE_TYPE" == "validate-config" ]]; then
  echo "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/validate.conf.sigil"
fi
EOF
  chmod +x "$PLUGIN_DIR/nginx-app-template-source"

  cat <<'EOF' >"$PLUGIN_DIR/validate.conf.sigil"
{{ if $.NGINX_USER }}user {{ $.NGINX_USER }};{{ end }}
events { worker_connections 768; }
http {
  access_log off;
  error_log /dev/null;
  limit_req_zone $binary_remote_addr zone=dokkuprevalidate:10m rate=100r/s;
  include {{ $.NGINX_CONF }};
}
EOF

  dokku plugin:enable "$NGINX_VALIDATE_PLUGIN_NAME"
}

custom_nginx_template_with_limit_req() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mkdir -p "$APP_REPO_DIR"

  echo "injecting custom_nginx_template_with_limit_req -> $APP_REPO_DIR/nginx.conf.sigil"
  cat <<EOF >"$APP_REPO_DIR/nginx.conf.sigil"
{{ range \$port_map := .PROXY_PORT_MAP | split " " }}
{{ \$port_map_list := \$port_map | split ":" }}
{{ \$scheme := index \$port_map_list 0 }}
{{ \$listen_port := index \$port_map_list 1 }}
{{ \$upstream_port := index \$port_map_list 2 }}

server {
  listen      [::]:{{ \$listen_port }};
  listen      {{ \$listen_port }};
  server_name {{ $.NOSSL_SERVER_NAME }} customtemplate.${DOKKU_DOMAIN};

  location    / {
    limit_req zone=dokkuprevalidate burst=100 nodelay;
    proxy_pass  http://{{ $.APP }}-{{ \$upstream_port }};
    proxy_http_version 1.1;
    proxy_set_header Upgrade \$http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host \$http_host;
    proxy_set_header X-Forwarded-Proto \$scheme;
    proxy_set_header X-Forwarded-For \$remote_addr;
    proxy_set_header X-Forwarded-Port \$server_port;
    proxy_set_header X-Request-Start \$msec;
  }
  include {{ $.DOKKU_ROOT }}/{{ $.APP }}/nginx.conf.d/*.conf;
}
{{ end }}

{{ if $.DOKKU_APP_WEB_LISTENERS }}
{{ range \$upstream_port := $.PROXY_UPSTREAM_PORTS | split " " }}
upstream {{ $.APP }}-{{ \$upstream_port }} {
{{ range \$listeners := $.DOKKU_APP_WEB_LISTENERS | split " " }}
{{ \$listener_list := \$listeners | split ":" }}
{{ \$listener_ip := index \$listener_list 0 }}
  server {{ \$listener_ip }}:{{ \$upstream_port }};{{ end }}
}
{{ end }}{{ end }}

EOF
  cat "$APP_REPO_DIR/nginx.conf.sigil"
}

nginx_host_user() {
  declare desc="returns the user the host nginx runs its workers as"

  awk '{ sub(/#.*/, "") } $1 == "user" && NF >= 2 { value = $2; sub(/;.*$/, "", value); print value; exit }' /etc/nginx/nginx.conf
}

setup_nginx_cache_dir() {
  declare desc="creates a proxy_cache_path directory owned by the host nginx user"
  local NGINX_USER="$1"

  mkdir -p "$NGINX_VALIDATE_CACHE_DIR"
  chown "$NGINX_USER" "$NGINX_VALIDATE_CACHE_DIR"
}

nginx_cache_config() {
  declare desc="writes an nginx config declaring a proxy_cache_path at the test cache directory"
  local OUTPUT_PATH="$1"

  echo "injecting nginx_cache_config -> $OUTPUT_PATH"
  cat <<EOF >"$OUTPUT_PATH"
proxy_cache_path $NGINX_VALIDATE_CACHE_DIR levels=1:2 keys_zone=dokkuvalidatecache:1m max_size=10m inactive=1m use_temp_path=off;

server {
  listen 127.0.0.1:18999;
  server_name dokku-validate-cache.test;

  location / {
    proxy_cache dokkuvalidatecache;
    proxy_pass http://127.0.0.1:18998;
  }
}
EOF
}

setup_nginx_validate_legacy_plugin() {
  declare desc="installs a plugin whose validate-config template predates the NGINX_USER variable"
  local PLUGIN_DIR="$PLUGIN_AVAILABLE_PATH/$NGINX_VALIDATE_LEGACY_PLUGIN_NAME"
  mkdir -p "$PLUGIN_DIR"

  cat <<EOF >"$PLUGIN_DIR/plugin.toml"
[plugin]
description = "test plugin overriding the validate-config nginx template without a user directive"
version = "0.1.0"
[plugin.config]
EOF

  cat <<'EOF' >"$PLUGIN_DIR/nginx-app-template-source"
#!/usr/bin/env bash

set -eo pipefail
[[ $DOKKU_TRACE ]] && set -x

TEMPLATE_TYPE="$2"
if [[ "$TEMPLATE_TYPE" == "validate-config" ]]; then
  echo "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/validate.conf.sigil"
fi
EOF
  chmod +x "$PLUGIN_DIR/nginx-app-template-source"

  cat <<'EOF' >"$PLUGIN_DIR/validate.conf.sigil"
events { worker_connections 768; }
http {
  access_log off;
  error_log /dev/null;
  include {{ $.NGINX_CONF }};
}
EOF

  dokku plugin:enable "$NGINX_VALIDATE_LEGACY_PLUGIN_NAME"
}
