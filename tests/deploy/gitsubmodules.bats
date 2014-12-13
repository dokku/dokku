#!/usr/bin/env bats

load ../unit/test_helper

@test "deploy gitsubmodules app" {
  run bash -c "cd tests && ./test_deploy ./apps/gitsubmodules dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
