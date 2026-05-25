#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) proxy:build-config (wildcard SSL and custom nginx template)" {
  setup_test_tls wildcard
  run /bin/bash -c "dokku domains:add $TEST_APP wildcard1.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP wildcard2.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app nodejs-express dokku@$DOKKU_DOMAIN:$TEST_APP custom_ssl_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_ssl_domain "wildcard1.${DOKKU_DOMAIN}"
  assert_ssl_domain "wildcard2.${DOKKU_DOMAIN}"
  assert_http_redirect "http://${CUSTOM_TEMPLATE_SSL_DOMAIN}" "https://${CUSTOM_TEMPLATE_SSL_DOMAIN}:443/"
  assert_http_success "https://${CUSTOM_TEMPLATE_SSL_DOMAIN}"
}

@test "(nginx-vhosts) proxy:build-config (custom nginx template - no ssl)" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP worker=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_nonssl_domain "www.test.app.${DOKKU_DOMAIN}"
  assert_http_localhost_response "http" "customtemplate.${DOKKU_DOMAIN}"

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "${TEST_APP}-worker-5000"
}

@test "(nginx-vhosts) proxy:build-config (disable custom nginx template - no ssl)" {
  run /bin/bash -c "dokku nginx:set $TEST_APP  disable-custom-config true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP worker=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_nonssl_domain "www.test.app.${DOKKU_DOMAIN}"

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "${TEST_APP}-worker-5000" 0
}

@test "(nginx-vhosts) proxy:build-config (failed validate_nginx)" {
  run deploy_app nodejs-express dokku@$DOKKU_DOMAIN:$TEST_APP bad_custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Pre-validating custom nginx.conf.sigil"
  assert_output_contains "Custom nginx.conf.sigil failed nginx -t validation"
}

@test "(nginx-vhosts) proxy:build-config (custom nginx template with deprecated listen http2)" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP custom_nginx_template_with_http2_listen
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_output_contains "Deprecated: Usage of 'listen ... http2' within nginx.conf.sigil templates is deprecated" -1
}

custom_nginx_template_with_http2_listen() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mkdir -p "$APP_REPO_DIR"

  echo "injecting custom_nginx_template_with_http2_listen -> $APP_REPO_DIR/nginx.conf.sigil"
  cat <<EOF >"$APP_REPO_DIR/nginx.conf.sigil"
{{ range \$port_map := .PROXY_PORT_MAP | split " " }}
{{ \$port_map_list := \$port_map | split ":" }}
{{ \$scheme := index \$port_map_list 0 }}
{{ \$listen_port := index \$port_map_list 1 }}
{{ \$upstream_port := index \$port_map_list 2 }}

server {
  listen      [::]:{{ \$listen_port }} http2;
  listen      {{ \$listen_port }} http2;
  server_name {{ $.NOSSL_SERVER_NAME }} customtemplate.${DOKKU_DOMAIN};

  location    / {
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
