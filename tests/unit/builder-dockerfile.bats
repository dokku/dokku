#!/usr/bin/env bats

load test_helper

setup() {
  create_app
  DOCKERFILE="$BATS_TMPDIR/Dockerfile"
}

teardown() {
  rm -f "$DOCKERFILE"
  destroy_app
}

@test "(builder-dockerfile:set)" {
  run deploy_app dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-dockerfile:set $TEST_APP dockerfile-path nonexistent-dockerfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku builder-dockerfile:set $TEST_APP dockerfile-path second.Dockerfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'echo hi'

  run /bin/bash -c "dokku builder-dockerfile:set $TEST_APP dockerfile-path"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'echo hi' 0
}

@test "(builder-dockerfile) exec" {
  run /bin/bash -c "dokku builder-dockerfile:set $TEST_APP dockerfile-path exec.Dockerfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku app-json:set $TEST_APP appjson-path app.json-fake"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:set $TEST_APP procfile-path exec.Procfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP task"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Found 'task' in Procfile, running that command"
  assert_output_contains 'hi dokku'
}

@test "(builder-dockerfile) config export" {
  run /bin/bash -c "dokku config:set $TEST_APP GITHUB_TOKEN=custom-value"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:add $TEST_APP build '--build-arg GITHUB_TOKEN'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "TOKEN is: custom-value" 2
}

@test "(builder-dockerfile) port exposure (dockerfile raw port)" {
  source "$PLUGIN_CORE_AVAILABLE_PATH/builder-dockerfile/internal-functions"
  cat <<EOF >$DOCKERFILE
EXPOSE 3001/udp
EXPOSE 3003
EXPOSE  3000/tcp
EOF
  run fn-builder-dockerfile-get-ports-from-dockerfile $DOCKERFILE
  echo "output: $output"
  echo "status: $status"
  assert_output "3001/udp 3003 3000/tcp"
}

@test "(builder-dockerfile) port exposure (dockerfile tcp port)" {
  source "$PLUGIN_CORE_AVAILABLE_PATH/builder-dockerfile/internal-functions"
  cat <<EOF >$DOCKERFILE
EXPOSE 3001/udp
EXPOSE  3000/tcp
EXPOSE 3003
EOF
  run fn-builder-dockerfile-get-ports-from-dockerfile $DOCKERFILE
  echo "output: $output"
  echo "status: $status"
  assert_output "3001/udp 3000/tcp 3003"
}

@test "(builder-dockerfile) ps:rebuild fetches files from image" {
  run /bin/bash -c "dokku --trace git:from-image $TEST_APP dokku/smoke-test-app:dockerfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(builder-dockerfile) run" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected dockerfile"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP convert_to_dockerfile
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP python task.py test"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "['task.py', 'test']"

  run /bin/bash -c "dokku run $TEST_APP task"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "['task.py', 'test']"

  run /bin/bash -c "dokku run $TEST_APP env"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "SECRET_KEY=fjdkslafjdk"
}
