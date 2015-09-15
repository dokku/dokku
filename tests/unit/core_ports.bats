#!/usr/bin/env bats

load test_helper

setup() {
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -f "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  [[ -f "$DOKKU_ROOT/HOSTNAME" ]] && cp -f "$DOKKU_ROOT/HOSTNAME" "$DOKKU_ROOT/HOSTNAME.bak"
  DOCKERFILE="$BATS_TMPDIR/Dockerfile"
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST"
  [[ -f "$DOKKU_ROOT/HOSTNAME.bak" ]] && mv "$DOKKU_ROOT/HOSTNAME.bak" "$DOKKU_ROOT/HOSTNAME"
}


check_urls() {
  local PATTERN="$1"
  run bash -c "dokku --quiet urls $TEST_APP | egrep \"${1}\""
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(core) port exposure (with global VHOST)" {
  echo "dokku.me" > "$DOKKU_ROOT/VHOST"
  deploy_app
  CONTAINER_ID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run bash -c "docker port $CONTAINER_ID | sed 's/[0-9.]*://' | egrep -q '[0-9]*'"
  echo "output: "$output
  echo "status: "$status
  assert_failure

  check_urls http://${TEST_APP}.dokku.me
}

@test "(core) port exposure (without global VHOST and real HOSTNAME)" {
  rm "$DOKKU_ROOT/VHOST"
  echo "${TEST_APP}.dokku.me" > "$DOKKU_ROOT/HOSTNAME"
  deploy_app
  CONTAINER_ID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run bash -c "docker port $CONTAINER_ID | sed 's/[0-9.]*://' | egrep -q '[0-9]*'"
  echo "output: "$output
  echo "status: "$status
  assert_success

  HOSTNAME=$(< "$DOKKU_ROOT/HOSTNAME")
  check_urls http://${HOSTNAME}:[0-9]+
}

@test "(core) port exposure (with NO_VHOST set)" {
  deploy_app
  dokku config:set $TEST_APP NO_VHOST=1
  CONTAINER_ID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run bash -c "docker port $CONTAINER_ID | sed 's/[0-9.]*://' | egrep -q '[0-9]*'"
  echo "output: "$output
  echo "status: "$status
  assert_success

  HOSTNAME=$(< "$DOKKU_ROOT/HOSTNAME")
  check_urls http://${HOSTNAME}:[0-9]+
}

@test "(core) port exposure (without global VHOST and IPv4 address as HOSTNAME)" {
  rm "$DOKKU_ROOT/VHOST"
  echo "127.0.0.1" > "$DOKKU_ROOT/HOSTNAME"
  deploy_app
  CONTAINER_ID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run bash -c "docker port $CONTAINER_ID | sed 's/[0-9.]*://' | egrep -q '[0-9]*'"
  echo "output: "$output
  echo "status: "$status
  assert_success

  HOSTNAME=$(< "$DOKKU_ROOT/HOSTNAME")
  check_urls http://${HOSTNAME}:[0-9]+
}

@test "(core) port exposure (without global VHOST and IPv6 address as HOSTNAME)" {
  rm "$DOKKU_ROOT/VHOST"
  echo "fda5:c7db:a520:bb6d::aabb:ccdd:eeff" > "$DOKKU_ROOT/HOSTNAME"
  deploy_app
  CONTAINER_ID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run bash -c "docker port $CONTAINER_ID | sed 's/[0-9.]*://' | egrep -q '[0-9]*'"
  echo "output: "$output
  echo "status: "$status
  assert_success

  HOSTNAME=$(< "$DOKKU_ROOT/HOSTNAME")
  check_urls http://${HOSTNAME}:[0-9]+
}

@test "(core) port exposure (pre-deploy domains:add)" {
  create_app
  run dokku domains:add $TEST_APP www.test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success

  deploy_app
  sleep 5 # wait for nginx to reload

  CONTAINER_ID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run bash -c "docker port $CONTAINER_ID | sed 's/[0-9.]*://' | egrep -q '[0-9]*'"
  echo "output: "$output
  echo "status: "$status
  assert_failure

  run bash -c "response=\"$(curl -s -S www.test.app.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success

  check_urls http://www.test.app.dokku.me
}

@test "(core) port exposure (no global VHOST and domains:add post deploy)" {
  rm "$DOKKU_ROOT/VHOST"
  deploy_app

  run dokku domains:add $TEST_APP www.test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success

  run dokku ps:restart $TEST_APP
  echo "output: "$output
  echo "status: "$status
  assert_success

  CONTAINER_ID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run bash -c "docker port $CONTAINER_ID | sed 's/[0-9.]*://' | egrep -q '[0-9]*'"
  echo "output: "$output
  echo "status: "$status
  assert_failure

  run bash -c "response=\"$(curl -s -S www.test.app.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success

  check_urls http://www.test.app.dokku.me
}

@test "(core) dockerfile port exposure" {
  deploy_app dockerfile
  run bash -c "grep -A1 upstream $DOKKU_ROOT/$TEST_APP/nginx.conf | grep -q 3000"
  echo "output: "$output
  echo "status: "$status
  assert_success

  check_urls http://${TEST_APP}.dokku.me
}

@test "(core) port exposure (xip.io style hostnames)" {
  echo "127.0.0.1.xip.io" > "$DOKKU_ROOT/VHOST"
  deploy_app

  run bash -c "response=\"$(curl -s -S my-cool-guy-test-app.127.0.0.1.xip.io)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success

  check_urls http://my-cool-guy-test-app.127.0.0.1.xip.io
}

@test "(core) port exposure (dockerfile raw port)" {
  source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
  cat<<EOF > $DOCKERFILE
EXPOSE 3001/udp
EXPOSE 3003
EXPOSE  3000/tcp
EOF
  run get_dockerfile_exposed_port $DOCKERFILE
  echo "output: "$output
  echo "status: "$status
  assert_output 3003
}

@test "(core) port exposure (dockerfile tcp port)" {
  source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
  cat<<EOF > $DOCKERFILE
EXPOSE 3001/udp
EXPOSE  3000/tcp
EXPOSE 3003
EOF
  run get_dockerfile_exposed_port $DOCKERFILE
  echo "output: "$output
  echo "status: "$status
  assert_output 3000
}
