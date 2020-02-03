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

@test "(init) buildpack" {
  deploy_app zombies-buildpack
  run ps ax
  echo "output: $output"
  assert_output_contains "<defunct>" "0"
}
