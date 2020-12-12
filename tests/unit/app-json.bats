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

@test "(app-json) app.json scripts" {
  run deploy_app nodejs-express
  echo "output: $output"
  echo "status: $status"
  assert_success

  run docker inspect "${TEST_APP}.web.1" --format "{{json .Config.Cmd}}"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output '["/start","web"]'

  run /bin/bash -c "dokku --rm run $TEST_APP ls /app/prebuild.test"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku --rm run $TEST_APP ls /app/predeploy.test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  CID=$(docker ps -a -q  -f "ancestor=dokku/${TEST_APP}" -f "label=dokku_phase_script=postdeploy")
  DOCKER_COMMIT_LABEL_ARGS=("--change" "LABEL org.label-schema.schema-version=1.0" "--change" "LABEL org.label-schema.vendor=dokku" "--change" "LABEL com.dokku.app-name=$TEST_APP")
  IMAGE_ID=$(docker commit "${DOCKER_COMMIT_LABEL_ARGS[@]}" $CID dokku-test/${TEST_APP})
  run /bin/bash -c "docker run --rm -ti $IMAGE_ID ls /app/postdeploy.test"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(app-json) app.json scripts missing" {
  run deploy_app nodejs-express-noappjson
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(app-json) app.json dockerfile cmd" {
  run deploy_app dockerfile-procfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run docker inspect "dokku/${TEST_APP}:latest" --format "{{json .Config.Cmd}}"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output '["/bin/sh","-c","npm start"]'

  run docker inspect "dokku/${TEST_APP}:latest" --format "{{json .Config.Entrypoint}}"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output 'null'
}

@test "(app-json) app.json dockerfile entrypoint" {
  run deploy_app dockerfile-entrypoint dokku@dokku.me:$TEST_APP add_release_command
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "touch /app/release.test"
}

@test "(app-json) app.json dockerfile release" {
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP SECRET_KEY=fjdkslafjdk ENVIRONMENT=dev DATABASE_URL=sqlite:///db.sqlite3"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app dockerfile-release
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Executing release task from Procfile"
  assert_output_contains "SECRET_KEY: fjdkslafjdk"
  assert_success
}

@test "(app-json) app.json dockerfile entrypoint" {
  run deploy_app dockerfile-entrypoint
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Executing predeploy task from app.json"
  assert_output_contains "entrypoint script started with arguments touch /app/predeploy.test"
  assert_success

  run /bin/bash -c "dokku --rm run $TEST_APP ls /app/predeploy.test"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
