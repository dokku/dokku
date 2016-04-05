#!/usr/bin/env bats

load test_helper

#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(checks) checks" {
  run bash -c "dokku checks $TEST_APP | grep -q true"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(checks) checks:disable" {
  dokku checks:disable $TEST_APP
  assert_success

  run bash -c "dokku checks $TEST_APP | grep -q false"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(checks) checks:enable" {
  dokku checks:disable $TEST_APP
  assert_success

  dokku checks:enable $TEST_APP
  assert_success

  run bash -c "dokku checks $TEST_APP | grep -q true"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
