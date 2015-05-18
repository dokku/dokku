#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "docker-options:add (all phases)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP | egrep '\-v /tmp' | wc -l | grep -q 3"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "docker-options:add (build phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP build | egrep '\-v /tmp' | wc -l | grep -q 1"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "docker-options:add (deploy phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP deploy | egrep '\-v /tmp' | wc -l | grep -q 1"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "docker-options:add (run phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP run \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP run | egrep '\-v /tmp' | wc -l | grep -q 1"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "docker-options:remove (all phases)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP | egrep '\-v /tmp' | wc -l | grep -q 3"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options:remove $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP | wc -l | grep -q 0"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "docker-options:remove (build phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP | egrep '\-v /tmp' | wc -l | grep -q 3"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options:remove $TEST_APP build \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP build"
  echo "output: "$output
  echo "status: "$status
  assert_output "Build options: none"
}

@test "docker-options:remove (deploy phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP deploy | egrep '\-v /tmp' | wc -l | grep -q 1"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options:remove $TEST_APP deploy \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP deploy"
  echo "output: "$output
  echo "status: "$status
  assert_output "Deploy options: none"
}

@test "docker-options:remove (run phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP run | egrep '\-v /tmp' | wc -l | grep -q 1"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options:remove $TEST_APP run \"-v /tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP run"
  echo "output: "$output
  echo "status: "$status
  assert_output "Run options: none"
}

@test "docker-options (deploy with options)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy \"-v /var/tmp\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "echo '-v /tmp' >> $DOKKU_ROOT/$TEST_APP/DOCKER_OPTIONS_DEPLOY"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "echo '# comment' >> $DOKKU_ROOT/$TEST_APP/DOCKER_OPTIONS_DEPLOY"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP deploy | egrep '\-v /tmp' | wc -l | grep -q 1"
  echo "output: "$output
  echo "status: "$status
  assert_success
  deploy_app
}

@test "docker-options:add (all phases over SSH)" {
  run ssh dokku@dokku.me docker-options:add $TEST_APP build,deploy,run "-v /tmp"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run /bin/bash -c "dokku docker-options $TEST_APP | egrep '\-v /tmp' | wc -l | grep -q 3"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
