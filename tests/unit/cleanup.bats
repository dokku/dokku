#!/usr/bin/env bats

load test_helper

setup() {
  deploy_app
}

teardown() {
  destroy_app
}

@test "remove exited containers" {
  # make sure we have many exited containers of the same 'type'
  run bash -c "for cnt in 1 2 3; do dokku run $TEST_APP hostname; done"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "docker ps -a -f 'status=exited' --no-trunc=false | grep '/exec hostname'"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku cleanup
  echo "output: "$output
  echo "status: "$status
  assert_success
  sleep 5  # wait for dokku cleanup to happen in the background
  run bash -c "docker ps -a -f 'status=exited' --no-trunc=false | grep '/exec hostname'"
  echo "output: "$output
  echo "status: "$status
  assert_failure
  run bash -c "docker ps -a -f 'status=exited' -q --no-trunc=false"
  echo "output: "$output
  echo "status: "$status
  assert_output ""
}
