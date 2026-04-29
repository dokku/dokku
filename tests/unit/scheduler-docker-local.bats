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
  assert_output_contains "Deploys may fail when publishing ports and scaling to multiple containers" 0
  assert_output_contains "Deploys may fail when publishing ports and enabling zero downtime" 0

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy '--publish 5000:5000'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale --skip-deploy $TEST_APP web=2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # the expected output will be seen twice due to how parallel re-outputs stderr in its own output...
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Deploys may fail when publishing ports and scaling to multiple containers" 2
  assert_output_contains "Deploys may fail when publishing ports and enabling zero downtime" 0

  run /bin/bash -c "dokku ps:scale --skip-deploy $TEST_APP web=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Deploys may fail when publishing ports and scaling to multiple containers" 0
  assert_output_contains "Deploys may fail when publishing ports and enabling zero downtime" 2

  run /bin/bash -c "dokku checks:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Deploys may fail when publishing ports and scaling to multiple containers" 0
  assert_output_contains "Deploys may fail when publishing ports and enabling zero downtime" 0
}

@test "(scheduler-docker-local) sends SIGTERM immediately on deploy" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku checks:set $TEST_APP wait-to-retire 600"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  local old_cid
  old_cid="$(docker ps --filter label=com.dokku.app-name=$TEST_APP --filter label=com.dokku.process-type=web --format '{{.ID}}')"
  [[ -n "$old_cid" ]]

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 5

  run /bin/bash -c "docker inspect --format '{{.State.Status}}' $old_cid"
  echo "output: $output"
  echo "status: $status"
  assert_output "exited"

  run /bin/bash -c "docker ps --filter label=com.dokku.app-name=$TEST_APP --filter label=com.dokku.process-type=web --format '{{.Names}}'"
  echo "output: $output"
  echo "status: $status"
  assert_output "$TEST_APP.web.1"
}

@test "(scheduler-docker-local) scale down retires orphaned containers" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku checks:set $TEST_APP wait-to-retire 1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:scale $TEST_APP web=2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 2

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker ps --filter label=com.dokku.app-name=$TEST_APP --filter label=com.dokku.process-type=web --format '{{.Names}}' | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_output "2"

  run /bin/bash -c "dokku ps:scale $TEST_APP web=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 2

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker ps --filter label=com.dokku.app-name=$TEST_APP --filter label=com.dokku.process-type=web --format '{{.Names}}'"
  echo "output: $output"
  echo "status: $status"
  assert_output "$TEST_APP.web.1"
}

@test "(scheduler-docker-local) ps:rebuild with image-based deploy keeps running image" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku checks:set $TEST_APP wait-to-retire 1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP cockpithq/cockpit:core-latest"
  echo "output: $output"
  echo "status: $status"
  assert_success

  local running_image
  running_image="$(docker container inspect "$TEST_APP.web.1" --format '{{.Image}}' | cut -d: -f2)"
  [[ -n "$running_image" ]]

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 3

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "has running containers, skipping rm" 0

  run /bin/bash -c "grep -F \"$running_image\" /var/lib/dokku/data/scheduler-docker-local/dead-images 2>/dev/null || true"
  echo "output: $output"
  echo "status: $status"
  assert_output ""

  run /bin/bash -c "docker container inspect $TEST_APP.web.1 --format '{{.State.Status}}'"
  echo "output: $output"
  echo "status: $status"
  assert_output "running"
}
