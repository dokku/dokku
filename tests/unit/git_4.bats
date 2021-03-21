#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  touch /home/dokku/.ssh/known_hosts
  chown dokku:dokku /home/dokku/.ssh/known_hosts
}

teardown() {
  rm -f /home/dokku/.ssh/id_rsa.pub || true
  destroy_app
  global_teardown
}

@test "(git) git:from-image [missing]" {
  run /bin/bash -c "dokku git:from-image $TEST_APP dokku/python:missing-tag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(git) git:from-image [normal]" {
  run /bin/bash -c "dokku git:from-image $TEST_APP linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:from-image [normal-git-init]" {
  run rm -rf "/home/dokku/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run mkdir "/home/dokku/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run chown -R dokku:dokku "/home/dokku/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:from-image [normal-cnb]" {
  run /bin/bash -c "dokku git:from-image $TEST_APP dokku/node-js-getting-started:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:from-image [onbuild]" {
  local TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM

  run /bin/bash -c "dokku storage:mount $TEST_APP /var/run/docker.sock:/var/run/docker.sock"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP gliderlabs/logspout:v3.2.13"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  cat <<EOF >"$TMP/build.sh"
#!/bin/sh
set -e
apk add --update go build-base git mercurial ca-certificates
cd /src
go build -ldflags "-X main.Version=\$1" -o /bin/logspout
apk del go git mercurial build-base
rm -rf /root/go /var/cache/apk/*

# backwards compatibility
ln -fs /tmp/docker.sock /var/run/docker.sock
EOF

  cat <<EOF >"$TMP/modules.go"
package main

import (
  _ "github.com/gliderlabs/logspout/adapters/multiline"
  _ "github.com/gliderlabs/logspout/adapters/raw"
  _ "github.com/gliderlabs/logspout/adapters/syslog"
  _ "github.com/gliderlabs/logspout/healthcheck"
  _ "github.com/gliderlabs/logspout/httpstream"
  _ "github.com/gliderlabs/logspout/routesapi"
  _ "github.com/gliderlabs/logspout/transports/tcp"
  _ "github.com/gliderlabs/logspout/transports/tls"
  _ "github.com/gliderlabs/logspout/transports/udp"
)
EOF

  run sudo chown -R dokku:dokku "$TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image --build-dir $TMP $TEST_APP gliderlabs/logspout:v3.2.13"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP gliderlabs/logspout:v3.2.13"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
