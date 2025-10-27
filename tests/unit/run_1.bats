#!/usr/bin/env bats

load test_helper

setup_file() {
  install_pack
}

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(run) run:help" {
  run /bin/bash -c "dokku run:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Run a one-off process inside a container"
}

@test "(run) run (with --options)" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force --quiet run $TEST_APP python3 -V"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(run) run herokuish (with --env / -e)" {
  run /bin/bash -c "dokku config:set --no-restart --global GLOBAL_SECRET=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_requirements_txt
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run --env TEST=testvalue -e TEST2=testvalue2 $TEST_APP env | grep -E '^TEST=testvalue'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run --env TEST=testvalue -e TEST2=testvalue2 $TEST_APP env | grep -E '^TEST2=testvalue2'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(run) run cnb (with --env / -e)" {
  run /bin/bash -c "dokku config:set --no-restart --global GLOBAL_SECRET=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected pack"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP add_requirements_txt_cnb
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run --env TEST=testvalue -e TEST2=testvalue2 $TEST_APP env"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run --env TEST=testvalue -e TEST2=testvalue2 $TEST_APP env | grep -E '^TEST=testvalue'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run --env TEST=testvalue -e TEST2=testvalue2 $TEST_APP env | grep -E '^TEST2=testvalue2'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
