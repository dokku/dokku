#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  global_teardown
}

@test "(apps) apps:help" {
  run /bin/bash -c "dokku apps"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage apps"
  help_output="$output"

  run /bin/bash -c "dokku apps:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage apps"
  assert_output "$help_output"
}

@test "(apps) apps:list" {
  run /bin/bash -c "dokku apps:list 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "You haven't deployed any applications yet"
  assert_output_contains "$TEST_APP" 0

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:list 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "You haven't deployed any applications yet" 0
  assert_output_contains "$TEST_APP"

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps) apps:create" {
  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku apps:list | grep $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output $TEST_APP
  destroy_app

  run /bin/bash -c "dokku apps:create 1994testapp"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:create testapp:latest"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create testApp:latest"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create TestApp"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create test/app"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create test_app"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku --app $TEST_APP apps:create"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku apps:list | grep $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output $TEST_APP

  destroy_app
}

@test "(apps) app autocreate disabled" {
  run /bin/bash -c "dokku config:set --no-restart --global DOKKU_DISABLE_APP_AUTOCREATION='true'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_failure
  run /bin/bash -c "dokku config:unset --no-restart --global DOKKU_DISABLE_APP_AUTOCREATION"
}

@test "(apps) apps:destroy" {
  create_app
  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  create_app
  run /bin/bash -c "dokku --force --app $TEST_APP apps:destroy"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps) apps:rename" {
  deploy_app
  run /bin/bash -c "dokku apps:rename $TEST_APP great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku apps:list | grep $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output ""
  run /bin/bash -c "curl --silent --write-out '%{http_code}\n' $(dokku url great-test-name) | grep 404"
  echo "output: $output"
  echo "status: $status"
  assert_output ""
  run /bin/bash -c "dokku --force apps:destroy great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku buildpacks:add $TEST_APP https://github.com/heroku/heroku-buildpack-ruby.git"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku git:set $TEST_APP deploy-branch SOME_BRANCH_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku network:set $TEST_APP attach-post-create test-network"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku nginx:set $TEST_APP  bind-address-ipv4 127.0.0.1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku resource:limit --memory 100 $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku scheduler-docker-local:set  $TEST_APP disable-chown true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:rename $TEST_APP great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku buildpacks:report great-test-name --buildpacks-list"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "https://github.com/heroku/heroku-buildpack-ruby.git"
  run /bin/bash -c "dokku git:report great-test-name --git-deploy-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "SOME_BRANCH_NAME"
  run /bin/bash -c "dokku network:report great-test-name --network-attach-post-create"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test-network"
  run /bin/bash -c "dokku nginx:report great-test-name --nginx-bind-address-ipv4"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "127.0.0.1"
  run /bin/bash -c "dokku resource:report great-test-name --resource-_default_.limit.memory"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "100"
  run /bin/bash -c "dokku scheduler-docker-local:report great-test-name --scheduler-docker-local-disable-chown"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku --force apps:destroy great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps) apps:rename with tls" {
  setup_test_tls
  deploy_app
  run /bin/bash -c "dokku apps:rename $TEST_APP great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku --force apps:destroy great-test-name"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(apps) apps:exists" {
  run /bin/bash -c "dokku apps:exists $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  create_app
  run /bin/bash -c "dokku apps:exists $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

}

@test "(apps) apps:lock/locked/unlock" {
  create_app

  run /bin/bash -c "dokku apps:report $TEST_APP --app-locked"
  echo "output: $output"
  echo "status: $status"
  assert_output "false"

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:lock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:report $TEST_APP --app-locked"
  echo "output: $output"
  echo "status: $status"
  assert_output "true"

  run /bin/bash -c "dokku apps:locked $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:unlock $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:report $TEST_APP --app-locked"
  echo "output: $output"
  echo "status: $status"
  assert_output "false"

  destroy_app
}
