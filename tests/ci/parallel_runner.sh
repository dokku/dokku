#!/usr/bin/env bash

case "$CIRCLE_NODE_INDEX" in
  0)
    echo "=====> make unit-tests (1/2)"
    sudo -E UNIT_TEST_BATCH=1 make -e unit-tests
    ;;

  1)
    echo "=====> make unit-tests (2/2)"
    sudo -E UNIT_TEST_BATCH=2 make -e unit-tests
    ;;

  2)
    echo "=====> make deploy-tests (buildstep release)"
    sudo -E make -e deploy-tests
    ;;

  3)
    echo "=====> make deploy-tests (buildstep master)"
    docker rmi -f progrium/buildstep && \
    sudo -E BUILD_STACK=true make -e stack && \
    sudo -E make -e deploy-tests
    ;;
esac
