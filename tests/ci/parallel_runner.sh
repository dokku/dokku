#!/usr/bin/env bash

set -eo pipefail

MODE="$1"; MODE=${MODE:="testing"}

# shellcheck disable=SC2120
setup_circle() {
  echo "=====> setup_circle on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
  sudo -E CI=true make -e sshcommand
  # need to add the dokku user to the docker group
  sudo usermod -G docker dokku
  #### circle does some weird *expletive* with regards to root and gh auth (needed for gitsubmodules test)
  sudo rsync -a ~ubuntu/.ssh/ ~root/.ssh/
  sudo chown -R root:root ~root/.ssh/
  sudo sed --in-place 's:/home/ubuntu:/root:g' ~root/.ssh/config
  ####
  [[ "$1" == "buildstack" ]] && BUILD_STACK=true make -e stack
  sudo -E CI=true make -e install
  sudo -E make -e setup-deploy-tests
  make -e ci-dependencies
  # circleci runs Ubuntu 12.04 and thus a previous version of bash (4.2) than 14.04
  # sudo apt-get install -y -q "bash=$(apt-cache show bash | egrep "^Version: 4.3" | head -1 | awk -F: '{ print $2 }' | xargs)"
  bash --version
  docker version
  # setup .dokkurc
  sudo -E mkdir -p /home/dokku/.dokkurc
  sudo -E chown dokku:ubuntu /home/dokku/.dokkurc
  sudo -E chmod 775 /home/dokku/.dokkurc
}

if [[ -n "$CIRCLE_NODE_INDEX" ]] && [[ "$MODE" == "setup" ]]; then
  # shellcheck disable=SC2119
  setup_circle
  exit $?
  # case "$CIRCLE_NODE_INDEX" in
  #     3)
  #       setup_circle buildstack
  #       exit $?
  #       ;;
  #     *)
  #       setup_circle
  #       exit $?
  #       ;;
  # esac
fi

case "$CIRCLE_NODE_INDEX" in
  0)
    echo "=====> make unit-tests (1/4) on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
    sudo -E UNIT_TEST_BATCH=1 make -e unit-tests
    ;;

  1)
    echo "=====> make unit-tests (2/4) on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
    sudo -E UNIT_TEST_BATCH=2 make -e unit-tests
    ;;

  2)
    echo "=====> make unit-tests (3/4) on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
    sudo -E UNIT_TEST_BATCH=3 make -e unit-tests
    echo "=====> make deploy tests"
    sudo -E make -e deploy-test-checks-root deploy-test-config deploy-test-multi
    ;;

  3)
    echo "=====> make unit-tests (4/4) on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
    sudo -E UNIT_TEST_BATCH=4 make -e unit-tests
    ;;
esac
