#!/usr/bin/env bats

load ../unit/test_helper

@test "deploy multi app" {
  skip "fails on postinstall with Arguments to path.join must be strings"
  run bash -c "cd tests && ./test_deploy ./apps/multi dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
