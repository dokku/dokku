#!/usr/bin/env bats

load ../unit/test_helper

@test "deploy python-flask app" {
  run bash -c "cd tests && ./test_deploy ./apps/python-flask dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
