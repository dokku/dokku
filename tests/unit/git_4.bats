#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  touch /home/dokku/.ssh/known_hosts
  chown dokku:dokku /home/dokku/.ssh/known_hosts
}

teardown() {
  release_deploy_lock
  docker image rm "dokku-test/$TEST_APP:second-procfile" "dokku-test/$TEST_APP:default-procfile" &>/dev/null || true
  rm -f /home/dokku/.ssh/id_rsa.pub || true
  destroy_app
  global_teardown
}

build_python_image() {
  declare TAG="$1" PROCFILE="$2"
  local TMP_IMAGE_DIR exit_code=0
  TMP_IMAGE_DIR="$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")"
  cp -r "${BATS_TEST_DIRNAME}/../../tests/apps/python/." "$TMP_IMAGE_DIR"
  if [[ -n "$PROCFILE" ]]; then
    cp "$TMP_IMAGE_DIR/$PROCFILE" "$TMP_IMAGE_DIR/Procfile"
  fi

  docker image build --label "com.dokku.test-app=$TEST_APP" -t "dokku-test/$TEST_APP:$TAG" -f "$TMP_IMAGE_DIR/alt.Dockerfile" "$TMP_IMAGE_DIR" || exit_code="$?"
  rm -rf "$TMP_IMAGE_DIR"
  return "$exit_code"
}

deploy_second_procfile_image() {
  run build_python_image second-procfile second.Procfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP dokku-test/$TEST_APP:second-procfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:trigger procfile-get-command $TEST_APP web 5001"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "second.Procfile"
}

deploy_branch_rev() {
  sudo -u dokku git -C "/home/dokku/$TEST_APP" rev-parse master
}

@test "(git) git:from-image [missing]" {
  run /bin/bash -c "dokku git:from-image $TEST_APP dokku/python:missing-tag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(git) git:from-image [normal]" {
  run /bin/bash -c "dokku git:from-image $TEST_APP linuxserver/foldingathome:7.6.21"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/HEAD"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ref: refs/heads/master"
}

@test "(git) git:from-image --force" {
  run /bin/bash -c "docker image rm linuxserver/foldingathome:7.6.21 2>/dev/null || true"

  run /bin/bash -c "dokku git:from-image $TEST_APP linuxserver/foldingathome:7.6.21"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Pulling image"

  run /bin/bash -c "dokku git:from-image $TEST_APP linuxserver/foldingathome:7.6.21"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Image exists on host, skipping pull"
  assert_output_contains "No changes detected, skipping git commit"

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku git:from-image --force $TEST_APP linuxserver/foldingathome:7.6.21"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Force pulling image"
  assert_output_contains "No changes detected, skipping git commit"

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(git) git:from-image [normal-custom-branch]" {
  run /bin/bash -c "dokku git:set $TEST_APP deploy-branch main"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP linuxserver/foldingathome:7.6.21"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/HEAD"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ref: refs/heads/main"
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

  run /bin/bash -c "dokku git:from-image $TEST_APP linuxserver/foldingathome:7.6.21"
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

@test "(git) git:from-image [failing deploy]" {
  local CUSTOM_TMP=$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")
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

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP dokku-test/$TEST_APP:latest"
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

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP dokku-test/$TEST_APP:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku-test/$TEST_APP:latest"

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP dokku-test/$TEST_APP:v2"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku-test/$TEST_APP:latest"

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP dokku-test/$TEST_APP:v2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "No changes detected, skipping git commit" 0

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku-test/$TEST_APP:v2"

  run /bin/bash -c "docker image rm dokku-test/$TEST_APP:latest dokku-test/$TEST_APP:v2"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:from-image [onbuild]" {
  local TMP=$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")
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

@test "(git) git:from-image labels correctly" {
  run /bin/bash -c "dokku git:from-image $TEST_APP linuxserver/foldingathome:7.6.21"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image inspect dokku/$TEST_APP:latest --format '{{ index .Config.Labels \"com.dokku.docker-image-labeler/alternate-tags\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "linuxserver/foldingathome:7.6.21"
}

@test "(git) git:from-image source-image after failed deploy and ps:rebuild" {
  deploy_second_procfile_image

  run build_python_image default-procfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP dokku-test/$TEST_APP:default-procfile"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku-test/$TEST_APP:second-procfile"

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku-test/$TEST_APP:second-procfile"

  run /bin/bash -c "dokku plugin:trigger procfile-get-command $TEST_APP web 5001"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "second.Procfile"
  assert_output_contains "first.Procfile" 0
}

@test "(git) git:from-image then git:from-archive ignores the previous source-image" {
  deploy_second_procfile_image

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "tar -C ${BATS_TEST_DIRNAME}/../../tests/apps -cf - python | dokku git:from-archive $TEST_APP --"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku-test/$TEST_APP:second-procfile"

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP FAIL_ON_STARTUP=false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "tar -C ${BATS_TEST_DIRNAME}/../../tests/apps -cf - python | dokku git:from-archive $TEST_APP --"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku plugin:trigger procfile-get-command $TEST_APP web 5001"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "first.Procfile"
  assert_output_contains "second.Procfile" 0
}

@test "(git) git:from-image then git push ignores the previous source-image" {
  local TMP=$(mktemp -d "/tmp/${DOKKU_DOMAIN}.XXXXX")
  trap 'rm -rf "$TMP"' INT TERM

  deploy_second_procfile_image

  rmdir "$TMP" && cp -r "${BATS_TEST_DIRNAME}/../../tests/apps/python" "$TMP"
  run /bin/bash -c "cd $TMP && git init -q && git checkout -q -b master && git config user.email 'robot@example.com' && git config user.name 'Test Robot' && git add . && git commit -q -m 'initial commit' && git push --force dokku@${DOKKU_DOMAIN}:$TEST_APP master:master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  rm -rf "$TMP"

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku plugin:trigger procfile-get-command $TEST_APP web 5001"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "first.Procfile"
  assert_output_contains "second.Procfile" 0
}

@test "(git) git:from-image then git:sync clears the source-image" {
  deploy_second_procfile_image

  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 2.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku git:from-image $TEST_APP dokku-test/$TEST_APP:second-procfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku-test/$TEST_APP:second-procfile"

  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git 2.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""
}

@test "(git) source-image is unchanged when the deploy lock is held" {
  deploy_second_procfile_image

  run build_python_image default-procfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  local DEPLOY_REV
  DEPLOY_REV="$(deploy_branch_rev)"
  echo "deploy branch rev: $DEPLOY_REV"

  hold_deploy_lock

  run /bin/bash -c "dokku git:from-image $TEST_APP dokku-test/$TEST_APP:default-procfile"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "currently has a deploy lock in place"

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku-test/$TEST_APP:second-procfile"

  run deploy_branch_rev
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$DEPLOY_REV"

  run /bin/bash -c "cat $DEPLOY_LOCK_FILE"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test-lock-holder"

  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 2.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "currently has a deploy lock in place"

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku-test/$TEST_APP:second-procfile"

  run deploy_branch_rev
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$DEPLOY_REV"

  run /bin/bash -c "cat $DEPLOY_LOCK_FILE"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test-lock-holder"

  release_deploy_lock

  run /bin/bash -c "dokku git:from-image $TEST_APP dokku-test/$TEST_APP:default-procfile"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "No changes detected, skipping git commit" 0

  run /bin/bash -c "dokku git:report $TEST_APP --git-source-image"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku-test/$TEST_APP:default-procfile"
}
