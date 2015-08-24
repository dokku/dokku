#!/usr/bin/env bash

set -eo pipefail

MODE="$1"; MODE=${MODE:="testing"}

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
}

if [[ -n "$CIRCLE_NODE_INDEX" ]] && [[ "$MODE" == "setup" ]]; then
  case "$CIRCLE_NODE_INDEX" in
      3)
        setup_circle buildstack
        exit $?
        ;;
      *)
        setup_circle
        exit $?
        ;;
  esac
fi

case "$CIRCLE_NODE_INDEX" in
  0)
    echo "=====> make unit-tests (1/2) on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
    sudo -E UNIT_TEST_BATCH=1 make -e unit-tests
    ;;

  1)
    echo "=====> make unit-tests (2/2) on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
    sudo -E UNIT_TEST_BATCH=2 make -e unit-tests
    ;;

  2)
    echo "=====> make deploy-tests (herokuish release) on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
    sudo -E make -e deploy-tests
    ;;

  3)
    echo "=====> make deploy-tests (herokuish master) on CIRCLE_NODE_INDEX: $CIRCLE_NODE_INDEX"
    sudo -E make -e deploy-tests
    ;;
esac
