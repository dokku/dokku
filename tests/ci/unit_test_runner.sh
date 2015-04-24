#!/usr/bin/env bash

is_number() {
  local NUMBER=$1; local NUM_RE='^[0-9]+$'
  if [[ $NUMBER =~ $NUM_RE ]];then
    return 0
  else
    return 1
  fi
}

usage() {
  echo "usage: $0 1|2"
  exit 0
}

BATCH_NUM="$1"
is_number $BATCH_NUM || usage

TESTS=($(find "$(dirname $0)"/../unit -maxdepth 1 -name "*.bats"))
HALF_TESTS=$(( ${#TESTS[@]} / 2 ))
FIRST_HALF=("${TESTS[@]:0:${HALF_TESTS}}")
LAST_HALF=("${TESTS[@]:${HALF_TESTS}:${#TESTS[@]}}")

case "$BATCH_NUM" in
  1)
    bats "${FIRST_HALF[@]}"
  ;;

  2)
    bats "${LAST_HALF[@]}"
  ;;

  *)
    usage
  ;;
esac
