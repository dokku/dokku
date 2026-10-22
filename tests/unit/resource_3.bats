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

  run_plugn_trigger docker-args-process-build "$TEST_APP" herokuish
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

  run_plugn_trigger docker-args-process-build "$TEST_APP" dockerfile
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

  run_plugn_trigger docker-args-process-build "$TEST_APP" pack
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "--memory"

  run_plugn_trigger docker-args-process-build "$TEST_APP" lambda
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

  run_plugn_trigger docker-args-process-build "$TEST_APP" herokuish
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

  run_plugn_trigger docker-args-process-build "$TEST_APP" herokuish
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

  # DOKKU_OMIT_RESOURCE_ARGS has to survive the privilege drop in
  # plugins/plugin/subcommands/trigger, so this also covers that forwarding
  run_plugn_trigger DOKKU_OMIT_RESOURCE_ARGS=1 docker-args-process-build "$TEST_APP" herokuish
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

@test "(resource) resource:limit uppercase memory units" {
  run /bin/bash -c "dokku resource:limit --memory 1G --memory-swap 2G $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku resource:reserve --memory 512M $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run_plugin_script resource docker-args-process-deploy "$TEST_APP" herokuish "dokku/$TEST_APP:latest" web "< /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--memory=1G "
  assert_output_contains "--memory-swap=2G "
  assert_output_contains "--memory-reservation=512M "
  assert_output_not_contains "1Gm"
  assert_output_not_contains "512Mm"

  run /bin/bash -c "dokku resource:limit --memory 1G --process-type build $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run_plugn_trigger docker-args-process-build "$TEST_APP" herokuish
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "--memory=1G "
  assert_output_not_contains "1Gm"
}

@test "(resource) resource:limit rejects invalid memory values" {
  run /bin/bash -c "dokku resource:limit --memory '100G --foo' $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku resource:limit --memory 1x $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku resource:limit --memory-swap 1x $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku resource:reserve --memory 1Gm $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku resource:report --resource-_default_.limit.memory $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku resource:report --resource-_default_.reserve.memory $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku resource:limit --memory-swap -1 $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku resource:report --resource-_default_.limit.memory-swap $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "-1"
}
