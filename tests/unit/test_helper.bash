#!/usr/bin/env bash

# constants
TEST_APP=my-cool-guy-test-app

# test functions
flunk() {
  { if [ "$#" -eq 0 ]; then cat -
    else echo "$*"
    fi
  }
  return 1
}

assert_success() {
  if [ "$status" -ne 0 ]; then
    flunk "command failed with exit status $status"
  elif [ "$#" -gt 0 ]; then
    assert_output "$1"
  fi
}

assert_failure() {
  if [ "$status" -eq 0 ]; then
    flunk "expected failed exit status"
  elif [ "$#" -gt 0 ]; then
    assert_output "$1"
  fi
}

assert_equal() {
  if [ "$1" != "$2" ]; then
    { echo "expected: $1"
      echo "actual:   $2"
    } | flunk
  fi
}

assert_output() {
  local expected
  if [ $# -eq 0 ]; then expected="$(cat -)"
  else expected="$1"
  fi
  assert_equal "$expected" "$output"
}

assert_line() {
  if [ "$1" -ge 0 ] 2>/dev/null; then
    assert_equal "$2" "${lines[$1]}"
  else
    local line
    for line in "${lines[@]}"; do
      if [ "$line" = "$1" ]; then return 0; fi
    done
    flunk "expected line \`$1'"
  fi
}

refute_line() {
  if [ "$1" -ge 0 ] 2>/dev/null; then
    local num_lines="${#lines[@]}"
    if [ "$1" -lt "$num_lines" ]; then
      flunk "output has $num_lines lines"
    fi
  else
    local line
    for line in "${lines[@]}"; do
      if [ "$line" = "$1" ]; then
        flunk "expected to not find line \`$line'"
      fi
    done
  fi
}

assert() {
  if ! "$*"; then
    flunk "failed: $*"
  fi
}

# dokku functions
create_app() {
  dokku apps:create $TEST_APP
}

destroy_app() {
  echo $TEST_APP | dokku apps:destroy $TEST_APP
}

deploy_app() {
  TMP=$(mktemp -d -t "$TARGET.XXXXX")
  rmdir $TMP && cp -r ./tests/apps/config $TMP
  cd $TMP
  git init
  git config user.email "robot@example.com"
  git config user.name "Test Robot"
  git remote add target dokku@dokku.me:$TEST_APP

  [[ -f gitignore ]] && mv gitignore .gitignore
  git add .
  git commit -m 'initial commit'
  git push target master || destroy_app
}

setup_test_tls() {
  TLS="/home/dokku/$TEST_APP/tls"
  mkdir -p $TLS
  tar xf $BATS_TEST_DIRNAME/server_ssl.tar -C $TLS
  sudo chown -R dokku:dokku $TLS
}
