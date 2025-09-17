#!/usr/bin/env bash

# constants
DOKKU_ROOT=${DOKKU_ROOT:=~dokku}
DOCKER_BIN=${DOCKER_BIN:="docker"}
DOKKU_LIB_ROOT=${DOKKU_LIB_PATH:="/var/lib/dokku"}
DOKKU_DOMAIN=${DOKKU_DOMAIN:="dokku.me"}
PLUGIN_PATH=${PLUGIN_PATH:="$DOKKU_LIB_ROOT/plugins"}
PLUGIN_AVAILABLE_PATH=${PLUGIN_AVAILABLE_PATH:="$PLUGIN_PATH/available"}
PLUGIN_ENABLED_PATH=${PLUGIN_ENABLED_PATH:="$PLUGIN_PATH/enabled"}
PLUGIN_CORE_PATH=${PLUGIN_CORE_PATH:="$DOKKU_LIB_ROOT/core-plugins"}
PLUGIN_CORE_AVAILABLE_PATH=${PLUGIN_CORE_AVAILABLE_PATH:="$PLUGIN_CORE_PATH/available"}
CUSTOM_TEMPLATE_SSL_DOMAIN="customssltemplate.${DOKKU_DOMAIN}"
UUID=$(uuidgen)
TEST_APP="rdmtestapp-${UUID}"
TEST_NETWORK="test-network-${UUID}"
SKIPPED_TEST_ERR_MSG="previous test failed! skipping remaining tests..."
TEST_PLUGIN_NAME=smoke-test-plugin
TEST_PLUGIN_GIT_REPO=https://github.com/dokku/${TEST_PLUGIN_NAME}.git
TEST_PLUGIN_LOCAL_REPO="$(mktemp -d)/$TEST_PLUGIN_NAME"

# global setup() and teardown()
# skips remaining tests on first failure
global_setup() {
  [[ ! -f "${BATS_PARENT_TMPNAME}.skip" ]] || skip "$SKIPPED_TEST_ERR_MSG"

  free -m
  cleanup_apps
  cleanup_containers
}

global_teardown() {
  [[ -n "$BATS_TEST_COMPLETED" ]] || touch "${BATS_PARENT_TMPNAME}.skip"
  rm "${BATS_PARENT_TMPNAME}.skip" || true
  cleanup_apps
  cleanup_containers
}

cleanup_apps() {
  rm -rf $DOKKU_ROOT/*/nginx.conf

  dokku --quiet apps:list | xargs --no-run-if-empty -n1 dokku --force apps:destroy
}

cleanup_containers() {
  containers=$(docker container ls --quiet)
  if [[ -n "$containers" ]]; then
    docker inspect -f '{{ if ne "true" (index .Config.Labels "com.dokku.devcontainer") }}{{.ID}} {{ end }}' $(docker ps -q) | xargs --no-run-if-empty -n1 docker container rm -f || true
  fi
}

clone_test_plugin() {
  git clone "$TEST_PLUGIN_GIT_REPO" "$TEST_PLUGIN_LOCAL_REPO"
}

remove_test_plugin() {
  rm -rf "${PLUGIN_ENABLED_PATH:?}/$TEST_PLUGIN_NAME" "${PLUGIN_AVAILABLE_PATH:?}/$TEST_PLUGIN_NAME"
  rm -rf "$TEST_PLUGIN_LOCAL_REPO"
}

# test functions
flunk() {
  {
    if [[ "$#" -eq 0 ]]; then
      cat -
    else
      echo "$*"
    fi
  }
  return 1
}

# ShellCheck doesn't know about $status from Bats
assert_success() {
  if [[ "$status" -ne 0 ]]; then
    flunk "command failed with exit status $status"
  elif [[ "$#" -gt 0 ]]; then
    assert_output "$1"
  fi
}

# ShellCheck doesn't know about $status from Bats
assert_failure() {
  if [[ "$status" -eq 0 ]]; then
    flunk "expected failed exit status"
  elif [[ "$#" -gt 0 ]]; then
    assert_output "$1"
  fi
}

assert_equal() {
  if [[ "$1" != "$2" ]]; then
    {
      echo "expected: $1"
      echo "actual:   $2"
    } | flunk
  fi
}

assert_not_equal() {
  if [[ "$1" == "$2" ]]; then
    {
      echo "unexpected: $1"
      echo "actual:     $2"
    } | flunk
  fi
}

# ShellCheck doesn't know about $output from Bats
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
assert_not_output() {
  local expected
  if [[ $# -eq 0 ]]; then
    unexpected="$(cat -)"
  else
    unexpected="$1"
  fi
  assert_not_equal "$unexpected" "$output"
}

# ShellCheck doesn't know about $output from Bats
assert_output_exists() {
  [[ -n "$output" ]] || flunk "expected output, found none"
}

# ShellCheck doesn't know about $output from Bats
assert_output_not_exists() {
  [[ -z "$output" ]] || flunk "expected no output, found some"
}

# ShellCheck doesn't know about $output from Bats
assert_output_contains() {
  local input="$output"
  local expected="$1"
  local count="${2:-1}"
  local found=0
  until [ "${input/$expected/}" = "$input" ]; do
    input="${input/$expected/}"
    found=$((found + 1))
  done
  assert_equal "$count" "$found"
}

# ShellCheck doesn't know about $lines from Bats
assert_line() {
  if [[ "$1" -ge 0 ]] 2>/dev/null; then
    assert_equal "$2" "${lines[$1]}"
  else
    local line
    for line in "${lines[@]}"; do
      [[ "$line" == "$1" ]] && return 0
    done
    flunk "expected line \`$1'"
  fi
}

# ShellCheck doesn't know about $lines from Bats
assert_line_count() {
  declare EXPECTED="$1"
  local num_lines="${#lines[@]}"
  assert_equal "$EXPECTED" "$num_lines"
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
      if [[ "$line" == "$1" ]]; then
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
  local APP="$1"
  local TEST_APP=${APP:=$TEST_APP}
  dokku apps:create "$TEST_APP"
}

create_key() {
  ssh-keygen -P "" -f /tmp/testkey &>/dev/null
}

destroy_app() {
  declare RC="${1:-0}" APP="$2"
  local TEST_APP=${APP:=$TEST_APP}
  dokku --force apps:destroy "$TEST_APP"
  return "$RC"
}

destroy_key() {
  rm -f /tmp/testkey* &>/dev/null || true
}

check_urls() {
  local PATTERN="$1"
  run /bin/bash -c "dokku urls $TEST_APP | grep -E \"${PATTERN}\""
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

assert_http_localhost_response() {
  local scheme="$1" domain="$2" port="${3:-80}" path="${4:-}" content="${5:-}" status_code="${6:-200}"
  run curl --connect-to "$domain:$port:localhost:$port" -kSso /dev/null -w "%{http_code}" "$scheme://$domain:$port$path"
  echo "curl: curl --connect-to $domain:$port:localhost:$port -kSso /dev/null -w %{http_code} $scheme://$domain:$port$path"
  echo "output: $output"
  echo "status: $status"
  assert_output "$status_code"

  if [[ -n "$content" ]]; then
    run curl --connect-to "$domain:$port:localhost:$port" -kSs "$scheme://$domain:$port$path"
    echo "output: $output"
    echo "status: $status"
    assert_output "$content"
  fi
}

assert_http_localhost_response_contains() {
  local scheme="$1" domain="$2" port="${3:-80}" path="${4:-}" content="${5:-}" status_code="${6:-200}"
  run curl --connect-to "$domain:$port:localhost:$port" -kSso /dev/null -w "%{http_code}" "$scheme://$domain:$port$path"
  echo "curl: curl --connect-to $domain:$port:localhost:$port -kSso /dev/null -w %{http_code} $scheme://$domain:$port$path"
  echo "output: $output"
  echo "status: $status"
  assert_output "$status_code"

  if [[ -n "$content" ]]; then
    run curl --connect-to "$domain:$port:localhost:$port" -kSs "$scheme://$domain:$port$path"
    echo "output: $output"
    echo "status: $status"
    assert_output_contains "$content"
  fi
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
  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-vhosts | tr \" \" \"\n\" | grep -xF ${domain}"
  echo "app domains: $(dokku domains:report "$TEST_APP" --domains-app-vhosts)"
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

assert_external_port() {
  local CID="$1"
  local EXTERNAL_PORT_COUNT
  EXTERNAL_PORT_COUNT="$(docker port "$CID" | wc -l)"
  run /bin/bash -c "[[ $EXTERNAL_PORT_COUNT -gt 0 ]]"
  assert_success
}

assert_not_external_port() {
  local CID="$1"
  local EXTERNAL_PORT_COUNT
  EXTERNAL_PORT_COUNT="$(docker port "$CID" | wc -l)"
  run /bin/bash -c "[[ $EXTERNAL_PORT_COUNT -gt 0 ]]"
  assert_failure
}

assert_url() {
  url="$1"
  run /bin/bash -c "dokku url $TEST_APP | xargs"
  echo "VHOST: $(cat $DOKKU_ROOT/$TEST_APP/VHOST | xargs)"
  echo "tls: $(ls $DOKKU_ROOT/$TEST_APP/tls || true)"
  echo "proxy-is-enabled: $(dokku plugin:trigger proxy-is-enabled "$TEST_APP")"
  echo "port-map: $(dokku ports:report "$TEST_APP" --ports-map)"
  echo "url: $(dokku urls $TEST_APP)"
  echo "output: $output"
  echo "status: $status"
  assert_output "${url}"
}

assert_urls() {
  # shellcheck disable=SC2124
  urls="$@"
  run /bin/bash -c "dokku urls $TEST_APP | xargs"
  echo "VHOST: $(cat $DOKKU_ROOT/$TEST_APP/VHOST | xargs)"
  echo "tls: $(ls $DOKKU_ROOT/$TEST_APP/tls || true)"
  echo "proxy-is-enabled: $(dokku plugin:trigger proxy-is-enabled "$TEST_APP")"
  echo "port-map: $(dokku ports:report "$TEST_APP" --ports-map)"
  echo "urls: $(dokku urls $TEST_APP)"
  echo "output: $output"
  echo "status: $status"
  for url in "$@"; do
    assert_output_contains "$url"
  done
}

deploy_app() {
  declare APP_TYPE="$1" GIT_REMOTE="$2" CUSTOM_TEMPLATE="$3" CUSTOM_PATH="$4"
  local APP_TYPE=${APP_TYPE:="python"}
  local GIT_REMOTE=${GIT_REMOTE:="dokku@${DOKKU_DOMAIN}:$TEST_APP"}
  local GIT_REMOTE_BRANCH=${GIT_REMOTE_BRANCH:="master"}
  local TMP=${CUSTOM_TMP:=$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")}

  rmdir "$TMP" && cp -r "${BATS_TEST_DIRNAME}/../../tests/apps/$APP_TYPE" "$TMP"

  [[ -n "$CUSTOM_TEMPLATE" ]] && $CUSTOM_TEMPLATE $TEST_APP $TMP/$CUSTOM_PATH

  pushd "$TMP" &>/dev/null || exit 1
  [[ -z "$CUSTOM_TMP" ]] && trap 'popd &>/dev/null || true; rm -rf "$TMP"' RETURN INT TERM

  git config --global init.defaultBranch master
  git init
  git config user.email "robot@example.com"
  git config user.name "Test Robot"
  echo "setting up remote: $GIT_REMOTE"
  git remote add target "$GIT_REMOTE"

  [[ -f gitignore ]] && mv gitignore .gitignore
  git add .
  git commit -m 'initial commit'
  git push target "master:${GIT_REMOTE_BRANCH}" || destroy_app $?
}

setup_client_repo() {
  local TMP
  TMP=$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")
  rmdir "$TMP" && cp -r "${BATS_TEST_DIRNAME}/../../tests/apps/nodejs-express" "$TMP"
  cd "$TMP" || exit 1
  git init
  git config user.email "robot@example.com"
  git config user.name "Test Robot"

  [[ -f gitignore ]] && mv gitignore .gitignore
  git add .
  git commit -m 'initial commit'
}

setup_test_tls() {
  local TLS_TYPE="$1"
  local TLS="/home/dokku/$TEST_APP/tls"

  if ! dokku apps:exists "$TEST_APP" &>/dev/null; then
    create_app "$TEST_APP"
  fi
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

teardown_test_tls() {
  local TLS="/home/dokku/$TEST_APP/tls"
  rm -rf "$TLS"
}

custom_ssl_nginx_template() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mkdir -p "$APP_REPO_DIR"

  echo "injecting custom_ssl_nginx_template -> $APP_REPO_DIR/nginx.conf.sigil"
  cat <<EOF >"$APP_REPO_DIR/nginx.conf.sigil"
{{ range \$port_map := .PROXY_PORT_MAP | split " " }}
{{ \$port_map_list := \$port_map | split ":" }}
{{ \$scheme := index \$port_map_list 0 }}
{{ \$listen_port := index \$port_map_list 1 }}
{{ \$upstream_port := index \$port_map_list 2 }}
{{ if eq \$scheme "http" }}
server {
  listen      [::]:{{ \$listen_port }};
  listen      {{ \$listen_port }};
  server_name {{ $.NOSSL_SERVER_NAME }} $CUSTOM_TEMPLATE_SSL_DOMAIN;
  return 301 https://\$host:{{ $.PROXY_SSL_PORT }}\$request_uri;
}
{{ else if eq \$scheme "https"}}
server {
  listen      [::]:{{ $.PROXY_SSL_PORT }} ssl;
  listen      {{ $.PROXY_SSL_PORT }} ssl;
  {{ if $.SSL_SERVER_NAME }}server_name {{ $.SSL_SERVER_NAME }} $CUSTOM_TEMPLATE_SSL_DOMAIN; {{ end }}
  {{ if $.NOSSL_SERVER_NAME }}server_name {{ $.NOSSL_SERVER_NAME }} $CUSTOM_TEMPLATE_SSL_DOMAIN; {{ end }}
  ssl_certificate     {{ $.APP_SSL_PATH }}/server.crt;
  ssl_certificate_key {{ $.APP_SSL_PATH }}/server.key;

  keepalive_timeout   70;
  location    / {
    proxy_pass  http://{{ $.APP }}-{{ \$upstream_port }};
    proxy_http_version 1.1;
    proxy_set_header Upgrade \$http_upgrade;
    proxy_set_header Connection \$http_connection;
    proxy_set_header Host \$http_host;
    proxy_set_header X-Forwarded-Proto \$scheme;
    proxy_set_header X-Forwarded-For \$remote_addr;
    proxy_set_header X-Forwarded-Port \$server_port;
    proxy_set_header X-Request-Start \$msec;
  }
  include {{ $.DOKKU_ROOT }}/{{ $.APP }}/nginx.conf.d/*.conf;
}
{{ end }}{{ end }}

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

custom_nginx_template() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mkdir -p "$APP_REPO_DIR"

  echo "injecting custom_nginx_template -> $APP_REPO_DIR/nginx.conf.sigil"
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
{{ if $.DOKKU_APP_WORKER_LISTENERS }}
{{ range \$upstream_port := $.PROXY_UPSTREAM_PORTS | split " " }}
upstream {{ $.APP }}-worker-{{ \$upstream_port }} {
{{ range \$listeners := $.DOKKU_APP_WORKER_LISTENERS | split " " }}
  server {{ \$listeners }};{{ end }}
}
{{ end }}{{ end }}

EOF
  cat "$APP_REPO_DIR/nginx.conf.sigil"
}

bad_custom_nginx_template() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting bad_custom_nginx_template -> $APP_REPO_DIR/nginx.conf.sigil"
  cat <<EOF >"$APP_REPO_DIR/nginx.conf.sigil"
some lame nginx config
EOF
}

template_checks_file() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting templated CHECKS file -> $APP_REPO_DIR/CHECKS"
  cat <<EOF >"$APP_REPO_DIR/CHECKS"
WAIT=2 # wait 2 seconds
TIMEOUT=5 # set timeout to 5 seconds
ATTEMPTS=2 # try 2 times

{{ var "HEALTHCHECK_ENDPOINT" }} {{ var "HEALTHCHECK_ENDPOINT" }}
EOF
}

add_release_command() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "release: touch /app/release.test" >>"$APP_REPO_DIR/Procfile"
}

add_postdeploy_command() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  touch "$APP_REPO_DIR/app.json"
  contents=$(jq '.scripts.postdeploy = "touch /app/heroku-postdeploy.test"' "$APP_REPO_DIR/app.json")
  echo "${contents}" >"$APP_REPO_DIR/app.json"
}

move_dockerfile_into_place() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mv "$APP_REPO_DIR/alt.Dockerfile" "$APP_REPO_DIR/Dockerfile"
}

move_expose_dockerfile_into_place() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  cat "$APP_REPO_DIR/expose.Dockerfile"
  mv "$APP_REPO_DIR/expose.Dockerfile" "$APP_REPO_DIR/Dockerfile"
}

add_requirements_txt() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "flask" >>"$APP_REPO_DIR/requirements.txt"
}

add_requirements_txt_cnb() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"

  mv "$APP_REPO_DIR/app-cnb.json" "$APP_REPO_DIR/app.json"
  add_requirements_txt "$APP" "$APP_REPO_DIR"
}

create_network() {
  local NETWORK_NAME="${1:-$TEST_NETWORK}"

  NETWORK=$(docker network ls -q -f name="$NETWORK_NAME")
  [[ -z "$NETWORK" ]] && docker network create "$NETWORK_NAME"
}

attach_network() {
  local NETWORK_NAME="${1:-$TEST_NETWORK}"

  NETWORK=$(docker network ls -q -f name="$NETWORK_NAME")
  [[ -n "$NETWORK" ]] && docker network connect "$NETWORK_NAME" "${TEST_APP}.web.1"
}

create_attach_network() {
  local NETWORK_NAME="${1:-$TEST_NETWORK}"

  create_network "$NETWORK_NAME"
  attach_network "$NETWORK_NAME"
}

delete_network() {
  local NETWORK_NAME="${1:-$TEST_NETWORK}"

  NETWORK=$(docker network ls -q -f name="$NETWORK_NAME")
  [[ -n "$NETWORK" ]] && docker network rm "$NETWORK_NAME"
}

detach_network() {
  local NETWORK_NAME="${1:-$TEST_NETWORK}"

  NETWORK=$(docker network ls -q -f name="$NETWORK_NAME")
  [[ -z "$NETWORK" ]] && docker network disconnect "$NETWORK_NAME" "${TEST_APP}.web.1"
}

detach_delete_network() {
  local NETWORK_NAME="${1:-$TEST_NETWORK}"

  detach_network "$NETWORK_NAME"
  delete_network "$NETWORK_NAME"
}

convert_to_dockerfile() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  mv "$APP_REPO_DIR/dockerfile.Dockerfile" "$APP_REPO_DIR/Dockerfile"
}

install_pack() {
  if [[ ! -x /usr/bin/pack ]]; then
    add-apt-repository --yes ppa:cncf-buildpacks/pack-cli
    apt-get update
    apt-get --yes install pack-cli
  fi
}

install_nixpacks() {
  if ! command -v "nixpacks" &>/dev/null; then
    curl -sSL https://nixpacks.com/install.sh | FORCE=1 bash
  fi
}

install_railpack() {
  if ! command -v "railpack" &>/dev/null; then
    curl -sSL https://railpack.com/install.sh | FORCE=1 bash
  fi
}

install_k3s() {
  run /bin/bash -c "dokku proxy:set --global k3s"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set --global server hub.docker.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set --global image-repo-template '$DOCKERHUB_USERNAME/{{ .AppName }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set --global push-on-release true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku scheduler:set --global selected k3s"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:login docker.io $DOCKERHUB_USERNAME $DOCKERHUB_TOKEN"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Login Succeeded"

  INGRESS_CLASS="${INGRESS_CLASS:-traefik}"
  local args="--ingress-class $INGRESS_CLASS"
  if [[ "$TAINT_SCHEDULING" == "true" ]]; then
    args="$args --taint-scheduling"
  fi

  if [[ -n "$CI_SERVER_IP" ]]; then
    args="$args --server-ip $CI_SERVER_IP"
  fi

  run /bin/bash -c "dokku scheduler-k3s:initialize ${args}"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

uninstall_k3s() {
  run /bin/bash -c "dokku scheduler-k3s:uninstall"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
