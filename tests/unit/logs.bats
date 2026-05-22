#!/usr/bin/env bats

load test_helper

setup() {
  rm "${BATS_PARENT_TMPNAME}.skip" || true
  global_setup
}

teardown() {
  destroy_app
  dokku logs:set --global vector-networks >/dev/null 2>/dev/null || true
  docker network rm test-vector-net-a >/dev/null || true
  docker network rm test-vector-net-b >/dev/null || true
  global_teardown
}

@test "(logs) logs:help" {
  run /bin/bash -c "dokku logs:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage log integration for an app"
}

@test "(logs) logs:report" {
  run /bin/bash -c "dokku logs:report"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "You haven't deployed any applications yet"

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:report 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$TEST_APP logs information"
}

@test "(logs) logs:report app" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:report $TEST_APP 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$TEST_APP logs information"

  run /bin/bash -c "dokku logs:report $TEST_APP --invalid-flag 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "$TEST_APP logs information" 0
  assert_output_contains "Invalid flag passed, valid flags: --logs-app-label-alias, --logs-computed-app-label-alias, --logs-computed-max-size, --logs-global-app-label-alias, --logs-global-max-size, --logs-global-vector-sink, --logs-max-size, --logs-vector-global-image, --logs-vector-global-networks, --logs-vector-sink"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$TEST_APP logs information" 0
  assert_output_contains "Invalid flag passed" 0

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$TEST_APP logs information" 0
  assert_output_contains "Invalid flag passed" 0
}

@test "(logs) logs:set [error]" {
  run /bin/bash -c "dokku logs:set 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  assert_output_contains "Please specify an app to run the command on"
  run /bin/bash -c "dokku logs:set ${TEST_APP}-non-existent" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "App $TEST_APP-non-existent does not exist"

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "No property specified"

  run /bin/bash -c "dokku logs:set $TEST_APP invalid" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid property specified, valid properties include: app-label-alias, max-size, vector-image, vector-networks, vector-sink"

  run /bin/bash -c "dokku logs:set $TEST_APP invalid value" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid property specified, valid properties include: app-label-alias, max-size, vector-image, vector-networks, vector-sink"

  run /bin/bash -c "dokku logs:set $TEST_APP vector-image timberio/vector:latest-debian 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "vector-image may only be set globally with --global"
}

@test "(logs) logs:set app" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink console://?encoding[codec]=json" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "console://?encoding[codec]=json"

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink datadog_logs://?api_key=abc123" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "datadog_logs://?api_key=abc123"

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting vector-sink"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:set $TEST_APP max-size" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting max-size"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:set $TEST_APP max-size 20m" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting max-size"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "20m"

  run /bin/bash -c "dokku logs:set $TEST_APP max-size unlimited" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting max-size"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "unlimited"

  run /bin/bash -c "dokku logs:set "$TEST_APP" max-size" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting max-size"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
}

@test "(logs) logs:set equals in uri" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink 'loki://?endpoint=https://host&encoding[codec]=text&auth[token]=foobar%3D&auth[strategy]=bearer'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"
  # type: loki
  # endpoint: https://host
  # encoding[codec]: text
  # auth[token]: foobar=
  # auth[strategy]: bearer

  run /bin/bash -c "jq -r '.sinks[\"docker-sink:$TEST_APP\"].auth.strategy' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "bearer"

  run /bin/bash -c "jq -r '.sinks[\"docker-sink:$TEST_APP\"].auth.token' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "foobar="

  run /bin/bash -c "jq -r '.sinks[\"docker-sink:$TEST_APP\"].endpoint' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "https://host"

  run /bin/bash -c "jq -r '.sinks[\"docker-sink:$TEST_APP\"].encoding.codec' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "text"
}

@test "(logs) logs:set escaped uri" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink http://?uri=https%3A//loggerservice.com%3A1234/%3Ftoken%3Dabc1234%26type%3Dvector%26key%3Dvalue%2Bvalue2"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "http://?uri=https%3A//loggerservice.com%3A1234/%3Ftoken%3Dabc1234%26type%3Dvector%26key%3Dvalue%2Bvalue2"

  run /bin/bash -c "jq -r '.sinks[\"docker-sink:$TEST_APP\"].uri' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "https://loggerservice.com:1234/?token=abc1234&type=vector&key=value+value2"

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink 'aws_cloudwatch_logs://?create_missing_group=true&create_missing_stream=true&group_name=groupname&encoding[codec]=json&region=sa-east-1&stream_name={{ host }}&auth[access_key_id]=KSDSIDJSAJD&auth[secret_access_key]=2932JSDJ%252BKSDSDJ'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "aws_cloudwatch_logs://?create_missing_group=true&create_missing_stream=true&group_name=groupname&encoding[codec]=json&region=sa-east-1&stream_name={{ host }}&auth[access_key_id]=KSDSIDJSAJD&auth[secret_access_key]=2932JSDJ%252BKSDSDJ"

  run /bin/bash -c "jq -r '.sinks[\"docker-sink:$TEST_APP\"].auth.secret_access_key' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2932JSDJ+KSDSDJ"
}

@test "(logs) logs:set global" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:set --global vector-sink" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:set --global vector-sink console://?encoding[codec]=json" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "console://?encoding[codec]=json"

  run /bin/bash -c "dokku logs:set --global vector-sink datadog_logs://?api_key=abc123" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "datadog_logs://?api_key=abc123"

  run /bin/bash -c "dokku logs:set --global vector-sink" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:set --global vector-sink" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting vector-sink"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-vector-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-computed-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "10m"

  run /bin/bash -c "dokku logs:set --global max-size" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting max-size"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-computed-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "10m"

  run /bin/bash -c "dokku logs:set --global max-size 20m" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting max-size"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "20m"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-computed-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "20m"

  run /bin/bash -c "dokku logs:set --global max-size unlimited" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting max-size"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "unlimited"

  run /bin/bash -c "dokku logs:set --global max-size" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting max-size"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-computed-max-size 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "10m"
}

@test "(logs:report) vector-sink raw vs global" {
  run /bin/bash -c "dokku logs:set --global vector-sink"
  assert_success

  run /bin/bash -c "dokku --quiet logs:report $TEST_APP --format json | jq -r '.\"logs-vector-sink\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku --quiet logs:report $TEST_APP --format json | jq -r '.\"logs-global-vector-sink\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku logs:set --global vector-sink console://?encoding[codec]=json"
  assert_success

  run /bin/bash -c "dokku --quiet logs:report $TEST_APP --format json | jq -r '.\"logs-global-vector-sink\"'"
  assert_success
  assert_output "console://?encoding[codec]=json"

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink datadog_logs://?api_key=abc"
  assert_success

  run /bin/bash -c "dokku --quiet logs:report $TEST_APP --format json | jq -r '.\"logs-vector-sink\"'"
  assert_success
  assert_output "datadog_logs://?api_key=abc"

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink"
  assert_success

  run /bin/bash -c "dokku logs:set --global vector-sink"
  assert_success
}

@test "(logs:report) vector-global-image and vector-global-networks raw" {
  run /bin/bash -c "dokku logs:set --global vector-image"
  assert_success

  run /bin/bash -c "dokku --quiet logs:report --global --format json | jq -r '.\"logs-vector-global-image\"'"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku logs:set --global vector-image timberio/vector:custom"
  assert_success

  run /bin/bash -c "dokku --quiet logs:report --global --format json | jq -r '.\"logs-vector-global-image\"'"
  assert_success
  assert_output "timberio/vector:custom"

  run /bin/bash -c "dokku logs:set --global vector-image"
  assert_success

  run /bin/bash -c "dokku --quiet logs:report --global --format json | jq -r '.\"logs-vector-global-networks\"'"
  assert_success
  assert_output ""
}

@test "(logs) logs:set --global vector-networks" {
  docker network create test-vector-net-a >/dev/null
  docker network create test-vector-net-b >/dev/null

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-networks test-vector-net-a 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "vector-networks may only be set globally with --global"

  run /bin/bash -c "dokku logs:set --global vector-networks does-not-exist 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Network \"does-not-exist\" does not exist"

  run /bin/bash -c "dokku logs:set --global vector-networks bridge 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "\"bridge\" is not a valid entry for vector-networks"

  run /bin/bash -c "dokku logs:set --global vector-networks 'test-vector-net-a,'  2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "empty entry in comma-separated list"

  run /bin/bash -c "dokku logs:set --global vector-networks test-vector-net-a 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-networks"

  run /bin/bash -c "dokku logs:report --global --logs-vector-global-networks 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test-vector-net-a"

  run /bin/bash -c "dokku logs:set --global vector-networks test-vector-net-a,test-vector-net-b 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-networks"

  run /bin/bash -c "dokku logs:report --global --logs-vector-global-networks 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test-vector-net-a,test-vector-net-b"

  run /bin/bash -c "dokku logs:set --global vector-networks 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting vector-networks"

  run /bin/bash -c "dokku logs:report --global --logs-vector-global-networks 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
}

@test "(logs) logs:vector-start attaches configured networks" {
  docker network create test-vector-net-a >/dev/null
  docker network create test-vector-net-b >/dev/null

  run /bin/bash -c "dokku logs:set --global vector-networks test-vector-net-a,test-vector-net-b 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-start 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container is running"

  run /bin/bash -c "sudo docker inspect --format='{{range \$k, \$v := .NetworkSettings.Networks}}{{\$k}} {{end}}' vector-vector-1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "test-vector-net-a"
  assert_output_contains "test-vector-net-b"
  assert_output_contains "bridge" 0

  run /bin/bash -c "dokku logs:vector-stop 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-start 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container is running"

  run /bin/bash -c "sudo docker inspect --format='{{range \$k, \$v := .NetworkSettings.Networks}}{{\$k}} {{end}}' vector-vector-1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "test-vector-net-a"
  assert_output_contains "test-vector-net-b"
  assert_output_contains "bridge" 0

  run /bin/bash -c "dokku logs:set --global vector-networks 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-stop 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-start 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sudo docker inspect --format='{{range \$k, \$v := .NetworkSettings.Networks}}{{\$k}} {{end}}' vector-vector-1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "bridge"

  run /bin/bash -c "dokku logs:vector-stop 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(logs) logs:set app-label-alias" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink console://?encoding[codec]=json" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:set --global app-label-alias" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting app-label-alias"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "jq -r '.sources[\"docker-source:$TEST_APP\"].include_labels[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "com.dokku.app-name=$TEST_APP"

  run /bin/bash -c "dokku logs:set --global app-label-alias global-alt-name" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-computed-app-label-alias" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "global-alt-name"

  run /bin/bash -c "cat /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "jq -r '.sources[\"docker-source:$TEST_APP\"].include_labels[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "global-alt-name=$TEST_APP"

  run /bin/bash -c "dokku logs:set --global app-label-alias alt-name" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-computed-app-label-alias" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "alt-name"

  run /bin/bash -c "jq -r '.sources[\"docker-source:$TEST_APP\"].include_labels[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "alt-name=$TEST_APP"
}

@test "(logs) logs:set max-size with alternate log-driver daemon" {
  if [[ "$REMOTE_CONTAINERS" == "true" ]]; then
    skip "skipping due non-existent docker service in remote dev container"
  fi

  if [[ ! -f /etc/docker/daemon.json ]]; then
    echo "{}" >/etc/docker/daemon.json
  fi

  driver="$(jq -r '."log-driver"' /etc/docker/daemon.json)"
  local TMP_FILE=$(mktemp "/tmp/${DOKKU_DOMAIN}.XXXX")

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP max-size 20m 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting max-size"

  run /bin/bash -c "echo '' | dokku plugin:trigger docker-args-process-deploy $TEST_APP 2>&1 | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--log-opt=max-size=20m"

  DRIVER="journald" jq '."log-driver" = env.DRIVER' <"/etc/docker/daemon.json" >"$TMP_FILE"
  mv "$TMP_FILE" /etc/docker/daemon.json

  sudo service docker restart

  run /bin/bash -c "dokku logs:set $TEST_APP max-size 20m 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting max-size"

  run /bin/bash -c "echo '' | dokku plugin:trigger docker-args-process-deploy $TEST_APP 2>&1 | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  if [[ "$driver" = "null" ]]; then
    DRIVER="$driver" jq 'del(."log-driver")' <"/etc/docker/daemon.json" >"$TMP_FILE"
  else
    DRIVER="$driver" jq '."log-driver" = env.DRIVER' <"/etc/docker/daemon.json" >"$TMP_FILE"
  fi

  mv "$TMP_FILE" /etc/docker/daemon.json
  sudo service docker restart

  run /bin/bash -c "dokku logs:set $TEST_APP max-size 20m 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting max-size"

  run /bin/bash -c "echo '' | dokku plugin:trigger docker-args-process-deploy $TEST_APP 2>&1 | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--log-opt=max-size=20m"
}

@test "(logs) logs:set max-size with alternate log-driver" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP max-size 20m" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting max-size"

  run /bin/bash -c "echo "" | dokku plugin:trigger docker-args-process-deploy $TEST_APP 2>&1 | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--log-opt=max-size=20m"

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy --log-driver=local" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo "" | dokku plugin:trigger docker-args-process-deploy $TEST_APP 2>&1 | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--log-opt=max-size=20m"

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy --log-driver=json-file" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo "" | dokku plugin:trigger docker-args-process-deploy $TEST_APP 2>&1 | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--log-opt=max-size=20m"

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy --log-driver=journald" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo "" | dokku plugin:trigger docker-args-process-deploy $TEST_APP 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
}

@test "(logs) logs:vector" {
  run /bin/bash -c "dokku logs:vector-logs 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Vector container does not exist"

  run /bin/bash -c "dokku apps:create example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-start 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container is running"

  run /bin/bash -c "sudo docker inspect --format='{{.HostConfig.RestartPolicy.Name}}' vector-vector-1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "unless-stopped"

  run /bin/bash -c "dokku logs:vector-logs 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container logs"

  run /bin/bash -c "dokku --force apps:destroy example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-logs --num 10 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container logs"
  assert_output_contains "vector:" 10
  assert_line_count 11

  run /bin/bash -c "dokku logs:vector-logs --num 5 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container logs"
  assert_output_contains "vector:" 5
  assert_line_count 6

  run /bin/bash -c "docker stop vector-vector-1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-logs 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container logs"
  assert_output_contains "Vector container is not running"

  run /bin/bash -c "dokku logs:vector-stop 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Stopping and removing vector container"
}
