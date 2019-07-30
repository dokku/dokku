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

@test "(app-json) app.json scripts" {
  deploy_app nodejs-express

  run /bin/bash -c "dokku run $TEST_APP ls /app/prebuild.test"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku run $TEST_APP ls /app/predeploy.test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  CID=$(docker ps -a -q  -f "ancestor=dokku/${TEST_APP}" -f "label=dokku_phase_script=postdeploy")
  DOCKER_COMMIT_LABEL_ARGS=("--change" "LABEL org.label-schema.schema-version=1.0" "--change" "LABEL org.label-schema.vendor=dokku" "--change" "LABEL com.dokku.app-name=$TEST_APP")
  IMAGE_ID=$(docker commit "${DOCKER_COMMIT_LABEL_ARGS[@]}" $CID dokku-test/${TEST_APP})
  run /bin/bash -c "docker run -ti $IMAGE_ID ls /app/postdeploy.test"
  echo "output: $output"
  echo "status: $status"
  assert_success
}


@test "(app-json) app.json scripts missing" {
  deploy_app nodejs-express-noappjson
  echo "output: $output"
  echo "status: $status"
  assert_success
}
