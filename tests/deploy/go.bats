#!/usr/bin/env bats

load ../unit/test_helper

@test "deploy go app" {
  run bash -c "cd tests && ./test_deploy ./apps/go dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
