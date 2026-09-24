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

@test "(common) [preserve-env] fn-sudo-preserve-env-flag emits only the requested names" {
  run /bin/bash -c "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; fn-sudo-preserve-env-flag DOKKU_TRACE DOKKU_APP_NAME"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--preserve-env=DOKKU_TRACE,DOKKU_APP_NAME"
}

@test "(common) [preserve-env] fn-sudo-preserve-env-flag omits sudo managed names" {
  run /bin/bash -c "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; export DOKKU_TEST_CROSSING=1; fn-sudo-preserve-env-flag | sed 's/^--preserve-env=//' | tr ',' '\\n' | grep -cx DOKKU_TEST_CROSSING"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; fn-sudo-preserve-env-flag | sed 's/^--preserve-env=//' | tr ',' '\\n' | grep -cxE 'HOME|LOGNAME|MAIL|PATH|SHELL|USER|_|SUDO_.*'"
  echo "output: $output"
  echo "status: $status"
  assert_output "0"
}

@test "(common) [preserve-env] fn-sudo-preserve-env-flag honors DOKKU_PRESERVE_ENV" {
  run /bin/bash -c "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; DOKKU_PRESERVE_ENV=ALPHA,BETA fn-sudo-preserve-env-flag DOKKU_TRACE"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--preserve-env=DOKKU_TRACE,ALPHA,BETA"

  run /bin/bash -c "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; DOKKU_PRESERVE_ENV=ALPHA fn-sudo-preserve-env-flag | sed 's/^--preserve-env=//' | tr ',' '\\n' | grep -cx ALPHA"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"
}

@test "(common) [preserve-env] fn-cli-preserved-env-names covers every parse_args export" {
  run /bin/bash -c "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; fn-cli-preserved-env-names | tr ' ' '\\n' | grep -cxE 'DOKKU_APP_NAME|DOKKU_APPS_FORCE_DELETE|DOKKU_GLOBAL_FLAGS|DOKKU_PRESERVE_ENV|DOKKU_QUIET_OUTPUT|DOKKU_TRACE|SSH_USER'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "7"
}

@test "(common) [preserve-env] fn-cli-preserved-env-names covers the docker image build variables" {
  # neither is read by the dokku script itself, so the derivation test below does not
  # catch them. both exist only in the environment while the docker image is built
  run /bin/bash -c "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; fn-cli-preserved-env-names | tr ' ' '\\n' | grep -cxE 'DOKKU_INIT_SYSTEM|DOKKU_LIB_HOST_ROOT'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"
}

@test "(common) [preserve-env] every environment-overridable dokku variable is preserved" {
  # test_helper leaves these unexported, so the single-quoted script below cannot see
  # them without this
  export PLUGIN_CORE_AVAILABLE_PATH
  export DOKKU_SOURCE="${BATS_TEST_DIRNAME}/../../dokku"
  run /bin/bash -c 'source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"; comm -23 <(grep -oE "[{][A-Z_]+:=" "$DOKKU_SOURCE" | tr -cd "A-Z_\n" | sort -u | grep -vx DOKKU_PID) <(fn-cli-preserved-env-names | tr " " "\n" | sort -u)'
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""
  unset DOKKU_SOURCE
}

@test "(common) [preserve-env] parse_args records global flags for replay" {
  run /bin/bash -c "source '$PLUGIN_CORE_AVAILABLE_PATH/common/functions'; parse_args --label=a=b run foo env; echo \"\$DOKKU_GLOBAL_FLAGS\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--label=a=b"
}

@test "(core) cleanup missing app" {
  run /bin/bash -c "dokku cleanup missing-app"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "App missing-app does not exist"
}
