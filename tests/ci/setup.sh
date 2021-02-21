#!/usr/bin/env bash

set -eo pipefail
readonly ROOT_DIR="$(cd "$(dirname "$(dirname "$(dirname "${BASH_SOURCE[0]}")")")" && pwd)"

install_dependencies() {
  echo "=====> install_dependencies on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"

  mkdir -p "$ROOT_DIR/build/"

  DOCKER_IMAGE_LABELER_VERSION=$(grep DOCKER_IMAGE_LABELER_VERSION "${ROOT_DIR}/Makefile" | head -n1 | cut -d' ' -f3)
  DOCKER_IMAGE_LABELER_PACKAGE_NAME="docker-image-labeler_${DOCKER_IMAGE_LABELER_VERSION}_amd64.deb"
  if [[ ! -f "$ROOT_DIR/build/${DOCKER_IMAGE_LABELER_PACKAGE_NAME}" ]]; then
    curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/bionic/docker-image-labeler_${DOCKER_IMAGE_LABELER_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${DOCKER_IMAGE_LABELER_PACKAGE_NAME}"
  fi

  HEROKUISH_VERSION=$(grep HEROKUISH_VERSION "${ROOT_DIR}/Makefile" | head -n1 | cut -d' ' -f3)
  HEROKUISH_PACKAGE_NAME="herokuish_${HEROKUISH_VERSION}_amd64.deb"
  if [[ ! -f "$ROOT_DIR/build/${HEROKUISH_PACKAGE_NAME}" ]]; then
    curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/bionic/herokuish_${HEROKUISH_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${HEROKUISH_PACKAGE_NAME}"
  fi

  PLUGN_VERSION=$(grep PLUGN_VERSION "${ROOT_DIR}/Makefile" | head -n1 | cut -d' ' -f3)
  PLUGN_PACKAGE_NAME="plugn_${PLUGN_VERSION}_amd64.deb"
  if [[ ! -f "$ROOT_DIR/build/${PLUGN_PACKAGE_NAME}" ]]; then
    curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/bionic/plugn_${PLUGN_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${PLUGN_PACKAGE_NAME}"
  fi

  SSHCOMMAND_VERSION=$(grep SSHCOMMAND_VERSION "${ROOT_DIR}/Makefile" | head -n1 | cut -d' ' -f3)
  SSHCOMMAND_PACKAGE_NAME="sshcommand_${SSHCOMMAND_VERSION}_amd64.deb"
  if [[ ! -f "$ROOT_DIR/build/${SSHCOMMAND_PACKAGE_NAME}" ]]; then
    curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/bionic/sshcommand_${SSHCOMMAND_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${SSHCOMMAND_PACKAGE_NAME}"
  fi

  SIGIL_VERSION=$(grep SIGIL_VERSION "${ROOT_DIR}/Makefile" | head -n1 | cut -d' ' -f3)
  SIGIL_PACKAGE_NAME="gliderlabs-sigil_${SIGIL_VERSION}_amd64.deb"
  if [[ ! -f "$ROOT_DIR/build/${SIGIL_PACKAGE_NAME}" ]]; then
    curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/bionic/gliderlabs-sigil_${SIGIL_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${SIGIL_PACKAGE_NAME}"
  fi

  PROCFILE_VERSION=$(grep PROCFILE_VERSION "${ROOT_DIR}/Makefile" | head -n1 | cut -d' ' -f3)
  PROCFILE_UTIL_PACKAGE_NAME="procfile-util_${PROCFILE_VERSION}_amd64.deb"
  if [[ ! -f "$ROOT_DIR/build/${PROCFILE_UTIL_PACKAGE_NAME}" ]]; then
    curl -L "https://packagecloud.io/dokku/dokku/packages/ubuntu/bionic/procfile-util_${PROCFILE_VERSION}_amd64.deb/download.deb" -o "$ROOT_DIR/build/${PROCFILE_UTIL_PACKAGE_NAME}"
  fi

  sudo add-apt-repository -y ppa:nginx/stable
  sudo apt-get update -qq
  sudo apt-get -qq -y --no-install-recommends install cgroupfs-mount dos2unix jq nginx debconf-utils
  sudo cp "${ROOT_DIR}/tests/dhparam.pem" /etc/nginx/dhparam.pem

  sudo dpkg -i \
    "${ROOT_DIR}/build/$DOCKER_IMAGE_LABELER_PACKAGE_NAME" \
    "${ROOT_DIR}/build/$HEROKUISH_PACKAGE_NAME" \
    "${ROOT_DIR}/build/$PLUGN_PACKAGE_NAME" \
    "${ROOT_DIR}/build/$PROCFILE_UTIL_PACKAGE_NAME" \
    "${ROOT_DIR}/build/$SIGIL_PACKAGE_NAME" \
    "${ROOT_DIR}/build/$SSHCOMMAND_PACKAGE_NAME"
}

build_dokku() {
  if [[ "$FROM_SOURCE" == "true" ]]; then
    sudo -E CI=true make -e install
    return
  fi

  echo "=====> build_dokku on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
  "${ROOT_DIR}/contrib/release-dokku" build
  cp "$(cat "${ROOT_DIR}/build/deb-filename")" "${ROOT_DIR}/build/dokku.deb"
  cp "$(cat "${ROOT_DIR}/build/rpm-filename")" "${ROOT_DIR}/build/dokku.rpm"
}

install_dokku() {
  echo "=====> install_dokku on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"

  if [[ ! -f "${ROOT_DIR}/build/deb-filename" ]]; then
    build_dokku
  fi

  cat <<EOF | sudo debconf-set-selections
dokku dokku/hostname string dokku.me
dokku dokku/key_file string /root/.ssh/id_rsa.pub
dokku dokku/nginx_enable boolean true
dokku dokku/skip_key_file boolean true
dokku dokku/vhost_enable boolean true
dokku dokku/web_config boolean false
EOF

  echo "-----> Start debconf selections"
  sudo debconf-get-selections | grep ^dokku
  echo "-----> End debconf selections"

  sleep 5
  echo "-----> Start install $(cat "${ROOT_DIR}/build/deb-filename") via dpkg"
  sudo TRACE=1 dpkg -i "$(cat "${ROOT_DIR}/build/deb-filename")"
  echo "-----> End install $(cat "${ROOT_DIR}/build/deb-filename") via dpkg"
}

build_dokku_docker_image() {
  echo "=====> build_dokku_docker_image on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
  docker build -t dokku/dokku:test .
}

run_dokku_container() {
  echo "=====> run_dokku_container on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
  docker run -d \
    --env DOKKU_HOSTNAME=dokku.me \
    --name dokku \
    --publish 3022:22 \
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
      docker exec dokku ps -ef | grep "/usr/sbin/sshd -D" >/dev/null 2>&1
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

create_package() {
  echo "=====> setup_circle on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
  sudo -E CI=true make -e sshcommand
  # need to add the dokku user to the docker group
  sudo usermod -G docker dokku
  [[ "$1" == "buildstack" ]] && BUILD_STACK=true make -e stack

  install_dependencies
  build_dokku
}

# shellcheck disable=SC2120
setup_circle() {
  install_dependencies
  install_dokku

  sudo -E make -e setup-deploy-tests
  lsb_release -a
  # setup .dokkurc

  getent passwd | cut -d: -f1 | sort

  sudo -E mkdir -p /home/dokku/.dokkurc
  if [[ -n "$CIRCLECI" ]]; then
    sudo -E chown dokku:ubuntu /home/dokku/.dokkurc
  else
    sudo -E chown dokku:dokku /home/dokku/.dokkurc
  fi
  sudo -E chmod 775 /home/dokku/.dokkurc
  # pull node:4-alpine image for testing
  sudo docker pull node:4-alpine
}

case "$1" in
  docker)
    sudo /etc/init.d/nginx stop
    build_dokku_docker_image
    run_dokku_container
    ;;
  build)
    # shellcheck disable=SC2119
    create_package
    exit $?
    ;;
  *)
    # shellcheck disable=SC2119
    setup_circle
    exit $?
    ;;
esac
