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
  rm -f "$DOCKERFILE"
  global_teardown
}

@test "(core) remove exited containers" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  # make sure we have many exited containers of the same 'type'
  run /bin/bash -c "for cnt in 1 2 3; do dokku run $TEST_APP echo $TEST_APP; done"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "docker ps -a -f 'status=exited' --no-trunc=true | grep \"/exec echo $TEST_APP\""
  echo "output: $output"
  echo "status: $status"
  assert_failure

  RANDOM_RUN_CID="$(docker run -d gliderlabs/herokuish bash)"
  docker ps -a
  run /bin/bash -c "dokku cleanup"
  echo "output: $output"
  echo "status: $status"
  assert_success
  sleep 5 # wait for dokku cleanup to happen in the background

  run /bin/bash -c "docker inspect $RANDOM_RUN_CID"
  echo "output: $output"
  echo "status: $status"
  assert_success
  docker rm $RANDOM_RUN_CID
}

@test "(core) image type detection (herokuish default user)" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:trigger builder-image-is-herokuish $TEST_APP dokku/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output "true"
}

@test "(core) image type detection (herokuish custom user)" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  CID=$(<"$DOKKU_ROOT/$TEST_APP/CONTAINER.web.1")
  docker commit --change "ENV USER postgres" "$CID" "dokku/${TEST_APP}:latest"
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP DOKKU_APP_USER=postgres"
  echo "output: $output"
  echo "status: $status"
  assert_success

  source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
  source "$PLUGIN_CORE_AVAILABLE_PATH/config/functions"

  run /bin/bash -c "dokku plugin:trigger builder-image-is-herokuish $TEST_APP dokku/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output "true"
}

@test "(core) image type detection (dockerfile)" {
  run deploy_app dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:trigger builder-image-is-herokuish $TEST_APP dokku/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output "false"
}

@test "(common) fn-docker-args-split does not expand injected commands" {
  export DOKKU_TEST_PAYLOAD='--label x=$(id)'
  run /bin/bash -c "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; fn-docker-args-split \"\$DOKKU_TEST_PAYLOAD\" | tr '\\0' '\\n'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 1 'x=$(id)'
  [[ "$output" != *"uid="* ]] || flunk "id command output leaked - value was expanded"

  export DOKKU_TEST_PAYLOAD='--label x=`id`'
  run /bin/bash -c "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; fn-docker-args-split \"\$DOKKU_TEST_PAYLOAD\" | tr '\\0' '\\n'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 1 'x=`id`'
  [[ "$output" != *"uid="* ]] || flunk "id command output leaked - value was expanded"

  unset DOKKU_TEST_PAYLOAD
}

@test "(common) fn-docker-args-split preserves quoted tokens" {
  run /bin/bash -c "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; fn-docker-args-split \"--label 'a b'\" | tr '\\0' '\\n'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line 0 "--label"
  assert_line 1 "a b"
}
