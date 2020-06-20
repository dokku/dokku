#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  DOCKERFILE="$BATS_TMPDIR/Dockerfile"
}

teardown() {
  rm -rf /home/dokku/$TEST_APP/tls
  destroy_app
  dokku config:unset --global DOKKU_RM_CONTAINER
  rm -f "$DOCKERFILE"
  global_teardown
}

@test "(core) run (with DOKKU_RM_CONTAINER/--rm-container)" {
  deploy_app

  run /bin/bash -c "dokku --rm-container run $TEST_APP echo $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "docker ps -a -f 'status=exited' --no-trunc=true | grep \"/exec echo $TEST_APP\""
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP DOKKU_RM_CONTAINER=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --rm-container run $TEST_APP echo $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "docker ps -a -f 'status=exited' --no-trunc=true | grep \"/exec echo $TEST_APP\""
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku config:unset --no-restart $TEST_APP DOKKU_RM_CONTAINER"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set --global DOKKU_RM_CONTAINER=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --rm-container run $TEST_APP echo $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "docker ps -a -f 'status=exited' --no-trunc=true | grep \"/exec echo $TEST_APP\""
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku config:unset --global DOKKU_RM_CONTAINER"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(core) run (detached)" {
  deploy_app

  RANDOM_RUN_CID="$(dokku --label=com.dokku.test-label=value --detach run $TEST_APP sleep 300)"
  run /bin/bash -c "docker inspect -f '{{ .State.Status }}' $RANDOM_RUN_CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "running"

  run /bin/bash -c "docker inspect $RANDOM_RUN_CID --format '{{ index .Config.Labels \"com.dokku.test-label\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_output "value"

  run /bin/bash -c "docker stop $RANDOM_RUN_CID"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku cleanup"
  echo "output: $output"
  echo "status: $status"
  assert_success
  sleep 5  # wait for dokku cleanup to happen in the background
}

@test "(core) run (with tty)" {
  deploy_app
  run /bin/bash -c "dokku run $TEST_APP ls /app/package.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(core) run (without tty)" {
  deploy_app
  run /bin/bash -c ": |dokku run $TEST_APP ls /app/package.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(core) run command from Procfile" {
  deploy_app
  run /bin/bash -c "dokku run $TEST_APP custom 'hi dokku' | tail -n 1"
  echo "output: $output"
  echo "status: $status"

  assert_success
  assert_output 'hi dokku'
}
