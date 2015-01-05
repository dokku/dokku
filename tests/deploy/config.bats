#!/usr/bin/env bats

load ../unit/test_helper

@test "deploy config app" {
  run bash -c "cd tests && ./test_deploy ./apps/config dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
