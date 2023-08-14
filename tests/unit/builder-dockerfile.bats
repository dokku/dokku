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
