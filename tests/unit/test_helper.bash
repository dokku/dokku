#!/usr/bin/env bash

# constants
DOKKU_ROOT=${DOKKU_ROOT:=~dokku}
DOKKU_LIB_ROOT=${DOKKU_LIB_PATH:="/var/lib/dokku"}
PLUGIN_PATH=${PLUGIN_PATH:="$DOKKU_LIB_ROOT/plugins"}
PLUGIN_AVAILABLE_PATH=${PLUGIN_AVAILABLE_PATH:="$PLUGIN_PATH/available"}
PLUGIN_ENABLED_PATH=${PLUGIN_ENABLED_PATH:="$PLUGIN_PATH/enabled"}
PLUGIN_CORE_PATH=${PLUGIN_CORE_PATH:="$DOKKU_LIB_ROOT/core-plugins"}
PLUGIN_CORE_AVAILABLE_PATH=${PLUGIN_CORE_AVAILABLE_PATH:="$PLUGIN_CORE_PATH/available"}
TEST_APP=my-cool-guy-test-app

# test functions
flunk() {
  { if [ "$#" -eq 0 ]; then cat -
    else echo "$*"
    fi
  }
  return 1
}

assert_success() {
  if [ "$status" -ne 0 ]; then
    flunk "command failed with exit status $status"
  elif [ "$#" -gt 0 ]; then
    assert_output "$1"
  fi
}

assert_failure() {
  if [ "$status" -eq 0 ]; then
    flunk "expected failed exit status"
  elif [ "$#" -gt 0 ]; then
    assert_output "$1"
  fi
}

assert_equal() {
  if [ "$1" != "$2" ]; then
    { echo "expected: $1"
      echo "actual:   $2"
    } | flunk
  fi
}

assert_output() {
  local expected
  if [ $# -eq 0 ]; then expected="$(cat -)"
  else expected="$1"
  fi
  assert_equal "$expected" "$output"
}

assert_line() {
  if [ "$1" -ge 0 ] 2>/dev/null; then
    assert_equal "$2" "${lines[$1]}"
  else
    local line
    for line in "${lines[@]}"; do
      if [ "$line" = "$1" ]; then return 0; fi
    done
    flunk "expected line \`$1'"
  fi
}

refute_line() {
  if [ "$1" -ge 0 ] 2>/dev/null; then
    local num_lines="${#lines[@]}"
    if [ "$1" -lt "$num_lines" ]; then
      flunk "output has $num_lines lines"
    fi
  else
    local line
    for line in "${lines[@]}"; do
      if [ "$line" = "$1" ]; then
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
  dokku apps:create $TEST_APP
}

destroy_app() {
  local RC="$1"; local RC=${RC:=0}
  local TEST_APP="$2"; local TEST_APP=${TEST_APP:=my-cool-guy-test-app}
  echo $TEST_APP | dokku apps:destroy $TEST_APP
  return $RC
}

add_domain() {
  dokku domains:add $TEST_APP $1
}

deploy_app() {
  APP_TYPE="$1"; APP_TYPE=${APP_TYPE:="nodejs-express"}
  GIT_REMOTE="$2"; GIT_REMOTE=${GIT_REMOTE:="dokku@dokku.me:$TEST_APP"}
  TMP=$(mktemp -d -t "dokku.me.XXXXX")
  rmdir $TMP && cp -r ./tests/apps/$APP_TYPE $TMP
  cd $TMP
  git init
  git config user.email "robot@example.com"
  git config user.name "Test Robot"
  echo "setting up remote: $GIT_REMOTE"
  git remote add target $GIT_REMOTE

  [[ -f gitignore ]] && mv gitignore .gitignore
  git add .
  git commit -m 'initial commit'
  git push target master || destroy_app $?
}

setup_client_repo() {
  TMP=$(mktemp -d -t "dokku.me.XXXXX")
  rmdir $TMP && cp -r ./tests/apps/nodejs-express $TMP
  cd $TMP
  git init
  git config user.email "robot@example.com"
  git config user.name "Test Robot"

  [[ -f gitignore ]] && mv gitignore .gitignore
  git add .
  git commit -m 'initial commit'
}

setup_test_tls() {
  TLS="/home/dokku/$TEST_APP/tls"
  mkdir -p $TLS
  tar xf $BATS_TEST_DIRNAME/server_ssl.tar -C $TLS
  sudo chown -R dokku:dokku $TLS
}

setup_test_tls_with_sans() {
  TLS="/home/dokku/$TEST_APP/tls"
  mkdir -p $TLS
  tar xf $BATS_TEST_DIRNAME/server_ssl_sans.tar -C $TLS
  sudo chown -R dokku:dokku $TLS
}

setup_test_tls_wildcard() {
  TLS="/home/dokku/tls"
  mkdir -p $TLS
  tar xf $BATS_TEST_DIRNAME/server_ssl_wildcard.tar -C $TLS
  sudo chown -R dokku:dokku $TLS
  sed -i -e "s:^# ssl_certificate $DOKKU_ROOT/tls/server.crt;:ssl_certificate $DOKKU_ROOT/tls/server.crt;:g" \
         -e "s:^# ssl_certificate_key $DOKKU_ROOT/tls/server.key;:ssl_certificate_key $DOKKU_ROOT/tls/server.key;:g" /etc/nginx/conf.d/dokku.conf
  kill -HUP "$(< /var/run/nginx.pid)"; sleep 5
}

disable_tls_wildcard() {
  TLS="/home/dokku/tls"
  rm -rf $TLS
  sed -i -e "s:^ssl_certificate $DOKKU_ROOT/tls/server.crt;:# ssl_certificate $DOKKU_ROOT/tls/server.crt;:g" \
         -e "s:^ssl_certificate_key $DOKKU_ROOT/tls/server.key;:# ssl_certificate_key $DOKKU_ROOT/tls/server.key;:g" /etc/nginx/conf.d/dokku.conf
  kill -HUP "$(< /var/run/nginx.pid)"; sleep 5
}

custom_ssl_nginx_template() {
  APP="$1"
  [[ -z "$APP" ]] && APP="$TEST_APP"
cat<<EOF > $DOKKU_ROOT/$APP/nginx.conf.template
server {
  listen      [::]:80;
  listen      80;
  server_name \$NOSSL_SERVER_NAME;
  return 301 https://\$SSL_SERVER_NAME\\\$request_uri;
}

server {
  listen      [::]:443 ssl spdy;
  listen      443 ssl spdy;
  server_name \$SSL_SERVER_NAME;
\$SSL_DIRECTIVES

  keepalive_timeout   70;
  add_header          Alternate-Protocol  443:npn-spdy/2;
  location    / {
    proxy_pass  http://\$APP;
    proxy_http_version 1.1;
    proxy_set_header Upgrade \\\$http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host \\\$http_host;
    proxy_set_header X-Forwarded-Proto \\\$scheme;
    proxy_set_header X-Forwarded-For \\\$remote_addr;
    proxy_set_header X-Forwarded-Port \\\$server_port;
    proxy_set_header X-Request-Start \\\$msec;
  }
  include \$DOKKU_ROOT/\$APP/nginx.conf.d/*.conf;
}
EOF
}

custom_nginx_template() {
  APP="$1"
  [[ -z "$APP" ]] && APP="$TEST_APP"
cat<<EOF > $DOKKU_ROOT/$APP/nginx.conf.template
server {
  listen      [::]:80;
  listen      80;
  server_name \$NOSSL_SERVER_NAME;

  location    / {
    proxy_pass  http://\$APP;
    proxy_http_version 1.1;
    proxy_set_header Upgrade \\\$http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host \\\$http_host;
    proxy_set_header X-Forwarded-Proto \\\$scheme;
    proxy_set_header X-Forwarded-For \\\$remote_addr;
    proxy_set_header X-Forwarded-Port \\\$server_port;
    proxy_set_header X-Request-Start \\\$msec;
  }
  include \$DOKKU_ROOT/\$APP/nginx.conf.d/*.conf;
}
EOF
}
