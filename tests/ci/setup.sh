#!/usr/bin/env bash

set -eo pipefail

# shellcheck disable=SC2120
setup_circle() {
  echo "=====> setup_circle on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
  sudo -E CI=true make -e sshcommand
  # need to add the dokku user to the docker group
  sudo usermod -G docker dokku
  [[ "$1" == "buildstack" ]] && BUILD_STACK=true make -e stack

  sudo add-apt-repository -y ppa:nginx/stable
  sudo apt-get update
  sudo apt-get -qq -y install nginx
  sudo cp tests/dhparam.pem /etc/nginx/dhparam.pem

  sudo -E CI=true make -e install
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
