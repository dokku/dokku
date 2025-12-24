#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

@test "(builder-lambda:set)" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected lambda"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app lambda-python dokku@$DOKKU_DOMAIN:$TEST_APP inject_lambda_yml
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'Building app with image mlupin/docker-lambda:python3.9-build'
  assert_output_contains 'Installing dependencies via pip'

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Hello World" 0
  assert_success

  run /bin/bash -c "curl -d {} --silent --write-out '%{http_code}\n' $(dokku url $TEST_APP)/2015-03-31/functions/function.handler/invocations | grep 200"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Hello World"
  assert_success

  run /bin/bash -c "dokku builder-lambda:set $TEST_APP lambdayml-path nonexistent.yml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'Building app with image mlupin/docker-lambda:python3.9-build'
  assert_output_contains 'Installing dependencies via pip'

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Hello World" 0
  assert_success

  run /bin/bash -c "curl -d {} --silent --write-out '%{http_code}\n' $(dokku url $TEST_APP)/2015-03-31/functions/function.handler/invocations | grep 200"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Hello World"
  assert_success

  run /bin/bash -c "dokku builder-lambda:set $TEST_APP lambdayml-path lambda2.yml"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'Building app with image mlupin/docker-lambda:python3.9-build'
  assert_output_contains 'Installing dependencies via pip'

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Hello World" 0
  assert_success

  run /bin/bash -c "curl -d {} --silent --write-out '%{http_code}\n' $(dokku url $TEST_APP)/2015-03-31/functions/function.handler/invocations | grep 200"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Hello World"
  assert_success

  run /bin/bash -c "dokku builder-lambda:set $TEST_APP lambdayml-path"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # cache will be used
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains 'Building app with image mlupin/docker-lambda:python3.9-build'
  assert_output_contains 'Installing dependencies via pip' 0

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Hello World" 0
  assert_success

  run /bin/bash -c "curl -d {} --silent --write-out '%{http_code}\n' $(dokku url $TEST_APP)/2015-03-31/functions/function.handler/invocations | grep 200"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Hello World"
  assert_success
}

@test "(builder-lambda) run" {
  run /bin/bash -c "dokku config:set $TEST_APP SECRET_KEY=fjdkslafjdk"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder:set $TEST_APP selected lambda"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app lambda-python dokku@$DOKKU_DOMAIN:$TEST_APP inject_lambda_yml
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:inspect $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP env"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "SECRET_KEY=fjdkslafjdk"
}

inject_lambda_yml() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "-----> Injecting lambda.yml"
  echo "---" >>"$APP_REPO_DIR/lambda.yml"

  echo "-----> Injecting lambda.yml"
  echo "---" >>"$APP_REPO_DIR/lambda.yml"

  echo "-----> Injecting lambda2.yml"
  echo "---" >>"$APP_REPO_DIR/lambda2.yml"
  echo "build_image: mlupin/docker-lambda:python3.9-build" >>"$APP_REPO_DIR/lambda2.yml"
  echo "run_image: mlupin/docker-lambda:python3.9" >>"$APP_REPO_DIR/lambda2.yml"
}
