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

@test "(init) dockerfile no tini" {
  deploy_app zombies-dockerfile-no-tini
  run ps auxf
  echo "output: $output"
  assert_output_contains "<defunct>" "0"
}
