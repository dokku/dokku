#!/usr/bin/env bats

load test_helper

setup() {
  deploy_app
}

teardown() {
  destroy_app
}

@test "(tag) tag:add, tag:list, tag:rm" {
  run /bin/bash -c "dokku tag:add $TEST_APP v0.9.0"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku tag:list $TEST_APP | egrep -q 'v0.9.0'"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run /bin/bash -c "dokku tag:rm $TEST_APP v0.9.0"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku tag:list $TEST_APP | egrep -q 'v0.9.0'"
  echo "output: "$output
  echo "status: "$status
  assert_failure

}

@test "(tag) tag:deploy" {
  run /bin/bash -c "dokku tag:add $TEST_APP v0.9.0"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku tag:deploy $TEST_APP v0.9.0"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "docker ps | egrep '/start web'| egrep -q dokku/${TEST_APP}:v0.9.0"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
