#!/usr/bin/env bats

load test_helper

setup () {
  global_setup
}

teardown () {
  global_teardown
}

@test "(apps) apps:create" {
  run dokku apps:create $TEST_APP
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "dokku apps:list | grep $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_output $TEST_APP
  destroy_app

  run dokku apps:create 1994testapp
  echo "output: "$output
  echo "status: "$status
  assert_success

  run dokku apps:create testapp:latest
  echo "output: "$output
  echo "status: "$status
  assert_failure

  run dokku apps:create testApp:latest
  echo "output: "$output
  echo "status: "$status
  assert_failure

  run dokku apps:create TestApp
  echo "output: "$output
  echo "status: "$status
  assert_failure

  run bash -c "dokku --app $TEST_APP apps:create"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "dokku apps:list | grep $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_output $TEST_APP

  destroy_app
}

@test "(apps) app autocreate disabled" {
  run dokku config:set --no-restart --global DOKKU_DISABLE_APP_AUTOCREATION='true'
  echo "output: "$output
  echo "status: "$status
  assert_success

  run deploy_app
  echo "output: "$output
  echo "status: "$status
  assert_failure
  run dokku config:unset --no-restart --global DOKKU_DISABLE_APP_AUTOCREATION
}

@test "(apps) apps:destroy" {
  create_app
  run bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  create_app
  run bash -c "dokku --force --app $TEST_APP apps:destroy"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(apps) apps:rename" {
  deploy_app
  run bash -c "dokku apps:rename $TEST_APP great-test-name"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "dokku apps:list | grep $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_output ""
  run bash -c "curl --silent --write-out '%{http_code}\n' `dokku url great-test-name` | grep 404"
  echo "output: "$output
  echo "status: "$status
  assert_output ""
  run bash -c "dokku --force apps:destroy great-test-name"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run dokku apps:create $TEST_APP
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "dokku apps:rename $TEST_APP great-test-name"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "dokku --force apps:destroy great-test-name"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(apps) apps:clone" {
  deploy_app
  run bash -c "dokku apps:clone $TEST_APP great-test-name"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "dokku apps:list | grep $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "curl --silent --write-out '%{http_code}\n' `dokku url great-test-name` | grep 404"
  echo "output: "$output
  echo "status: "$status
  assert_output ""
  run bash -c "dokku --force apps:destroy great-test-name"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(apps) apps:clone --skip-deploy" {
  deploy_app
  run bash -c "dokku apps:clone --skip-deploy $TEST_APP great-test-name"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "curl --silent --write-out '%{http_code}\n' `dokku url great-test-name` | grep 404"
  echo "output: "$output
  echo "status: "$status
  assert_failure
  run bash -c "dokku --force apps:destroy great-test-name"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(apps) apps:clone --ignore-existing" {
  deploy_app
  run bash -c "dokku apps:create great-test-name"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "dokku apps:clone --ignore-existing $TEST_APP great-test-name"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "dokku --force apps:destroy great-test-name"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(apps) apps:exists" {
  run dokku apps:exists $TEST_APP
  echo "output: "$output
  echo "status: "$status
  assert_failure

  create_app
  run dokku apps:exists $TEST_APP
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

}

@test "(apps) apps:lock/unlock" {
  create_app

  run bash -c "dokku apps:report $TEST_APP --locked"
  echo "output: "$output
  echo "status: "$status
  assert_output "false"

  run bash -c "dokku apps:lock $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku apps:report $TEST_APP --locked"
  echo "output: "$output
  echo "status: "$status
  assert_output "true"

  run bash -c "dokku apps:unlock $TEST_APP"
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku apps:report $TEST_APP --locked"
  echo "output: "$output
  echo "status: "$status
  assert_output "false"

  destroy_app
}