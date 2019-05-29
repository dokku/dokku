#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  deploy_app
  rm -f /tmp/fake-docker-bin
}

teardown() {
  rm -f /tmp/fake-docker-bin
  destroy_app
  global_teardown
}

@test "(report) report" {
  run /bin/bash -c "dokku report"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "herokuish version:"
  assert_success

  run /bin/bash -c "dokku report $TEST_APP 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "herokuish version:"
  assert_output_contains "not deployed" "0"
  assert_success

  run /bin/bash -c "dokku report fake-app-name"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku report fake-app-name 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "App fake-app-name does not exist"
  assert_failure

  dokku apps:create "${TEST_APP}-2"
  run /bin/bash -c "dokku report ${TEST_APP}-2 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "not deployed"
  assert_success

  dokku --force apps:destroy "${TEST_APP}-2"
}

@test "(report) custom docker bin" {
  export DOCKER_BIN="docker"
  run /bin/bash -c "dokku report"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "herokuish version:"
  assert_success

  export DOCKER_BIN="/usr/bin/docker"
  run /bin/bash -c "dokku report"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "herokuish version:"
  assert_success

  touch /tmp/fake-docker-bin
  echo '#!/usr/bin/env bash' >/tmp/fake-docker-bin
  echo '/usr/bin/docker "$@"' >>/tmp/fake-docker-bin
  chmod +x /tmp/fake-docker-bin

  export DOCKER_BIN="/tmp/fake-docker-bin"
  run /bin/bash -c "dokku report"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "herokuish version:"
  assert_success

  unset DOCKER_BIN
}
