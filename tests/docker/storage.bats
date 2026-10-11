#!/usr/bin/env bats

# Dokku-in-Docker coverage for DOKKU_LIB_HOST_ROOT. These tests talk to the
# `dokku` container started by tests/ci/setup.sh docker and read/write the
# host path that volume backs (/var/lib/dokku -> /mnt/dokku).

setup_file() {
  if ! docker inspect dokku >/dev/null 2>&1; then
    echo "dokku container is not running; skip docker storage tests" >&2
    return 1
  fi

  DOKKU_LIB_HOST_ROOT="$(docker exec dokku bash -c 'echo -n "$DOKKU_LIB_HOST_ROOT"')"
  if [[ -z "$DOKKU_LIB_HOST_ROOT" ]]; then
    echo "DOKKU_LIB_HOST_ROOT is unset inside the dokku container" >&2
    return 1
  fi
  export DOKKU_LIB_HOST_ROOT
}

setup() {
  ENTRY_NAME="rdmtest-docker-hostroot-$$"
  export ENTRY_NAME
}

teardown() {
  docker exec dokku dokku storage:destroy "$ENTRY_NAME" --force >/dev/null 2>&1 || true
  sudo rm -rf "${DOKKU_LIB_HOST_ROOT}/data/storage/${ENTRY_NAME}"
}

@test "(docker) storage:exec sees a file written under DOKKU_LIB_HOST_ROOT" {
  run docker exec dokku dokku storage:create "$ENTRY_NAME" --chown root
  echo "output: $output"
  echo "status: $status"
  [[ "$status" -eq 0 ]]

  # storage:create mkdirs under DOKKU_LIB_ROOT inside the container, which is
  # the same tree as DOKKU_LIB_HOST_ROOT on the host via the /mnt/dokku volume.
  HOST_STORAGE_DIR="${DOKKU_LIB_HOST_ROOT}/data/storage/${ENTRY_NAME}"
  [[ -d "$HOST_STORAGE_DIR" ]]

  echo "Hello World!" | sudo tee "${HOST_STORAGE_DIR}/welcome.txt" >/dev/null

  run docker exec dokku dokku storage:exec "$ENTRY_NAME" -- cat /data/welcome.txt
  echo "output: $output"
  echo "status: $status"
  [[ "$status" -eq 0 ]]
  [[ "$output" == *"Hello World!"* ]]
}
