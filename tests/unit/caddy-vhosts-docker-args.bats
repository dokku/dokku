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

  touch "$STUB_ROOT/core/common/property-functions"

  cat >"$STUB_ROOT/plugins/caddy-vhosts/internal-functions" <<EOF
#!/usr/bin/env bash
source "$REPO_ROOT/plugins/caddy-vhosts/internal-functions"
EOF
  cat >>"$STUB_ROOT/plugins/caddy-vhosts/internal-functions" <<'EOF'
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
  proxy-routes-list)
    ;;
  proxy-redirects-list)
    if [[ -n "${STUB_REDIRECTS:-}" ]]; then
      printf '%s\n' "$STUB_REDIRECTS"
    fi
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
    STUB_REDIRECTS="${STUB_REDIRECTS:-}" \
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

@test "(caddy-vhosts) readiness healthcheck maps attempts and headers" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="http:80:5000"
  STUB_APP_JSON='{"healthchecks":{"web":[{"type":"readiness","path":"/","attempts":4,"httpHeaders":[{"name":"X-Check","value":"dokku check"},{"name":"host","value":"check.example.com"}]}]}}'

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--label caddy.reverse_proxy.health_fails=4"
  assert_output_contains "--label 'caddy.reverse_proxy.health_headers.X-Check=\"dokku check\"'"
  assert_output_contains "--label caddy.reverse_proxy.health_headers.Host=check.example.com"
  assert_output_contains "health_headers.Host=example.com" 0
  assert_output_contains "health_headers.host" 0
}

@test "(caddy-vhosts) readiness healthcheck defaults attempts to 3" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="http:80:5000"
  STUB_APP_JSON='{"healthchecks":{"web":[{"type":"readiness","path":"/"}]}}'

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--label caddy.reverse_proxy.health_fails=3"
}

@test "(caddy-vhosts) invalid healthcheck headers are skipped" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="http:80:5000"
  STUB_APP_JSON='{"healthchecks":{"web":[{"type":"readiness","path":"/","httpHeaders":[{"name":"X.Dotted","value":"a"},{"name":"X-Order_1","value":"a"},{"name":"X-Quote","value":"it'"'"'s"},{"name":"Bad Name","value":"a"},{"name":"X-Good","value":"ok"}]}]}}'

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Skipping invalid app.json healthcheck header for caddy" 4
  assert_output_contains "health_headers.X-Good=" 1
  assert_output_contains "health_headers.X.Dotted" 0
  assert_output_contains "health_headers.X-Order_1" 0
  assert_output_contains "health_headers.X-Quote" 0
}

@test "(caddy-vhosts) tls-internal emits caddy.tls=internal alongside https site labels" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="$(printf 'http:80:5000\nhttps:443:5000')"
  STUB_LETSENCRYPT_EMAIL="admin@example.com"
  STUB_TLS_INTERNAL="true"

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--label 'caddy=example.com'"
  assert_output_contains "--label caddy.tls=internal"
  assert_output_contains "caddy.reverse_proxy="
}

@test "(caddy-vhosts) tls-internal does not emit caddy.tls=internal for http-only sites" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="http:80:5000"
  STUB_TLS_INTERNAL="true"

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--label 'caddy=example.com:80'"
  assert_output_contains "caddy.tls=internal" 0
}

@test "(caddy-vhosts) redirects emit http redirect site labels and drop the source from the app site" {
  STUB_DOMAINS="$(printf 'example.com\nwww.example.com')"
  STUB_PORTS="http:80:5000"
  STUB_REDIRECTS="$(printf 'old.example.com|example.com|302\nwww.example.com|example.com|301')"

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--label 'caddy=example.com:80'"
  assert_output_contains "--label 'caddy_1=old.example.com:80'"
  assert_output_contains "--label 'caddy_1.redir=http://example.com{uri} 302'"
  assert_output_contains "--label 'caddy_2=www.example.com:80'"
  assert_output_contains "--label 'caddy_2.redir=http://example.com{uri} 301'"
}

@test "(caddy-vhosts) redirects use https sites when a letsencrypt email is set" {
  STUB_DOMAINS="example.com"
  STUB_PORTS="$(printf 'http:80:5000\nhttps:443:5000')"
  STUB_LETSENCRYPT_EMAIL="admin@example.com"
  STUB_REDIRECTS="www.example.com|example.com|308"

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--label 'caddy=example.com'"
  assert_output_contains "--label 'caddy_1=www.example.com'"
  assert_output_contains "--label 'caddy_1.redir=https://example.com{uri} 308'"
}

@test "(caddy-vhosts) redirects still render when every app domain is a redirect source" {
  STUB_DOMAINS="www.example.com"
  STUB_PORTS="http:80:5000"
  STUB_REDIRECTS="www.example.com|example.com|301"

  run invoke_caddy_docker_args
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--label 'caddy=" 0
  assert_output_contains "--label 'caddy_1=www.example.com:80'"
  assert_output_contains "--label 'caddy_1.redir=http://example.com{uri} 301'"
}
