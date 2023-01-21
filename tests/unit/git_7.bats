#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  touch /home/dokku/.ssh/known_hosts
  chown dokku:dokku /home/dokku/.ssh/known_hosts
}

teardown() {
  docker image rm linuxserver/foldingathome:7.5.1-ls1 || true
  docker image rm dokku/node-js-getting-started:latest || true
  docker image rm dokku-test/$TEST_APP:latest || true
  docker image rm dokku-test/$TEST_APP:v2 || true
  docker image rm gliderlabs/logspout:v3.2.13 || true
  rm -f /tmp/image.tar /tmp/image-2.tar
  rm -f /home/dokku/.ssh/id_rsa.pub || true
  destroy_app
  global_teardown
}

@test "(git) git:load-image [normal]" {
  run /bin/bash -c "docker image pull linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image save -o /tmp/image.tar linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image rm linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /tmp/image.tar | dokku git:load-image $TEST_APP linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:load-image [normal-git-init]" {
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

  run /bin/bash -c "docker image pull linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image save -o /tmp/image.tar linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image rm linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /tmp/image.tar | dokku git:load-image $TEST_APP linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:load-image [normal-cnb]" {
  run /bin/bash -c "docker image pull dokku/node-js-getting-started:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image save -o /tmp/image.tar dokku/node-js-getting-started:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image rm dokku/node-js-getting-started:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /tmp/image.tar | dokku git:load-image $TEST_APP dokku/node-js-getting-started:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:load-image [failing deploy]" {
  local CUSTOM_TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$CUSTOM_TMP"' INT TERM
  rmdir "$CUSTOM_TMP" && cp -r "${BATS_TEST_DIRNAME}/../../tests/apps/python" "$CUSTOM_TMP"

  run /bin/bash -c "docker image build -t dokku-test/$TEST_APP:latest -f $CUSTOM_TMP/alt.Dockerfile $CUSTOM_TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image build -t dokku-test/$TEST_APP:v2 --build-arg BUILD_ARG=value -f $CUSTOM_TMP/alt.Dockerfile $CUSTOM_TMP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image save -o /tmp/image.tar dokku-test/$TEST_APP:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image save -o /tmp/image-2.tar dokku-test/$TEST_APP:v2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image rm dokku-test/$TEST_APP:latest dokku-test/$TEST_APP:v2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /tmp/image.tar | dokku git:load-image $TEST_APP dokku-test/$TEST_APP:latest"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku config:get $TEST_APP FAIL_ON_STARTUP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku git:status $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output "fatal: this operation must be run in a work tree"

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /tmp/image.tar | dokku git:load-image $TEST_APP dokku-test/$TEST_APP:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /tmp/image-2.tar | dokku git:load-image $TEST_APP dokku-test/$TEST_APP:v2"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /tmp/image-2.tar | dokku git:load-image $TEST_APP dokku-test/$TEST_APP:v2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "No changes detected, skipping git commit" 0
}

@test "(git) git:load-image [onbuild]" {
  local TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM

  run /bin/bash -c "dokku storage:mount $TEST_APP /var/run/docker.sock:/var/run/docker.sock"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image pull gliderlabs/logspout:v3.2.13"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image save -o /tmp/image-2.tar gliderlabs/logspout:v3.2.13"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image rm gliderlabs/logspout:v3.2.13"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /tmp/image.tar | dokku git:load-image $TEST_APP gliderlabs/logspout:v3.2.13"
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

  run /bin/bash -c "cat /tmp/image.tar | dokku git:load-image --build-dir $TMP $TEST_APP gliderlabs/logspout:v3.2.13"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /tmp/image.tar | dokku git:load-image $TEST_APP gliderlabs/logspout:v3.2.13"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(git) git:load-image labels correctly" {
  run /bin/bash -c "docker image pull linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image save -o /tmp/image.tar linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image rm linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /tmp/image.tar | dokku git:load-image $TEST_APP linuxserver/foldingathome:7.5.1-ls1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image inspect dokku/$TEST_APP:latest --format '{{ index .Config.Labels \"com.dokku.docker-image-labeler/alternate-tags\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "linuxserver/foldingathome:7.5.1-ls1"
}
