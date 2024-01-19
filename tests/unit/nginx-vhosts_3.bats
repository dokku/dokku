#!/usr/bin/env bats

load test_helper
source "$PLUGIN_CORE_AVAILABLE_PATH/config/functions"

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
  docker build "${BATS_TEST_DIRNAME}/../../tests/apps/gogrpc" -t ${TEST_APP}-docker-image
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) grpc endpoint" {
  run deploy_app gogrpc
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ports:add $TEST_APP grpc:80:50051"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker run --rm ${TEST_APP}-docker-image /go/bin/greeter_client -addr ${TEST_APP}.${DOKKU_DOMAIN}:80 -name grpc"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Greeting: Hello grpc"
}

@test "(nginx-vhosts) grpc endpoint on a port other than 80" {
  run deploy_app gogrpc
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ports:add $TEST_APP grpc:8080:50051"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker run --rm ${TEST_APP}-docker-image /go/bin/greeter_client -addr ${TEST_APP}.${DOKKU_DOMAIN}:8080 -name grpc8080"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Greeting: Hello grpc8080"
}

@test "(nginx-vhosts) grpcs endpoint" {
  setup_test_tls
  run deploy_app gogrpc
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ports:add $TEST_APP grpcs:443:50051"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker run --rm ${TEST_APP}-docker-image /go/bin/greeter_client -addr ${TEST_APP}.${DOKKU_DOMAIN}:443 -name grpcs -tls"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Greeting: Hello grpcs"
}
