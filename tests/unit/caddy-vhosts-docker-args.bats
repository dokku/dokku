#!/usr/bin/env bats

load test_helper

# Isolated coverage for plugins/caddy-vhosts/docker-args-process-deploy.
# Stubs plugn and plugin sources so the trigger can run without a Dokku install.

setup() {
  REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
  TRIGGER="$REPO_ROOT/plugins/caddy-vhosts/docker-args-process-deploy"
  STUB_ROOT="$BATS_TEST_TMPDIR/stub"
  mkdir -p "$STUB_ROOT/bin" \
    "$STUB_ROOT/core/common" \
    "$STUB_ROOT/plugins/proxy" \
    "$STUB_ROOT/plugins/caddy-vhosts"

  cat >"$STUB_ROOT/core/common/functions" <<'EOF'
#!/usr/bin/env bash
dokku_log_warn() { echo "WARN: $*" >&2; }
dokku_log_fail() {
  echo "FAIL: $*" >&2
  exit 1
}
EOF

  cat >"$STUB_ROOT/plugins/proxy/functions" <<EOF
#!/usr/bin/env bash
source "$REPO_ROOT/plugins/proxy/functions"
EOF
  cat >>"$STUB_ROOT/plugins/proxy/functions" <<'EOF'
fn-proxy-get-labels-file-path() {
  echo "${STUB_LABELS_FILE:-/nonexistent}"
}
EOF

  cat >"$STUB_ROOT/plugins/caddy-vhosts/internal-functions" <<'EOF'
#!/usr/bin/env bash
fn-caddy-computed-letsencrypt-email() {
  echo "${STUB_LETSENCRYPT_EMAIL:-}"
}
fn-caddy-computed-tls-internal() {
  echo "${STUB_TLS_INTERNAL:-false}"
}
EOF

  cat >"$STUB_ROOT/bin/plugn" <<'EOF'
#!/usr/bin/env bash
set -e
if [[ "$1" != "trigger" ]]; then
  echo "unexpected plugn command: $*" >&2
  exit 1
fi
shift
cmd="$1"
shift
case "$cmd" in
  proxy-type)
    echo "${STUB_PROXY_TYPE:-caddy}"
    ;;
  proxy-is-enabled)
    echo "${STUB_PROXY_ENABLED:-true}"
    ;;
  domains-vhost-enabled)
    [[ "${STUB_VHOST_ENABLED:-true}" == "true" ]]
    ;;
  ports-configure)
    ;;
  ports-get)
    if [[ -n "${STUB_PORTS:-}" ]]; then
      printf '%s\n' "$STUB_PORTS"
    fi
    ;;
  domains-list)
    if [[ -n "${STUB_DOMAINS:-}" ]]; then
      printf '%s\n' "$STUB_DOMAINS"
    fi
    ;;
  app-json-get-content)
    echo "${STUB_APP_JSON:-{\}}"
    ;;
  *)
    echo "unexpected plugn trigger: $cmd $*" >&2
    exit 1
    ;;
esac
EOF
  chmod +x "$STUB_ROOT/bin/plugn"
}

invoke_caddy_docker_args() {
  env \
    PLUGIN_CORE_AVAILABLE_PATH="$STUB_ROOT/core" \
    PLUGIN_AVAILABLE_PATH="$STUB_ROOT/plugins" \
    PATH="$STUB_ROOT/bin:$PATH" \
    STUB_PORTS="${STUB_PORTS:-}" \
    STUB_DOMAINS="${STUB_DOMAINS:-}" \
    STUB_TLS_INTERNAL="${STUB_TLS_INTERNAL:-false}" \
    STUB_LETSENCRYPT_EMAIL="${STUB_LETSENCRYPT_EMAIL:-}" \
    STUB_PROXY_TYPE="${STUB_PROXY_TYPE:-caddy}" \
    STUB_PROXY_ENABLED="${STUB_PROXY_ENABLED:-true}" \
    STUB_VHOST_ENABLED="${STUB_VHOST_ENABLED:-true}" \
    STUB_LABELS_FILE="${STUB_LABELS_FILE:-/nonexistent}" \
    STUB_APP_JSON="${STUB_APP_JSON:-}" \
    bash "$TRIGGER" "${APP_NAME:-testapp}" "dockerfile" "latest" "${PROC_TYPE:-web}" "1" <<<"${STDIN_DATA:-}"
}

@test "(caddy-vhosts) tls-internal does not leak caddy.tls=internal without http/https ports" {
  STUB_DOMAINS="example.com"
  STUB_PORTS=""
  STUB_TLS_INTERNAL="true"

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success

  if [[ "$output" == *"caddy.tls=internal"* ]]; then
    flunk "stray caddy.tls=internal label leaked with no http/https mapping: $output"
  fi
}

@test "(caddy-vhosts) is_app_listening gate compares == true not -n" {
  run grep -F '[[ "$is_app_listening" == "true" ]]' "$TRIGGER"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run grep -F '[[ -n "$is_app_listening" ]]' "$TRIGGER"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(caddy-vhosts) is_app_listening true still emits reverse_proxy labels" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="http:80:5000"
  STUB_TLS_INTERNAL="false"

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "caddy.reverse_proxy"
}

@test "(caddy-vhosts) readiness healthcheck emits health and retry labels" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="http:80:5000"
  STUB_APP_JSON='{"healthchecks":{"web":[{"type":"readiness","path":"/health","timeout":3,"wait":2}]}}'

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains '--label "caddy.reverse_proxy.health_uri=/health"'
  assert_output_contains "--label caddy.reverse_proxy.health_interval=2s"
  assert_output_contains "--label caddy.reverse_proxy.health_timeout=3s"
  assert_output_contains "--label caddy.reverse_proxy.health_headers.Host=example.com"
  assert_output_contains "--label caddy.reverse_proxy.lb_try_duration=5s"
  assert_output_contains "--label caddy.reverse_proxy.lb_try_interval=250ms"
  assert_output_contains "health_port" 0
}

@test "(caddy-vhosts) readiness healthcheck defaults and port" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="http:80:5000"
  STUB_APP_JSON='{"healthchecks":{"web":[{"type":"readiness","path":"/","port":5001}]}}'

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--label caddy.reverse_proxy.health_port=5001"
  assert_output_contains "--label caddy.reverse_proxy.health_interval=5s"
  assert_output_contains "--label caddy.reverse_proxy.health_timeout=5s"
}

@test "(caddy-vhosts) healthcheck host header skips localhost and catch-all domains" {
  STUB_DOMAINS="localhost * app.example.com other.example.com"
  STUB_PORTS="http:80:5000"
  STUB_APP_JSON='{"healthchecks":{"web":[{"type":"readiness","path":"/"}]}}'

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--label caddy.reverse_proxy.health_headers.Host=app.example.com"
}

@test "(caddy-vhosts) startup healthcheck emits no health or retry labels" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="http:80:5000"
  STUB_APP_JSON='{"healthchecks":{"web":[{"type":"startup","path":"/"}]}}'

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "caddy.reverse_proxy="
  assert_output_contains "health_" 0
  assert_output_contains "lb_try_" 0
}

@test "(caddy-vhosts) missing app.json emits no health or retry labels" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="http:80:5000"

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "caddy.reverse_proxy="
  assert_output_contains "health_" 0
  assert_output_contains "lb_try_" 0
}

@test "(caddy-vhosts) readiness healthcheck without routing labels emits no health labels" {
  STUB_DOMAINS="example.com"
  STUB_PORTS=""
  STUB_APP_JSON='{"healthchecks":{"web":[{"type":"readiness","path":"/"}]}}'

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "health_" 0
}

@test "(caddy-vhosts) invalid readiness healthcheck path fails" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="http:80:5000"
  STUB_APP_JSON='{"healthchecks":{"web":[{"type":"readiness","path":"/has space"}]}}'

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid app.json healthcheck path for caddy"
}
