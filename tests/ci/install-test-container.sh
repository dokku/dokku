#!/usr/bin/env bash
set -eo pipefail
[[ $TRACE ]] && set -x

readonly ROOT_DIR="$(cd "$(dirname "$(dirname "$(dirname "${BASH_SOURCE[0]}")")")" && pwd)"
readonly CONTAINER_NAME="dokku-install-test"

log-info() {
  declare desc="Log info formatter"
  echo "=====> $*"
}

log-fail() {
  declare desc="Log fail formatter"
  echo "!     $*" 1>&2
  exit 1
}

fn-build-image() {
  declare desc="Builds a systemd-enabled image from the given base image"
  declare BASE_IMAGE="$1" IMAGE="$2"

  log-info "Building $IMAGE from $BASE_IMAGE"
  docker build --progress plain -t "$IMAGE" - <<EOF
FROM $BASE_IMAGE
ENV container=docker
RUN apt-get update -qq \\
  && DEBIAN_FRONTEND=noninteractive apt-get -qq -y --no-install-recommends install ca-certificates curl dbus sudo systemd systemd-sysv \\
  && rm -rf /var/lib/apt/lists/*
STOPSIGNAL SIGRTMIN+3
CMD ["/sbin/init"]
EOF
}

fn-start-container() {
  declare desc="Boots the image with systemd as the init system"
  declare IMAGE="$1"

  log-info "Booting $IMAGE"
  docker run -d \
    --name "$CONTAINER_NAME" \
    --privileged \
    --cgroupns host \
    --tmpfs /run \
    --tmpfs /run/lock \
    -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
    -v /var/lib/docker \
    -v "$ROOT_DIR:/src" \
    "$IMAGE" >/dev/null
}

fn-wait-for-boot() {
  declare desc="Waits for systemd inside the container to finish booting"
  local attempt state

  for attempt in {1..60}; do
    docker exec "$CONTAINER_NAME" test -S /run/systemd/private && break
    sleep 1
  done

  state="$(docker exec "$CONTAINER_NAME" systemctl is-system-running --wait || true)"
  log-info "systemd state: $state"
  if [[ "$state" != "running" ]] && [[ "$state" != "degraded" ]]; then
    docker exec "$CONTAINER_NAME" systemctl --failed --no-pager || true
    log-fail "systemd failed to boot in $CONTAINER_NAME"
  fi
}

fn-print-diagnostics() {
  declare desc="Prints the docker daemon state inside the test container"
  log-info "docker service status"
  docker exec "$CONTAINER_NAME" systemctl status docker --no-pager || true
  log-info "docker service journal"
  docker exec "$CONTAINER_NAME" journalctl -u docker -u containerd --no-pager -n 100 || true
  log-info "docker info"
  docker exec "$CONTAINER_NAME" docker info || true
  log-info "docker pull hello-world"
  docker exec "$CONTAINER_NAME" docker pull hello-world || true
  log-info "install log"
  docker exec "$CONTAINER_NAME" bash -c 'tail -n 200 /var/log/dokku/install-*.log' || true
}

fn-cleanup() {
  declare desc="Prints diagnostics on failure and removes the test container"
  declare EXIT_CODE="$?"
  if [[ "$EXIT_CODE" -ne 0 ]] && docker container inspect "$CONTAINER_NAME" >/dev/null 2>/dev/null; then
    fn-print-diagnostics
  fi
  docker rm -f "$CONTAINER_NAME" >/dev/null || true
}

main() {
  declare desc="Runs the install test inside a systemd-enabled container for the given image"
  declare BASE_IMAGE="$1"
  local image

  [[ -n "$BASE_IMAGE" ]] || log-fail "Usage: install-test-container.sh <image>"
  image="dokku/install-test:${BASE_IMAGE//[:\/]/-}"

  trap fn-cleanup EXIT
  fn-build-image "$BASE_IMAGE" "$image"
  fn-start-container "$image"
  fn-wait-for-boot
  docker exec -w /src "$CONTAINER_NAME" ./tests/ci/install-test.sh
}

main "$@"
