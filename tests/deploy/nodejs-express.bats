#!/usr/bin/env bats

load ../unit/test_helper

@test "deploy nodejs-express app" {
  run bash -c "cd tests && ./test_deploy ./apps/nodejs-express dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
