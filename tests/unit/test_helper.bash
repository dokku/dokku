#!/usr/bin/env bash

# constants
DOKKU_ROOT=${DOKKU_ROOT:=~dokku}
DOKKU_LIB_ROOT=${DOKKU_LIB_PATH:="/var/lib/dokku"}
PLUGIN_PATH=${PLUGIN_PATH:="$DOKKU_LIB_ROOT/plugins"}
PLUGIN_AVAILABLE_PATH=${PLUGIN_AVAILABLE_PATH:="$PLUGIN_PATH/available"}
PLUGIN_ENABLED_PATH=${PLUGIN_ENABLED_PATH:="$PLUGIN_PATH/enabled"}
PLUGIN_CORE_PATH=${PLUGIN_CORE_PATH:="$DOKKU_LIB_ROOT/core-plugins"}
PLUGIN_CORE_AVAILABLE_PATH=${PLUGIN_CORE_AVAILABLE_PATH:="$PLUGIN_CORE_PATH/available"}
CUSTOM_TEMPLATE_SSL_DOMAIN=customssltemplate.dokku.me
UUID=$(tr -dc 'a-z0-9' < /dev/urandom | fold -w 32 | head -n 1)
TEST_APP="rdmtestapp-${UUID}"

# test functions
flunk() {
  { if [[ "$#" -eq 0 ]]; then cat -
    else echo "$*"
    fi
  }
  return 1
}

# ShellCheck doesn't know about $status from Bats
# shellcheck disable=SC2154
# shellcheck disable=SC2120
assert_success() {
  if [[ "$status" -ne 0 ]]; then
    flunk "command failed with exit status $status"
  elif [[ "$#" -gt 0 ]]; then
    assert_output "$1"
  fi
}

assert_failure() {
  if [[ "$status" -eq 0 ]]; then
    flunk "expected failed exit status"
  elif [[ "$#" -gt 0 ]]; then
    assert_output "$1"
  fi
}

assert_equal() {
  if [[ "$1" != "$2" ]]; then
    { echo "expected: $1"
      echo "actual:   $2"
    } | flunk
  fi
}

# ShellCheck doesn't know about $output from Bats
# shellcheck disable=SC2154
assert_output() {
  local expected
  if [[ $# -eq 0 ]]; then
    expected="$(cat -)"
  else
    expected="$1"
  fi
  assert_equal "$expected" "$output"
}

# ShellCheck doesn't know about $output from Bats
# shellcheck disable=SC2154
assert_output_contains() {
  local input="$output"; local expected="$1"; local count="${2:-1}"; local found=0
  until [ "${input/$expected/}" = "$input" ]; do
    input="${input/$expected/}"
    let found+=1
  done
  assert_equal "$count" "$found"
}

# ShellCheck doesn't know about $lines from Bats
# shellcheck disable=SC2154
assert_line() {
  if [[ "$1" -ge 0 ]] 2>/dev/null; then
    assert_equal "$2" "${lines[$1]}"
  else
    local line
    for line in "${lines[@]}"; do
      [[ "$line" = "$1" ]] && return 0
    done
    flunk "expected line \`$1'"
  fi
}

refute_line() {
  if [[ "$1" -ge 0 ]] 2>/dev/null; then
    local num_lines="${#lines[@]}"
    if [[ "$1" -lt "$num_lines" ]]; then
      flunk "output has $num_lines lines"
    fi
  else
    local line
    for line in "${lines[@]}"; do
      if [[ "$line" = "$1" ]]; then
        flunk "expected to not find line \`$line'"
      fi
    done
  fi
}

assert() {
  if ! "$*"; then
    flunk "failed: $*"
  fi
}

assert_exit_status() {
  assert_equal "$status" "$1"
}

# dokku functions
create_app() {
  dokku apps:create "$TEST_APP"
}

destroy_app() {
  local RC="$1"; local RC=${RC:=0}
  dokku --force apps:destroy "$TEST_APP"
  return "$RC"
}

add_domain() {
  dokku domains:add "$TEST_APP" "$1"
}

# shellcheck disable=SC2119
check_urls() {
  local PATTERN="$1"
  run bash -c "dokku --quiet urls $TEST_APP | egrep \"${1}\""
  echo "output: $output"
  echo "status: $status"
  assert_success
}

assert_http_success() {
  local url=$1
  run curl -kSso /dev/null -w "%{http_code}" "${url}"
  echo "output: $output"
  echo "status: $status"
  assert_output "200"
}

assert_ssl_domain() {
  local domain=$1
  assert_app_domain "${domain}"
  assert_http_redirect "http://${domain}" "https://${domain}:443/"
  assert_http_success "https://${domain}"
}

assert_nonssl_domain() {
  local domain=$1
  assert_app_domain "${domain}"
  assert_http_success "http://${domain}"
}

assert_app_domain() {
  local domain=$1
  run /bin/bash -c "dokku domains $TEST_APP | grep -xF ${domain}"
  echo "output: $output"
  echo "status: $status"
  assert_output "${domain}"
}

assert_http_redirect() {
  local from=$1
  local to=$2
  run curl -kSso /dev/null -w "%{redirect_url}" "${from}"
  echo "output: $output"
  echo "status: $status"
  assert_output "${to}"
}

deploy_app() {
  local APP_TYPE="$1"; local APP_TYPE=${APP_TYPE:="nodejs-express"}
  local GIT_REMOTE="$2"; local GIT_REMOTE=${GIT_REMOTE:="dokku@dokku.me:$TEST_APP"}
  local CUSTOM_TEMPLATE="$3"; local TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  local CUSTOM_PATH="$4"

  rmdir "$TMP" && cp -r "./tests/apps/$APP_TYPE" "$TMP"

  # shellcheck disable=SC2086
  [[ -n "$CUSTOM_TEMPLATE" ]] && $CUSTOM_TEMPLATE $TEST_APP $TMP/$CUSTOM_PATH

  pushd "$TMP" &> /dev/null || exit 1
  trap 'popd &> /dev/null || true; rm -rf "$TMP"' RETURN INT TERM

  git init
  git config user.email "robot@example.com"
  git config user.name "Test Robot"
  echo "setting up remote: $GIT_REMOTE"
  git remote add target "$GIT_REMOTE"

  [[ -f gitignore ]] && mv gitignore .gitignore
  git add .
  git commit -m 'initial commit'
  git push target master || destroy_app $?
}

setup_client_repo() {
  local TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  rmdir "$TMP" && cp -r ./tests/apps/nodejs-express "$TMP"
  cd "$TMP" || exit 1
  git init
  git config user.email "robot@example.com"
  git config user.name "Test Robot"

  [[ -f gitignore ]] && mv gitignore .gitignore
  git add .
  git commit -m 'initial commit'
}

setup_test_tls() {
  local TLS_TYPE="$1"; local TLS="/home/dokku/$TEST_APP/tls"
  mkdir -p "$TLS"

  case "$TLS_TYPE" in
    wildcard)
      local TLS_ARCHIVE=server_ssl_wildcard.tar
      ;;
    sans)
      local TLS_ARCHIVE=server_ssl_sans.tar
      ;;
    *)
      local TLS_ARCHIVE=server_ssl.tar
      ;;
  esac
  tar xf "$BATS_TEST_DIRNAME/$TLS_ARCHIVE" -C "$TLS"
  sudo chown -R dokku:dokku "${TLS}/.."
}

custom_ssl_nginx_template() {
  local APP="$1"; local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mkdir -p "$APP_REPO_DIR"

  echo "injecting custom_ssl_nginx_template -> $APP_REPO_DIR/nginx.conf.sigil"
cat<<EOF > "$APP_REPO_DIR/nginx.conf.sigil"
server {
  listen      [::]:{{ .NGINX_PORT }};
  listen      {{ .NGINX_PORT }};
  server_name {{ .NOSSL_SERVER_NAME }} $CUSTOM_TEMPLATE_SSL_DOMAIN;
  return 301 https://\$host:{{ .NGINX_SSL_PORT }}\$request_uri;
}

server {
  listen      [::]:{{ .NGINX_SSL_PORT }} ssl spdy;
  listen      {{ .NGINX_SSL_PORT }} ssl spdy;
  server_name \$SSL_SERVER_NAME $CUSTOM_TEMPLATE_SSL_DOMAIN;
  ssl_certificate     {{ .APP_SSL_PATH }}/server.crt;
  ssl_certificate_key {{ .APP_SSL_PATH }}/server.key;

  keepalive_timeout   70;
  add_header          Alternate-Protocol  {{ .NGINX_SSL_PORT }}:npn-spdy/2;
  location    / {
    proxy_pass  http://{{ .APP }};
    proxy_http_version 1.1;
    proxy_set_header Upgrade \$http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host \$http_host;
    proxy_set_header X-Forwarded-Proto \$scheme;
    proxy_set_header X-Forwarded-For \$remote_addr;
    proxy_set_header X-Forwarded-Port \$server_port;
    proxy_set_header X-Request-Start \$msec;
  }
  include {{ .DOKKU_ROOT }}/{{ .APP }}/nginx.conf.d/*.conf;
}
{{ if .DOKKU_APP_LISTENERS }}
upstream {{ .APP }} {
{{ range .DOKKU_APP_LISTENERS | split " " }}  server {{ . }};
{{ end }}}
{{ else if .PASSED_LISTEN_IP_PORT }}
upstream {{ .APP }} {
  server {{ .DOKKU_APP_LISTEN_IP }}:{{ .DOKKU_APP_LISTEN_PORT }};
}
{{ end }}
EOF
}

custom_nginx_template() {
  local APP="$1"; local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mkdir -p "$APP_REPO_DIR"

  echo "injecting custom_nginx_template -> $APP_REPO_DIR/nginx.conf.sigil"
cat<<EOF > "$APP_REPO_DIR/nginx.conf.sigil"
server {
  listen      [::]:{{ .NGINX_PORT }};
  listen      {{ .NGINX_PORT }};
  server_name {{ .NOSSL_SERVER_NAME }} customtemplate.dokku.me;

  location    / {
    proxy_pass  http://{{ .APP }};
    proxy_http_version 1.1;
    proxy_set_header Upgrade \$http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host \$http_host;
    proxy_set_header X-Forwarded-Proto \$scheme;
    proxy_set_header X-Forwarded-For \$remote_addr;
    proxy_set_header X-Forwarded-Port \$server_port;
    proxy_set_header X-Request-Start \$msec;
  }
  include {{ .DOKKU_ROOT }}/{{ .APP }}/nginx.conf.d/*.conf;
}
{{ if .DOKKU_APP_LISTENERS }}
upstream {{ .APP }} {
{{ range .DOKKU_APP_LISTENERS | split " " }}  server {{ . }};
{{ end }}}
{{ else if .PASSED_LISTEN_IP_PORT }}
upstream {{ .APP }} {
  server {{ .DOKKU_APP_LISTEN_IP }}:{{ .DOKKU_APP_LISTEN_PORT }};
}
{{ end }}
EOF
}

bad_custom_nginx_template() {
  local APP="$1"; local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting bad_custom_nginx_template -> $APP_REPO_DIR/nginx.conf.sigil"
cat<<EOF > "$APP_REPO_DIR/nginx.conf.sigil"
some lame nginx config
EOF
}
