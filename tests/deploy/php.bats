#!/usr/bin/env bats

load ../unit/test_helper

@test "deploy php app" {
  skip "fails. ref: https://github.com/progrium/buildstep/issues/126"
  run bash -c "cd tests && ./test_deploy ./apps/php dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
