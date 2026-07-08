#!/usr/bin/env bats

load test_helper

NGINX_VALIDATE_PLUGIN_NAME="test-nginx-validate"
NGINX_VALIDATE_ZONE_CONF="/etc/nginx/conf.d/00-dokku-test-limit-req.conf"

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
}

teardown() {
  rm -rf "${PLUGIN_ENABLED_PATH:?}/$NGINX_VALIDATE_PLUGIN_NAME" "${PLUGIN_AVAILABLE_PATH:?}/$NGINX_VALIDATE_PLUGIN_NAME"
  destroy_app
  rm -f "$NGINX_VALIDATE_ZONE_CONF"
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
