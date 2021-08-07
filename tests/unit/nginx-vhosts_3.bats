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
  destroy_app 0 $TEST_APP
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts) grpc endpoint" {
  deploy_app gogrpc
  dokku proxy:ports-add "$TEST_APP" "grpc:80:50051"
  run /bin/bash -c "docker run --rm ${TEST_APP}-docker-image /go/bin/greeter_client -address ${TEST_APP}.dokku.me:80 -name grpc"
  assert_output "Greeting: Hello grpc"
}

@test "(nginx-vhosts) grpc endpoint on a port other than 80" {
  deploy_app gogrpc
  dokku proxy:ports-add "$TEST_APP" "grpc:8080:50051"
  run /bin/bash -c "docker run --rm ${TEST_APP}-docker-image /go/bin/greeter_client -address ${TEST_APP}.dokku.me:8080 -name grpc8080"
  assert_output "Greeting: Hello grpc8080"
}

@test "(nginx-vhosts) grpcs endpoint" {
  setup_test_tls
  deploy_app gogrpc
  dokku proxy:ports-add "$TEST_APP" "grpcs:443:50051"
  run /bin/bash -c "docker run --rm ${TEST_APP}-docker-image /go/bin/greeter_client -address ${TEST_APP}.dokku.me:443 -name grpcs -tls"
  assert_output "Greeting: Hello grpcs"
}
