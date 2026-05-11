#!/usr/bin/env bats

REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"

NGINX_CONF="${REPO_ROOT}/docker/etc/nginx/conf.d/zz-dokku-health.conf"
FINISH_SCRIPT="${REPO_ROOT}/docker/etc/runit/runsvdir/default/dokku-restore/finish"
MY_INIT_SCRIPT="${REPO_ROOT}/docker/etc/my_init.d/10_dokku_init"
DOCKERFILE="${REPO_ROOT}/Dockerfile"

@test "(docker-image) [healthcheck] nginx health conf is shipped via the docker overlay" {
  [[ -f "$NGINX_CONF" ]]
  grep -qE "listen[[:space:]]+127\.0\.0\.1:18080" "$NGINX_CONF"
  grep -qE "location = /_dokku/health" "$NGINX_CONF"
  grep -qE "/var/run/dokku/ready" "$NGINX_CONF"
  grep -qE "default_type[[:space:]]+text/plain" "$NGINX_CONF"
}

@test "(docker-image) [healthcheck] finish script touches sentinel after waiting for sshd and nginx" {
  [[ -f "$FINISH_SCRIPT" ]]

  local sshd_line nginx_line touch_line
  sshd_line=$(grep -nE "/dev/tcp/127\.0\.0\.1/22" "$FINISH_SCRIPT" | head -n1 | cut -d: -f1)
  nginx_line=$(grep -nE "/dev/tcp/127\.0\.0\.1/80" "$FINISH_SCRIPT" | head -n1 | cut -d: -f1)
  touch_line=$(grep -nE "touch /var/run/dokku/ready" "$FINISH_SCRIPT" | head -n1 | cut -d: -f1)

  [[ -n "$sshd_line" ]]
  [[ -n "$nginx_line" ]]
  [[ -n "$touch_line" ]]
  ((touch_line > sshd_line))
  ((touch_line > nginx_line))
}

@test "(docker-image) [healthcheck] my_init clears stale sentinel on container boot" {
  [[ -f "$MY_INIT_SCRIPT" ]]

  local rm_line mkdir_line plugin_line hostname_line
  rm_line=$(grep -nE "rm -f /var/run/dokku/ready" "$MY_INIT_SCRIPT" | head -n1 | cut -d: -f1)
  mkdir_line=$(grep -nE "mkdir -p /var/run/dokku( |$)" "$MY_INIT_SCRIPT" | head -n1 | cut -d: -f1)
  plugin_line=$(grep -nE "plugin:install" "$MY_INIT_SCRIPT" | head -n1 | cut -d: -f1)
  hostname_line=$(grep -nE "domains:set-global" "$MY_INIT_SCRIPT" | head -n1 | cut -d: -f1)

  [[ -n "$rm_line" ]]
  [[ -n "$mkdir_line" ]]
  [[ -n "$plugin_line" ]]
  [[ -n "$hostname_line" ]]
  ((rm_line < plugin_line))
  ((rm_line < hostname_line))
  ((mkdir_line < plugin_line))
  ((mkdir_line < hostname_line))
}

@test "(docker-image) [healthcheck] Dockerfile declares HEALTHCHECK with the expected endpoint" {
  [[ -f "$DOCKERFILE" ]]
  grep -qE "^HEALTHCHECK" "$DOCKERFILE"
  grep -qE "http://127\.0\.0\.1:18080/_dokku/health" "$DOCKERFILE"
  grep -qE "\-\-start-period" "$DOCKERFILE"
}
