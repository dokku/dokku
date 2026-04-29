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

@test "(resource) resource:limit --process-type build (herokuish)" {
  run /bin/bash -c "dokku resource:limit --memory 256m --cpu 1 --memory-swap 512m --nvidia-gpu 1 --process-type build $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku resource:report --resource-build.limit.memory $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output "256m"

  run /bin/bash -c "PLUGIN_PATH=$PLUGIN_PATH PLUGIN_CORE_AVAILABLE_PATH=$PLUGIN_CORE_AVAILABLE_PATH DOKKU_LIB_ROOT=$DOKKU_LIB_ROOT plugn trigger docker-args-process-build $TEST_APP herokuish < /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--memory=256m"
  assert_output_contains "--cpus=1"
  assert_output_contains "--memory-swap=512m"
  assert_output_contains "--gpus=1"
}

@test "(resource) resource:limit --process-type build (dockerfile filters cpu and gpu)" {
  run /bin/bash -c "dokku resource:limit --memory 512m --cpu 1 --nvidia-gpu 1 --memory-swap 1g --process-type build $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "PLUGIN_PATH=$PLUGIN_PATH PLUGIN_CORE_AVAILABLE_PATH=$PLUGIN_CORE_AVAILABLE_PATH DOKKU_LIB_ROOT=$DOKKU_LIB_ROOT plugn trigger docker-args-process-build $TEST_APP dockerfile < /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--memory=512m"
  assert_output_contains "--memory-swap=1g"
  assert_output_not_contains "--cpus="
  assert_output_not_contains "--gpus="
}

@test "(resource) resource:limit --process-type build (pack/lambda emit nothing)" {
  run /bin/bash -c "dokku resource:limit --memory 512m --process-type build $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "PLUGIN_PATH=$PLUGIN_PATH PLUGIN_CORE_AVAILABLE_PATH=$PLUGIN_CORE_AVAILABLE_PATH DOKKU_LIB_ROOT=$DOKKU_LIB_ROOT plugn trigger docker-args-process-build $TEST_APP pack < /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "--memory"

  run /bin/bash -c "PLUGIN_PATH=$PLUGIN_PATH PLUGIN_CORE_AVAILABLE_PATH=$PLUGIN_CORE_AVAILABLE_PATH DOKKU_LIB_ROOT=$DOKKU_LIB_ROOT plugn trigger docker-args-process-build $TEST_APP lambda < /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "--memory"
}

@test "(resource) resource:limit defaults do not leak into build trigger" {
  run /bin/bash -c "dokku resource:limit --memory 128m $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "PLUGIN_PATH=$PLUGIN_PATH PLUGIN_CORE_AVAILABLE_PATH=$PLUGIN_CORE_AVAILABLE_PATH DOKKU_LIB_ROOT=$DOKKU_LIB_ROOT plugn trigger docker-args-process-build $TEST_APP herokuish < /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "--memory"
}

@test "(resource) resource:limit reservations do not apply at build time" {
  run /bin/bash -c "dokku resource:reserve --memory 128m --process-type build $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "PLUGIN_PATH=$PLUGIN_PATH PLUGIN_CORE_AVAILABLE_PATH=$PLUGIN_CORE_AVAILABLE_PATH DOKKU_LIB_ROOT=$DOKKU_LIB_ROOT plugn trigger docker-args-process-build $TEST_APP herokuish < /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "--memory-reservation"
  assert_output_not_contains "--memory"
}

@test "(resource) DOKKU_OMIT_RESOURCE_ARGS suppresses build trigger" {
  run /bin/bash -c "dokku resource:limit --memory 256m --process-type build $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "PLUGIN_PATH=$PLUGIN_PATH PLUGIN_CORE_AVAILABLE_PATH=$PLUGIN_CORE_AVAILABLE_PATH DOKKU_LIB_ROOT=$DOKKU_LIB_ROOT DOKKU_OMIT_RESOURCE_ARGS=1 plugn trigger docker-args-process-build $TEST_APP herokuish < /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "--memory"
}

@test "(resource) resource:limit-clear --process-type build" {
  run /bin/bash -c "dokku resource:limit --memory 256m --process-type build $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku resource:report --resource-build.limit.memory $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output "256m"

  run /bin/bash -c "dokku resource:limit-clear --process-type build $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku resource:report --resource-build.limit.memory $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
