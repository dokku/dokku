#!/usr/bin/env bats

load test_helper

setup() {
  deploy_app
}

teardown() {
  destroy_app
}

@test "remove exited containers" {
  run dokku run $TEST_APP hostname
  echo "output: "$output
  echo "status: "$status
  container="$output"
  assert_success
  run bash -c "docker ps -a -f 'status=exited' -q --no-trunc=false | grep -q $container"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku cleanup
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "docker ps -a -f 'status=exited' -q --no-trunc=false | grep -q $container"
  echo "output: "$output
  echo "status: "$status
  assert_failure
  run bash -c "docker ps -a -f 'status=exited' -q --no-trunc=false"
  echo "output: "$output
  echo "status: "$status
  assert_output ""
}
