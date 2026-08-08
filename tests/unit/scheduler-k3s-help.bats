#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  global_teardown
}

@test "(scheduler-k3s) scheduler-k3s:help" {
  run /bin/bash -c "dokku scheduler-k3s"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage scheduler-k3s settings for an app"
  help_output="$output"

  run /bin/bash -c "dokku scheduler-k3s:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage scheduler-k3s settings for an app"
  assert_output "$help_output"

  assert_output_contains "scheduler-k3s:report [<app>|--global] [--format stdout|json] [<flag>]"
  assert_output_contains "scheduler-k3s:set <app|--global> <property> (<value>)"
}
