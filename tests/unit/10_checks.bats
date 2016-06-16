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
  run bash -c "dokku checks $TEST_APP| grep $TEST_APP | xargs"
  echo "output: "$output
  echo "status: "$status
  assert_output "$TEST_APP none none"
}

@test "(checks) checks:disable" {
  run bash -c "dokku checks:disable $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_DISABLED"
  echo "output: "$output
  echo "status: "$status
  assert_output "_all_"
}

@test "(checks) checks:disable -> checks:enable" {
  run bash -c "dokku checks:disable $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_DISABLED"
  echo "output: "$output
  echo "status: "$status
  assert_output "_all_"

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_SKIPPED"
  echo "output: "$output
  echo "status: "$status
  assert_output ""

  run bash -c "dokku checks:enable $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_DISABLED"
  echo "output: "$output
  echo "status: "$status
  assert_output ""

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_SKIPPED"
  echo "output: "$output
  echo "status: "$status
  assert_output ""
}

@test "(checks) checks:disable -> checks:skip" {
  run bash -c "dokku checks:disable $TEST_APP web,worker,urgentworker,notifications"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_DISABLED"
  echo "output: "$output
  echo "status: "$status
  assert_output "web,worker,urgentworker,notifications"

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_SKIPPED"
  echo "output: "$output
  echo "status: "$status
  assert_output ""

  run bash -c "dokku checks:skip $TEST_APP urgentworker,worker"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_SKIPPED"
  echo "output: "$output
  echo "status: "$status
  assert_output "urgentworker,worker"

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_DISABLED"
  echo "output: "$output
  echo "status: "$status
  assert_output "web,notifications"
}

@test "(checks) checks:skip" {
  run bash -c "dokku checks:skip $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_SKIPPED"
  echo "output: "$output
  echo "status: "$status
  assert_output "_all_"
}

@test "(checks) checks:skip -> checks:enable" {
  run bash -c "dokku checks:skip $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_SKIPPED"
  echo "output: "$output
  echo "status: "$status
  assert_output "_all_"

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_DISABLED"
  echo "output: "$output
  echo "status: "$status
  assert_output ""

  run bash -c "dokku checks:enable $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_SKIPPED"
  echo "output: "$output
  echo "status: "$status
  assert_output ""

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_DISABLED"
  echo "output: "$output
  echo "status: "$status
  assert_output ""
}

@test "(checks) checks:skip -> checks:disable" {
  run bash -c "dokku checks:skip $TEST_APP web,worker,urgentworker,notifications"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_SKIPPED"
  echo "output: "$output
  echo "status: "$status
  assert_output "web,worker,urgentworker,notifications"

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_DISABLED"
  echo "output: "$output
  echo "status: "$status
  assert_output ""

  run bash -c "dokku checks:disable $TEST_APP urgentworker,worker"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_DISABLED"
  echo "output: "$output
  echo "status: "$status
  assert_output "urgentworker,worker"

  run bash -c "dokku config:get $TEST_APP DOKKU_CHECKS_SKIPPED"
  echo "output: "$output
  echo "status: "$status
  assert_output "web,notifications"
}
