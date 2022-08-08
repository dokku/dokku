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

@test "(scheduler-docker-local) complex labels" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  # the run command is equivalent to the following line, except with backslashes due to the enclosing doublequotes
  # dokku docker-options:add test deploy '--label "some.key=Host(\`$TEST_APP.dokku.me\`)"'
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy '--label \"some.key=Host(\\\`$TEST_APP.dokku.me\\\`)\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect --format '{{ index .Config.Labels \"some.key\"}}' $TEST_APP.web.1"
  echo "output: $output"
  echo "status: $status"
  assert_output "Host(\`$TEST_APP.dokku.me\`)"
  assert_success
}
