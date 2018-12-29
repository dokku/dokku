#!/usr/bin/env bash

set -eo pipefail
readonly ROOT_DIR="$(cd "$(dirname "$(dirname "$(dirname "${BASH_SOURCE[0]}")")")" && pwd)"

install_dependencies() {
  mkdir -p "$ROOT_DIR/build/"
  HEROKUISH_VERSION=$(grep HEROKUISH_VERSION "${ROOT_DIR}/deb.mk" | head -n1 | cut -d' ' -f3)
  HEROKUISH_PACKAGE_NAME="herokuish_${HEROKUISH_VERSION}_amd64.deb"
  curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/trusty/herokuish_${HEROKUISH_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${HEROKUISH_PACKAGE_NAME}"

  PLUGN_VERSION=$(grep PLUGN_VERSION "${ROOT_DIR}/deb.mk" | head -n1 | cut -d' ' -f3)
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

  sudo add-apt-repository -y ppa:nginx/stable
  sudo apt-get update
  sudo apt-get -qq -y install nginx
  sudo cp "${ROOT_DIR}/tests/dhparam.pem" /etc/nginx/dhparam.pem

  sudo dpkg -i "${ROOT_DIR}/build/$HEROKUISH_PACKAGE_NAME"
  sudo dpkg -i "${ROOT_DIR}/build/$PLUGN_PACKAGE_NAME"
  sudo dpkg -i "${ROOT_DIR}/build/$SSHCOMMAND_PACKAGE_NAME"
  sudo dpkg -i "${ROOT_DIR}/build/$SIGIL_PACKAGE_NAME"
  sudo dpkg -i "${ROOT_DIR}/build/$PROCFILE_UTIL_PACKAGE_NAME"
}

install_dokku() {
  if [[ "$FROM_SOURCE" == "true" ]]; then
    sudo -E CI=true make -e install
    return
  fi

  "${ROOT_DIR}/contrib/release" build

  echo "dokku dokku/hostname string dokku.me"              | sudo debconf-set-selections
  echo "dokku dokku/key_file string /root/.ssh/id_rsa.pub" | sudo debconf-set-selections
  echo "dokku dokku/nginx_enable boolean true"             | sudo debconf-set-selections
  echo "dokku dokku/skip_key_file boolean true"            | sudo debconf-set-selections
  echo "dokku dokku/vhost_enable boolean true"             | sudo debconf-set-selections
  echo "dokku dokku/web_config boolean false"              | sudo debconf-set-selections
  sudo dpkg -i "$(cat "${ROOT_DIR}/build/deb-filename")"
}

# shellcheck disable=SC2120
setup_circle() {
  echo "=====> setup_circle on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
  sudo -E CI=true make -e sshcommand
  # need to add the dokku user to the docker group
  sudo usermod -G docker dokku
  [[ "$1" == "buildstack" ]] && BUILD_STACK=true make -e stack

  install_dependencies
  install_dokku

  sudo -E make -e setup-deploy-tests
  bash --version
  docker version
  lsb_release -a
  # setup .dokkurc
  sudo -E mkdir -p /home/dokku/.dokkurc
  sudo -E chown dokku:ubuntu /home/dokku/.dokkurc
  sudo -E chmod 775 /home/dokku/.dokkurc
  # pull node:4 image for testing
  sudo docker pull node:4
}

# shellcheck disable=SC2119
setup_circle
exit $?
