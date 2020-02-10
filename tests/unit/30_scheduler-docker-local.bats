#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  global_teardown
}

@test "(scheduler-docker-local) scheduler-docker-local:help" {
  run /bin/bash -c "dokku scheduler-docker-local"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the docker-local scheduler integration for an app"
  help_output="$output"

  run /bin/bash -c "dokku scheduler-docker-local:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage the docker-local scheduler integration for an app"
  assert_output "$help_output"
}

@test "(scheduler-docker-local) timer installed" {
  run /bin/bash -c "systemctl list-timers | grep dokku-retire"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
