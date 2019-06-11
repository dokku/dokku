#!/usr/bin/env bash

set -eo pipefail
readonly ROOT_DIR="$(cd "$(dirname "$(dirname "$(dirname "${BASH_SOURCE[0]}")")")" && pwd)"

install_dependencies() {
  echo "=====> install_dependencies on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"

  mkdir -p "$ROOT_DIR/build/"
  HEROKUISH_VERSION=$(grep HEROKUISH_VERSION "${ROOT_DIR}/deb.mk" | head -n1 | cut -d' ' -f3)
  HEROKUISH_PACKAGE_NAME="herokuish_${HEROKUISH_VERSION}_amd64.deb"
  curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/trusty/herokuish_${HEROKUISH_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${HEROKUISH_PACKAGE_NAME}"

  PLUGN_VERSION=$(grep PLUGN_VERSION "${ROOT_DIR}/Makefile" | head -n1 | cut -d' ' -f3)
  PLUGN_PACKAGE_NAME="plugn_${PLUGN_VERSION}_amd64.deb"
  curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/trusty/plugn_${PLUGN_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${PLUGN_PACKAGE_NAME}"

  SSHCOMMAND_VERSION=$(grep SSHCOMMAND_VERSION "${ROOT_DIR}/deb.mk" | head -n1 | cut -d' ' -f3)
  SSHCOMMAND_PACKAGE_NAME="sshcommand_${SSHCOMMAND_VERSION}_amd64.deb"
  curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/trusty/sshcommand_${SSHCOMMAND_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${SSHCOMMAND_PACKAGE_NAME}"

  SIGIL_VERSION=$(grep SIGIL_VERSION "${ROOT_DIR}/deb.mk" | head -n1 | cut -d' ' -f3)
  SIGIL_PACKAGE_NAME="gliderlabs-sigil_${SIGIL_VERSION}_amd64.deb"
  curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/trusty/gliderlabs-sigil_${SIGIL_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${SIGIL_PACKAGE_NAME}"

  PROCFILE_VERSION=$(grep PROCFILE_VERSION "${ROOT_DIR}/Makefile" | head -n1 | cut -d' ' -f3)
  PROCFILE_UTIL_PACKAGE_NAME="procfile-util_${PROCFILE_VERSION}_amd64.deb"
  curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/trusty/procfile-util_${PROCFILE_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${PROCFILE_UTIL_PACKAGE_NAME}"

  sudo dpkg -i "${ROOT_DIR}/build/$HEROKUISH_PACKAGE_NAME" \
    "${ROOT_DIR}/build/$PLUGN_PACKAGE_NAME" \
    "${ROOT_DIR}/build/$SSHCOMMAND_PACKAGE_NAME" \
    "${ROOT_DIR}/build/$SIGIL_PACKAGE_NAME" \
    "${ROOT_DIR}/build/$PROCFILE_UTIL_PACKAGE_NAME"
}

build_dokku() {
  echo "=====> build_dokku on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
  "${ROOT_DIR}/contrib/release-dokku" build
  docker build -t dokku/dokku:test .
}

run_dokku_container() {
  echo "=====> run_dokku_container on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
  docker run -d \
    --env DOKKU_HOSTNAME=dokku.me \
    --name dokku \
    --publish 3022:22 \
    --publish 22333:22 \
    --publish 80:80 \
    --publish 443:443 \
    --volume /var/lib/dokku:/mnt/dokku \
    --volume /var/run/docker.sock:/var/run/docker.sock \
    dokku/dokku:test

  check_container
}

check_container() {
  echo "=====> check_container on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
  local is_up
  local cnt=0
  while true; do
    echo "$(date) [count: $cnt]: waiting for dokku startup"
    is_up=$(
      docker exec -ti dokku ps -ef | grep "/usr/sbin/sshd -D" >/dev/null 2>&1
      echo $?
    )
    if [[ $is_up -eq 0 ]]; then
      echo "" && docker logs dokku
      break
    fi
    sleep 2
    cnt=$((cnt + 1))
  done
}

install_dokku() {
  echo "=====> install_dokku on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"

  if [[ "$FROM_SOURCE" == "true" ]]; then
    sudo -E CI=true make -e install
    return
  fi

  build_dokku
  run_dokku_container

  sudo mkdir -p /usr/local/bin
  sudo cp -fv contrib/dokku-docker-bin.sh /usr/local/bin/dokku
}

# shellcheck disable=SC2120
setup_circle() {
  echo "=====> setup_circle on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"

  install_dependencies
  install_dokku

  sudo -E make -e setup-deploy-tests
  lsb_release -a
  # pull node:4 image for testing
  sudo docker pull node:4
}

# shellcheck disable=SC2119
setup_circle
exit $?
