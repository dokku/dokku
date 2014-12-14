#!/usr/bin/env bats

load ../unit/test_helper

@test "deploy static app" {
  skip "fails on apt-get update..."
  run bash -c "cd tests && ./test_deploy ./apps/static dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
