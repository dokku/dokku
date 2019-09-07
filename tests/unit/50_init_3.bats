#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(init) dockerfile with tini" {
  deploy_app zombies-dockerfile-tini
  run ps ax
  echo "output: $output"
  assert_output_contains "<defunct>" "0"
}
