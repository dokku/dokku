#!/usr/bin/env bats

load ../unit/test_helper

@test "deploy java app" {
  run bash -c "cd tests && ./test_deploy ./apps/java dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
