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
  deploy_app

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

@test "(core) port exposure (dockerfile raw port)" {
  source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
  cat <<EOF >$DOCKERFILE
EXPOSE 3001/udp
EXPOSE 3003
EXPOSE  3000/tcp
EOF
  run get_dockerfile_exposed_ports $DOCKERFILE
  echo "output: $output"
  echo "status: $status"
  assert_output "3001/udp 3003 3000/tcp"
}

@test "(core) port exposure (dockerfile tcp port)" {
  source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
  cat <<EOF >$DOCKERFILE
EXPOSE 3001/udp
EXPOSE  3000/tcp
EXPOSE 3003
EOF
  run get_dockerfile_exposed_ports $DOCKERFILE
  echo "output: $output"
  echo "status: $status"
  assert_output "3001/udp 3000/tcp 3003"
}

@test "(core) image type detection (herokuish default user)" {
  export DOKKU_ROOT
  deploy_app
  source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
  source "$PLUGIN_CORE_AVAILABLE_PATH/config/functions"

  run is_image_herokuish_based "dokku/$TEST_APP" "$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(core) image type detection (herokuish custom user)" {
  export DOKKU_ROOT
  deploy_app
  CID=$(<"$DOKKU_ROOT/$TEST_APP/CONTAINER.web.1")
  docker commit --change "ENV USER postgres" "$CID" "dokku/${TEST_APP}:latest"
  dokku config:set --no-restart "$TEST_APP" DOKKU_APP_USER=postgres
  source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
  source "$PLUGIN_CORE_AVAILABLE_PATH/config/functions"

  run is_image_herokuish_based "dokku/$TEST_APP" "$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(core) image type detection (dockerfile)" {
  export DOKKU_ROOT
  deploy_app dockerfile
  source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
  source "$PLUGIN_CORE_AVAILABLE_PATH/config/functions"

  run is_image_herokuish_based "dokku/$TEST_APP" "$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
