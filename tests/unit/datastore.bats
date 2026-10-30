#!/usr/bin/env bats

load test_helper

TEST_SERVICE="ds-test-service"

setup() {
  global_setup
  create_app
}

teardown() {
  dokku datastore:unlink redis "$TEST_SERVICE" "$TEST_APP" --no-restart >/dev/null 2>&1 || true
  dokku datastore:destroy redis "$TEST_SERVICE" --force >/dev/null 2>&1 || true
  destroy_app
  global_teardown
}

@test "(datastore) datastore:help" {
  run /bin/bash -c "dokku datastore"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Manage datastore services."
  assert_output_contains "datastore:create <datastore-type> <service>"
  help_output="$output"

  run /bin/bash -c "dokku datastore:help"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$help_output"

  run /bin/bash -c "dokku datastore:help redis"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku datastore:list redis"

  run /bin/bash -c "dokku datastore:help redis create"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku datastore:create redis <service>"

  run /bin/bash -c "dokku datastore:help nonsense"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku help --all"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "datastore:create <datastore-type> <service>"
}

@test "(datastore) datastore:version" {
  run /bin/bash -c "dokku datastore:version"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "v"
}

@test "(datastore) create, link, unlink and destroy a service" {
  run /bin/bash -c "dokku datastore:create redis $TEST_SERVICE"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku datastore:exists redis $TEST_SERVICE"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku datastore:list redis"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$TEST_SERVICE"

  run /bin/bash -c "dokku datastore:link redis $TEST_SERVICE $TEST_APP --no-restart"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP REDIS_URL"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "redis://"

  run /bin/bash -c "dokku datastore:unlink redis $TEST_SERVICE $TEST_APP --no-restart"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP REDIS_URL"
  echo "output: $output"
  echo "status: $status"
  assert_output ""

  run /bin/bash -c "dokku datastore:destroy redis $TEST_SERVICE --force"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku datastore:exists redis $TEST_SERVICE"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(datastore) deleting an app unlinks its services" {
  run /bin/bash -c "dokku datastore:create redis $TEST_SERVICE"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku datastore:link redis $TEST_SERVICE $TEST_APP --no-restart"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku datastore:linked redis $TEST_SERVICE $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  create_app
}
