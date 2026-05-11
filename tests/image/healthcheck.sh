#!/usr/bin/env bash
set -euo pipefail

IMAGE_TAG="${IMAGE_TAG:-dokku/dokku:latest}"
CONTAINER_NAME="${CONTAINER_NAME:-dokku-healthcheck-smoke}"
START_TIMEOUT_SECONDS="${START_TIMEOUT_SECONDS:-360}"
UNHEALTHY_TIMEOUT_SECONDS="${UNHEALTHY_TIMEOUT_SECONDS:-180}"

log-info() {
  echo "-----> $*"
}

log-fail() {
  echo " !     $*" >&2
  exit 1
}

cleanup() {
  docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
}
trap cleanup EXIT

require-image() {
  if ! docker image inspect "$IMAGE_TAG" >/dev/null 2>&1; then
    log-fail "Image '$IMAGE_TAG' not found in local docker daemon. Build it first (eg. 'docker build -t $IMAGE_TAG .')."
  fi
}

start-container() {
  log-info "Starting container '$CONTAINER_NAME' from image '$IMAGE_TAG'"
  docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
  docker run -d \
    --name "$CONTAINER_NAME" \
    --env DOKKU_HOSTNAME=dokku.me \
    -v /var/run/docker.sock:/var/run/docker.sock \
    "$IMAGE_TAG" >/dev/null
}

wait-for-catchall-vhost() {
  local deadline=$((SECONDS + 30))
  while ((SECONDS < deadline)); do
    if docker exec "$CONTAINER_NAME" test -f /etc/nginx/conf.d/00-default-vhost.conf; then
      return 0
    fi
    sleep 1
  done

  log-fail "Catch-all default vhost /etc/nginx/conf.d/00-default-vhost.conf was never created. \
Likely cause: a stale build/package/dokku-*.deb that predates the catch-all template shipped in \
plugins/nginx-vhosts/templates/default-site.conf. Rebuild the deb before running this test."
}

wait-for-healthy() {
  local deadline=$((SECONDS + START_TIMEOUT_SECONDS))
  local status
  while ((SECONDS < deadline)); do
    status=$(docker inspect --format='{{.State.Health.Status}}' "$CONTAINER_NAME" 2>/dev/null || echo "missing")
    if [[ "$status" == "healthy" ]]; then
      log-info "Container reached healthy status"
      return 0
    fi
    sleep 5
  done

  log-info "Container failed to become healthy within ${START_TIMEOUT_SECONDS}s. Last logs:"
  docker logs --tail 100 "$CONTAINER_NAME" || true
  log-fail "Container '$CONTAINER_NAME' did not reach healthy status (last status: ${status:-unknown})"
}

assert-endpoint-ok() {
  log-info "Verifying /_dokku/health returns ok from inside the container"
  local body
  body=$(docker exec "$CONTAINER_NAME" curl -fsS http://127.0.0.1:18080/_dokku/health)
  if [[ "$body" != "ok" ]]; then
    log-fail "Expected health endpoint body 'ok', got: ${body}"
  fi
}

assert-port-not-published() {
  log-info "Verifying port 18080 is not published to the host"
  local ports
  ports=$(docker inspect --format='{{json .NetworkSettings.Ports}}' "$CONTAINER_NAME")
  if echo "$ports" | grep -q "18080/tcp"; then
    log-fail "Port 18080/tcp should not be published to the host. Inspect output: $ports"
  fi

  if curl --max-time 2 -fsS http://127.0.0.1:18080/_dokku/health >/dev/null 2>&1; then
    log-fail "Health endpoint is reachable from the host on 127.0.0.1:18080 - it must be loopback-only inside the container."
  fi
}

assert-unhealthy-after-sentinel-removed() {
  log-info "Removing /var/run/dokku/ready and confirming endpoint returns failure"
  docker exec "$CONTAINER_NAME" rm -f /var/run/dokku/ready

  local exit_code=0
  docker exec "$CONTAINER_NAME" curl -fsS http://127.0.0.1:18080/_dokku/health >/dev/null 2>&1 || exit_code=$?
  if ((exit_code == 0)); then
    log-fail "Health endpoint still returned success after the readiness sentinel was removed"
  fi

  log-info "Waiting for docker to flip the health status to unhealthy"
  local deadline=$((SECONDS + UNHEALTHY_TIMEOUT_SECONDS))
  local status
  while ((SECONDS < deadline)); do
    status=$(docker inspect --format='{{.State.Health.Status}}' "$CONTAINER_NAME" 2>/dev/null || echo "missing")
    if [[ "$status" == "unhealthy" ]]; then
      log-info "Container flipped to unhealthy"
      return 0
    fi
    sleep 5
  done

  log-fail "Container did not flip to unhealthy within ${UNHEALTHY_TIMEOUT_SECONDS}s (last status: ${status:-unknown})"
}

main() {
  require-image
  start-container
  wait-for-catchall-vhost
  wait-for-healthy
  assert-endpoint-ok
  assert-port-not-published
  assert-unhealthy-after-sentinel-removed
  log-info "All healthcheck smoke tests passed."
}

main "$@"
