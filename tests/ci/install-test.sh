#!/usr/bin/env bash
set -eo pipefail
[[ $TRACE ]] && set -x

readonly ROOT_DIR="$(cd "$(dirname "$(dirname "$(dirname "${BASH_SOURCE[0]}")")")" && pwd)"
readonly TEST_APP="install-test"

log-info() {
  declare desc="Log info formatter"
  echo "=====> $*"
}

log-fail() {
  declare desc="Log fail formatter"
  echo "!     $*" 1>&2
  exit 1
}

fn-has-init-system() {
  declare desc="Returns true when the host is running systemd"
  [[ -d /run/systemd/system ]]
}

fn-install-prerequisites() {
  declare desc="Installs the packages bootstrap.sh expects to already exist"
  local missing=() command

  for command in curl sudo; do
    command -v "$command" >/dev/null || missing+=("$command")
  done

  if [[ "${#missing[@]}" -gt 0 ]]; then
    apt-get update -qq
    DEBIAN_FRONTEND=noninteractive apt-get -qq -y --no-install-recommends install ca-certificates "${missing[@]}"
  fi
}

fn-disable-service-starts() {
  declare desc="Stops packages from starting services when there is no init system"
  fn-has-init-system && return 0

  log-info "No init system detected, disabling service starts"
  printf '#!/bin/sh\nexit 101\n' >/usr/sbin/policy-rc.d
  chmod +x /usr/sbin/policy-rc.d
}

fn-run-bootstrap() {
  declare desc="Installs the latest released dokku via bootstrap.sh"
  log-info "Installing the latest release via bootstrap.sh"
  DOKKU_SKIP_KEY_FILE=true \
    DOKKU_HOSTNAME=dokku.invalid \
    DOKKU_VHOST_ENABLE=false \
    bash "$ROOT_DIR/bootstrap.sh"
}

fn-install-build() {
  declare desc="Installs the dokku package built for this commit on top of the release"
  declare PACKAGE="$1"
  log-info "Installing $PACKAGE"
  DEBIAN_FRONTEND=noninteractive apt-get -qq -y install "$PACKAGE"
}

fn-verify-install() {
  declare desc="Verifies the installed dokku package is the built one and that it runs"
  declare PACKAGE="$1"
  local expected installed

  expected="$(dpkg-deb -f "$PACKAGE" Version)"
  installed="$(dpkg-query -W -f '${Version}' dokku)"
  if [[ "$installed" != "$expected" ]]; then
    log-fail "Expected dokku $expected to be installed, found $installed"
  fi

  log-info "dokku version"
  dokku version
  log-info "dokku plugin:list"
  dokku plugin:list
  log-info "dokku apps:create $TEST_APP"
  dokku apps:create "$TEST_APP"
  dokku apps:exists "$TEST_APP"
  log-info "dokku apps:destroy $TEST_APP"
  dokku --force apps:destroy "$TEST_APP"
}

main() {
  declare desc="Installs dokku on the current host and verifies the install"
  local package

  [[ "$(id -u)" -eq 0 ]] || log-fail "install-test.sh must be run as root"

  package="$ROOT_DIR/build/package/dokku-$(dpkg --print-architecture).deb"
  [[ -f "$package" ]] || log-fail "Missing package $package"

  log-info "Testing install on $(. /etc/os-release && echo "$PRETTY_NAME") ($(dpkg --print-architecture))"
  fn-install-prerequisites
  fn-disable-service-starts
  fn-run-bootstrap
  fn-install-build "$package"
  fn-verify-install "$package"
  log-info "Install verified"
}

main "$@"
