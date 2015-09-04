#!/usr/bin/env bats

load test_helper

setup() {
  deploy_app
}

teardown() {
  destroy_app
}

@test "(tags) tags:create, tags, tags:destroy" {
  run /bin/bash -c "dokku tags:create $TEST_APP v0.9.0"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku tags $TEST_APP | egrep -q 'v0.9.0'"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run /bin/bash -c "dokku tags:destroy $TEST_APP v0.9.0"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku tags $TEST_APP | egrep -q 'v0.9.0'"
  echo "output: "$output
  echo "status: "$status
  assert_failure
}

@test "(tags) tags:deploy" {
  run /bin/bash -c "dokku tags:create $TEST_APP v0.9.0"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku tags:deploy $TEST_APP v0.9.0"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "docker ps | egrep '/start web'| egrep -q dokku/${TEST_APP}:v0.9.0"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
