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
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh apps"
  echo "output: $output"
  echo "status: $status"
  assert_exit_status 20
}

@test "(client) no args should print help" {
  # dokku container is not run with a TTY on GitHub Actions so we don't get normal output
  # https://github.com/actions/runner/issues/241
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage apps"
  assert_output_contains "Manage buildpack settings for an app"
  assert_success
}

@test "(client) arg parsing" {
  run /bin/bash -c "dokku config:set --global GLOBAL_KEY=VALUE"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set $TEST_APP APP_KEY=VALUE"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh --app $TEST_APP config:export --merged --format shell"
  echo "output: $output"
  echo "status: $status"
  assert_success

  common_output="$output"
  run /bin/bash -c "dokku config:export $TEST_APP --merged --format shell"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$common_output"

  run /bin/bash -c "dokku config:export --merged $TEST_APP --format shell"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$common_output"

  run /bin/bash -c "dokku config:export --merged --format shell $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$common_output"

  run /bin/bash -c "dokku --app $TEST_APP config:export --merged --format shell"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$common_output"
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
  run ${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh config:set $TEST_APP test_var=true test_var2=\"hello world\"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh config:get $TEST_APP test_var2 | grep -q 'hello world'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(client) config:unset" {
  run ${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh config:set $TEST_APP test_var=true test_var2=\"hello world\"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh config:get $TEST_APP test_var"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh config:unset $TEST_APP test_var"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh config:get $TEST_APP test_var | grep test_var"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(client) domains:add" {
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh domains:add $TEST_APP www.test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh domains:add $TEST_APP test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(client) domains:remove" {
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh domains:add $TEST_APP test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh domains:remove $TEST_APP test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  refute_line "test.app.dokku.me"
}

@test "(client) domains:clear" {
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh domains:add $TEST_APP test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh domains:clear $TEST_APP"
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
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(<$CID_FILE)"
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
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(<$CID_FILE)"
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
    run /bin/bash -c "docker ps -q --no-trunc | grep -q $(<$CID_FILE)"
    echo "output: $output"
    echo "status: $status"
    assert_success
  done
}

@test "(client) remote management commands" {
  setup_client_repo
  local test_app_name=test-apps-create-with-name
  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh apps:create $test_app_name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku"

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote:list"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku"

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote:set dokku2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku2"

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote:list"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku"

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote:add dokku2 dokku@dokku.me:dokku2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote:list"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(printf "dokku\ndokku2")"

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote:remove dokku2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote:list"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku"

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote:unset"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote:unset"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "${BATS_TEST_DIRNAME}/../../contrib/dokku_client.sh remote"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "dokku"
}
