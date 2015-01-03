#!/usr/bin/env bats

load test_helper

setup() {
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -f "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  [[ -f "$DOKKU_ROOT/HOSTNAME" ]] && cp -f "$DOKKU_ROOT/HOSTNAME" "$DOKKU_ROOT/HOSTNAME.bak"

}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST"
  [[ -f "$DOKKU_ROOT/HOSTNAME.bak" ]] && mv "$DOKKU_ROOT/HOSTNAME.bak" "$DOKKU_ROOT/HOSTNAME"
}

@test "port exposure (with global VHOST)" {
  echo "dokku.me" > "$DOKKU_ROOT/VHOST"
  deploy_app
  CONTAINER_ID=$(docker ps --no-trunc| grep dokku/$TEST_APP | grep "start web" | awk '{ print $1 }')
  run bash -c "docker port $CONTAINER_ID | sed 's/[0-9.]*://' | egrep '[0-9]*'"
  echo "output: "$output
  echo "status: "$status
  assert_failure
}

@test "port exposure (without global VHOST and real HOSTNAME)" {
  rm "$DOKKU_ROOT/VHOST"
  echo "dokku.me" > "$DOKKU_ROOT/HOSTNAME"
  deploy_app
  CONTAINER_ID=$(docker ps --no-trunc| grep dokku/$TEST_APP | grep "start web" | awk '{ print $1 }')
  run bash -c "docker port $CONTAINER_ID | sed 's/[0-9.]*://' | egrep '[0-9]*'"
  echo "output: "$output
  echo "status: "$status
  assert_failure
}

@test "port exposure (with NO_VHOST set)" {
  deploy_app
  dokku config:set $TEST_APP NO_VHOST=1
  CONTAINER_ID=$(docker ps --no-trunc| grep dokku/$TEST_APP | grep "start web" | awk '{ print $1 }')
  run bash -c "docker port $CONTAINER_ID | sed 's/[0-9.]*://' | egrep '[0-9]*'"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "port exposure (without global VHOST and ip as HOSTNAME)" {
  rm "$DOKKU_ROOT/VHOST"
  echo "127.0.0.1" > "$DOKKU_ROOT/HOSTNAME"
  deploy_app
  CONTAINER_ID=$(docker ps --no-trunc| grep dokku/$TEST_APP | grep "start web" | awk '{ print $1 }')
  run bash -c "docker port $CONTAINER_ID | sed 's/[0-9.]*://' | egrep '[0-9]*'"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
