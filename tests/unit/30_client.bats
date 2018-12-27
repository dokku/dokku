#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  export DOKKU_HOST=dokku.me
  create_app
}

teardown() {
  destroy_app
  unset DOKKU_HOST
  global_teardown
}

@test "(client) unconfigured DOKKU_HOST" {
  unset DOKKU_HOST
  run ./contrib/dokku_client.sh apps
  echo "output: $output"
  echo "status: $status"
  assert_exit_status 20
}

@test "(client) no args should print help" {
  run /bin/bash -c "./contrib/dokku_client.sh | head -1 | grep -E 'Usage: dokku \[.+\] COMMAND <app>.*'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(client) apps:create AND apps:destroy with random name" {
  setup_client_repo
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh apps:create"
  echo "output: $output"
  echo "status: $status"
  assert_success
  git remote | grep dokku
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh apps:destroy --force"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(client) apps:create AND apps:destroy with name" {
  setup_client_repo
  local test_app_name=test-apps-create-with-name
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh apps:create $test_app_name"
  echo "output: $output"
  echo "status: $status"
  assert_success
  git remote | grep dokku
  git remote -v | grep $test_app_name
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh apps:destroy --force"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(client) config:set" {
  run ./contrib/dokku_client.sh config:set $TEST_APP test_var=true test_var2=\"hello world\"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "./contrib/dokku_client.sh config:get $TEST_APP test_var2 | grep -q 'hello world'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(client) config:unset" {
  run ./contrib/dokku_client.sh config:set $TEST_APP test_var=true test_var2=\"hello world\"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run ./contrib/dokku_client.sh config:get $TEST_APP test_var
  echo "output: $output"
  echo "status: $status"
  assert_success
  run ./contrib/dokku_client.sh config:unset $TEST_APP test_var
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "./contrib/dokku_client.sh config:get $TEST_APP test_var | grep test_var"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(client) domains" {
  run /bin/bash -c "./contrib/dokku_client.sh domains:setup $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "./contrib/dokku_client.sh domains $TEST_APP | grep -q ${TEST_APP}.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(client) domains:add" {
  run ./contrib/dokku_client.sh domains:add $TEST_APP www.test.app.dokku.me
  echo "output: $output"
  echo "status: $status"
  assert_success
  run ./contrib/dokku_client.sh domains:add $TEST_APP test.app.dokku.me
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(client) domains:remove" {
  run ./contrib/dokku_client.sh domains:add $TEST_APP test.app.dokku.me
  echo "output: $output"
  echo "status: $status"
  assert_success
  run ./contrib/dokku_client.sh domains:remove $TEST_APP test.app.dokku.me
  echo "output: $output"
  echo "status: $status"
  refute_line "test.app.dokku.me"
}

@test "(client) domains:clear" {
  run ./contrib/dokku_client.sh domains:add $TEST_APP test.app.dokku.me
  echo "output: $output"
  echo "status: $status"
  assert_success
  run ./contrib/dokku_client.sh domains:clear $TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success
}

# @test "(client) ps" {
#   # CI support: 'Ah. I just spoke with our Docker expert --
#   # looks like docker exec is built to work with docker-under-libcontainer,
#   # but we're using docker-under-lxc. I don't have an estimated time for the fix, sorry
#   skip "circleci does not support docker exec at the moment."
#   deploy_app
#   run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh ps $TEST_APP | grep -q 'node web.js'"
#   echo "output: $output"
#   echo "status: $status"
#   assert_success
# }

@test "(client) ps:start" {
  deploy_app
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh ps:stop $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh ps:start $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(< $CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done
}

@test "(client) ps:stop" {
  deploy_app
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh ps:stop $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(< $CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_failure
  done
}

@test "(client) ps:restart" {
  deploy_app
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh ps:restart $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  for CID_FILE in $DOKKU_ROOT/$TEST_APP/CONTAINER.*; do
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(< $CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done
}
