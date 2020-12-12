#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(tags) tags:help" {
  run /bin/bash -c "dokku tags:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage docker image tags"
}

@test "(tags) tags:create, tags, tags:destroy" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku tags:create $TEST_APP v0.9.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku tags $TEST_APP | grep -q -E 'v0.9.0'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku tags:destroy $TEST_APP v0.9.0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku tags $TEST_APP | grep -q -E 'v0.9.0'"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(tags) tags:deploy" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku tags:create $TEST_APP v0.9.0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku tags:deploy $TEST_APP v0.9.0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "docker ps | grep -E '/start web'| grep -q -E dokku/${TEST_APP}:v0.9.0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "docker images | grep -E "dokku/${TEST_APP}"| grep -q -E latest"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(tags) tags:deploy (missing tag)" {
  run /bin/bash -c "dokku tags:deploy $TEST_APP missing-tag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(tags) tags:deploy (onbuild)" {
  run /bin/bash -c "docker image pull gliderlabs/logspout:v3.2.13"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image tag gliderlabs/logspout:v3.2.13 dokku/$TEST_APP:3.2.13"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP /var/run/docker.sock:/var/run/docker.sock"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku tags:deploy $TEST_APP 3.2.13"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Image contains one or more ONBUILD directives, skipping label injection"
  assert_output_contains "COPY failed" 0
  assert_success
}
