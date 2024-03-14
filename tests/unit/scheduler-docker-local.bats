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
  # dokku docker-options:add test deploy '--label "some.key=Host(\`$TEST_APP.$DOKKU_DOMAIN\`)"'
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy '--label \"some.key=Host(\\\`$TEST_APP.$DOKKU_DOMAIN\\\`)\"'"
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
  assert_output "Host(\`$TEST_APP.$DOKKU_DOMAIN\`)"
  assert_success
}

@test "(scheduler-docker-local) no-web" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:set $TEST_APP procfile-path worker.Procfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Skipping web as it is missing from the current Procfile"
}

@test "(scheduler-docker-local) init-process" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect --format '{{.HostConfig.Init}}' $TEST_APP.web.1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "docker exec $TEST_APP.web.1 ps auxf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "docker-init"

  run /bin/bash -c "dokku scheduler-docker-local:set $TEST_APP init-process false"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect --format '{{.HostConfig.Init}}' $TEST_APP.web.1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "<nil>"

  run /bin/bash -c "docker exec $TEST_APP.web.1 ps auxf"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "docker-init" 0
}

@test "(scheduler-docker-local) publish ports" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Port publishing may not work as expected with multiple containers" 0
  assert_output_contains "Port publishing may not work as expected with zero downtime" 0

  run /bin/bash -c "dokku ps:scale $TEST_APP web=2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Port publishing may not work as expected with multiple containers" 1
  assert_output_contains "Port publishing may not work as expected with zero downtime" 0

  run /bin/bash -c "dokku ps:scale --skip-deploy $TEST_APP web=2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Port publishing may not work as expected with multiple containers" 0
  assert_output_contains "Port publishing may not work as expected with zero downtime" 1

  run /bin/bash -c "dokku checks:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Port publishing may not work as expected with multiple containers" 0
  assert_output_contains "Port publishing may not work as expected with zero downtime" 0
}
