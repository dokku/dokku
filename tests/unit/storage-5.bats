#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  rm -rf "${DOKKU_LIB_ROOT}"/data/storage/rdmtest*
}

teardown() {
  destroy_app
  global_teardown
}

@test "(storage) without DOKKU_LIB_HOST_ROOT docker-args and storage:exec use DOKKU_LIB_ROOT" {
  unset DOKKU_LIB_HOST_ROOT

  run /bin/bash -c "dokku storage:create rdmtest-bare"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-bare --container-dir /data"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo '' | dokku plugin:trigger docker-args-process-deploy $TEST_APP herokuish latest web | grep -o -- '-v [^ ]*'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "-v ${DOKKU_LIB_ROOT}/data/storage/rdmtest-bare:/data"

  run /bin/bash -c "docker pull alpine:3 >/dev/null"
  assert_success

  run /bin/bash -c "dokku storage:exec rdmtest-bare -- /bin/sh -c 'echo bare-metal > /data/marker && cat /data/marker'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "bare-metal"

  run /bin/bash -c "cat ${DOKKU_LIB_ROOT}/data/storage/rdmtest-bare/marker"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "bare-metal"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-bare --force"
  assert_success
}

@test "(storage) with DOKKU_LIB_HOST_ROOT docker-args rewrites the mount host path" {
  HOST_ROOT="$(mktemp -d /tmp/dokku-lib-host.XXXXXX)"
  export DOKKU_LIB_HOST_ROOT="$HOST_ROOT"

  run /bin/bash -c "dokku storage:create rdmtest-hostroot"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Create still lands under DOKKU_LIB_ROOT inside Dokku.
  run /bin/bash -c "test -d ${DOKKU_LIB_ROOT}/data/storage/rdmtest-hostroot"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-hostroot --container-dir /data"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo '' | dokku plugin:trigger docker-args-process-deploy $TEST_APP herokuish latest web | grep -o -- '-v [^ ]*'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "-v ${HOST_ROOT}/data/storage/rdmtest-hostroot:/data"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-hostroot --force"
  assert_success
  rm -rf "$HOST_ROOT"
}

@test "(storage) with DOKKU_LIB_HOST_ROOT storage:exec mounts the host-root path" {
  # Simulate Dokku-in-Docker: filesystem ops use DOKKU_LIB_ROOT, but Docker -v
  # mounts must resolve under DOKKU_LIB_HOST_ROOT.
  HOST_ROOT="$(mktemp -d /tmp/dokku-lib-host.XXXXXX)"
  export DOKKU_LIB_HOST_ROOT="$HOST_ROOT"

  run /bin/bash -c "dokku storage:create rdmtest-hostroot-exec --chown root"
  echo "output: $output"
  echo "status: $status"
  assert_success

  mkdir -p "${HOST_ROOT}/data/storage/rdmtest-hostroot-exec"
  echo "Hello World!" >"${HOST_ROOT}/data/storage/rdmtest-hostroot-exec/welcome.txt"

  run /bin/bash -c "docker pull alpine:3 >/dev/null"
  assert_success

  run /bin/bash -c "dokku storage:exec rdmtest-hostroot-exec -- cat /data/welcome.txt"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Hello World!"

  # The in-container create path must not have received the host-written file.
  run /bin/bash -c "test ! -f ${DOKKU_LIB_ROOT}/data/storage/rdmtest-hostroot-exec/welcome.txt"
  assert_success

  run /bin/bash -c "dokku storage:destroy rdmtest-hostroot-exec --force"
  assert_success
  rm -rf "$HOST_ROOT"
}
