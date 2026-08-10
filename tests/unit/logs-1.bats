#!/usr/bin/env bats

load test_helper

setup() {
  rm "${BATS_PARENT_TMPNAME}.skip" || true
  global_setup
}

teardown() {
  destroy_app
  # a leftover app-label-alias makes every later test's vector source filter on
  # a label that dokku never applies to a container, which silently disables log
  # collection for the rest of the file
  dokku logs:set --global app-label-alias >/dev/null 2>/dev/null || true
  dokku logs:set --global vector-image >/dev/null 2>/dev/null || true
  dokku logs:set --global vector-networks >/dev/null 2>/dev/null || true
  dokku logs:set --global vector-sink >/dev/null 2>/dev/null || true
  dokku logs:set --global vector-cron-sink >/dev/null 2>/dev/null || true
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
  assert_output_contains "Invalid flag passed, valid flags: --logs-app-label-alias, --logs-computed-app-label-alias, --logs-computed-max-size, --logs-computed-vector-cron-sink, --logs-computed-vector-image, --logs-computed-vector-networks, --logs-computed-vector-sink, --logs-global-app-label-alias, --logs-global-max-size, --logs-global-vector-cron-sink, --logs-global-vector-image, --logs-global-vector-networks, --logs-global-vector-sink, --logs-max-size, --logs-vector-cron-sink, --logs-vector-sink"

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

@test "(logs) logs:report --global invalid flag" {
  run /bin/bash -c "dokku logs:report --global --invalid-flag 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid flag passed, valid flags: --logs-computed-app-label-alias, --logs-computed-max-size, --logs-computed-vector-cron-sink, --logs-computed-vector-image, --logs-computed-vector-networks, --logs-computed-vector-sink, --logs-global-app-label-alias, --logs-global-max-size, --logs-global-vector-cron-sink, --logs-global-vector-image, --logs-global-vector-networks, --logs-global-vector-sink"
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
  assert_output_contains "Invalid property specified, valid properties include: app-label-alias, max-size, vector-cron-sink, vector-image, vector-networks, vector-sink"

  run /bin/bash -c "dokku logs:set $TEST_APP invalid value" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid property specified, valid properties include: app-label-alias, max-size, vector-cron-sink, vector-image, vector-networks, vector-sink"

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
  run create_app
  assert_success

  run /bin/bash -c "dokku logs:set --global vector-sink"
  assert_success

  run /bin/bash -c "dokku logs:report $TEST_APP --format json | jq -r '.\"logs-vector-sink\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku logs:report $TEST_APP --format json | jq -r '.\"logs-global-vector-sink\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku logs:set --global vector-sink console://?encoding[codec]=json"
  assert_success

  run /bin/bash -c "dokku logs:report $TEST_APP --format json | jq -r '.\"logs-global-vector-sink\"'"
  assert_success
  assert_output "console://?encoding[codec]=json"

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink datadog_logs://?api_key=abc"
  assert_success

  run /bin/bash -c "dokku logs:report $TEST_APP --format json | jq -r '.\"logs-vector-sink\"'"
  assert_success
  assert_output "datadog_logs://?api_key=abc"

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink"
  assert_success

  run /bin/bash -c "dokku logs:set --global vector-sink"
  assert_success
}

@test "(logs:set) vector-cron-sink" {
  run create_app
  assert_success

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-cron-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku logs:set $TEST_APP vector-cron-sink console://?encoding[codec]=text"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-cron-sink"
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-cron-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "console://?encoding[codec]=text"

  # as with vector-sink, only the exact raw flag returns an unredacted value
  run /bin/bash -c "dokku logs:report $TEST_APP --logs-computed-vector-cron-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "console://redacted"

  run /bin/bash -c "dokku logs:report $TEST_APP --format json | jq -r '.\"computed-vector-cron-sink\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "console://?encoding[codec]=text"

  # a general report redacts everything but the scheme, on both the raw and
  # the computed row
  run /bin/bash -c "dokku logs:report $TEST_APP 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "console://redacted" 2

  run /bin/bash -c "dokku logs:set $TEST_APP vector-cron-sink"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting vector-cron-sink"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-vector-cron-sink 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""
}

@test "(logs) vector.json cron routing" {
  run create_app
  assert_success

  # a plain sink alone must generate no transforms at all
  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink console://?encoding[codec]=json"
  assert_success

  run /bin/bash -c "jq -r '.transforms' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "null"

  run /bin/bash -c "jq -r '.sinks[\"docker-sink:$TEST_APP\"].inputs[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker-source:$TEST_APP"

  # adding a cron sink splits the source and rewires the plain sink
  run /bin/bash -c "dokku logs:set $TEST_APP vector-cron-sink console://?encoding[codec]=text"
  assert_success

  run /bin/bash -c "jq -r '.transforms[\"docker-router:$TEST_APP\"].type' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "route"

  run /bin/bash -c "jq -r '.transforms[\"docker-router:$TEST_APP\"].reroute_unmatched' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "jq -r '.transforms[\"docker-router:$TEST_APP\"].route.cron.source' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "com.dokku.container-type"

  run /bin/bash -c "jq -r '.transforms[\"docker-cron-remap:$TEST_APP\"].source' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku_cron_id"

  run /bin/bash -c "jq -r '.sinks[\"docker-sink:$TEST_APP\"].inputs[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker-router:$TEST_APP._unmatched"

  run /bin/bash -c "jq -r '.sinks[\"docker-cron-sink:$TEST_APP\"].inputs[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker-cron-remap:$TEST_APP"

  # dropping the plain sink leaves nothing to consume the unmatched branch
  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink"
  assert_success

  run /bin/bash -c "jq -r '.transforms[\"docker-router:$TEST_APP\"].reroute_unmatched' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "jq -r '.sinks[\"docker-sink:$TEST_APP\"]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "null"

  run /bin/bash -c "dokku logs:set $TEST_APP vector-cron-sink"
  assert_success

  # the global scope gets its own router and cron sink
  run /bin/bash -c "dokku logs:set --global vector-cron-sink console://?encoding[codec]=text"
  assert_success

  run /bin/bash -c "jq -r '.transforms[\"docker-global-router\"].type' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "route"

  run /bin/bash -c "jq -r '.sinks[\"docker-global-cron-sink\"].inputs[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker-global-cron-remap"

  run /bin/bash -c "dokku logs:set --global vector-cron-sink"
  assert_success
}

@test "(logs:report) global-vector-image and global-vector-networks raw" {
  run create_app
  assert_success

  run /bin/bash -c "dokku logs:set --global vector-image"
  assert_success

  run /bin/bash -c "dokku logs:report --global --format json | jq -r '.\"logs-global-vector-image\"'"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku logs:report --global --format json | jq -r '.\"logs-computed-vector-image\"'"
  assert_success
  assert_output_exists

  run /bin/bash -c "dokku logs:set --global vector-image timberio/vector:custom"
  assert_success

  run /bin/bash -c "dokku logs:report --global --format json | jq -r '.\"logs-global-vector-image\"'"
  assert_success
  assert_output "timberio/vector:custom"

  run /bin/bash -c "dokku logs:set --global vector-image"
  assert_success

  run /bin/bash -c "dokku logs:report --global --format json | jq -r '.\"logs-global-vector-networks\"'"
  assert_success
  assert_output ""
}

@test "(logs:report) does not expose deprecated logs-vector-global-* keys" {
  run /bin/bash -c "dokku logs:report --global --format json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "logs-vector-global-image"
  assert_output_not_contains "logs-vector-global-networks"
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

  run /bin/bash -c "dokku logs:report --global --logs-global-vector-networks 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test-vector-net-a"

  run /bin/bash -c "dokku logs:set --global vector-networks test-vector-net-a,test-vector-net-b 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting vector-networks"

  run /bin/bash -c "dokku logs:report --global --logs-global-vector-networks 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test-vector-net-a,test-vector-net-b"

  run /bin/bash -c "dokku logs:set --global vector-networks 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting vector-networks"

  run /bin/bash -c "dokku logs:report --global --logs-global-vector-networks 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists
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

  run /bin/bash -c "jq -r '.transforms' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "null"

  run /bin/bash -c "dokku logs:set --global app-label-alias global_alt_name" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-computed-app-label-alias" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "global_alt_name"

  run /bin/bash -c "cat /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # the source filter never moves off the label dokku actually applies, or the
  # source would match no container at all and silently collect nothing
  run /bin/bash -c "jq -r '.sources[\"docker-source:$TEST_APP\"].include_labels[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "com.dokku.app-name=$TEST_APP"

  run /bin/bash -c "jq -r '.transforms[\"docker-relabel:$TEST_APP\"].source' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output '.label."global_alt_name" = del(.label."com.dokku.app-name")'

  run /bin/bash -c "jq -r '.sinks[\"docker-sink:$TEST_APP\"].inputs[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker-relabel:$TEST_APP"

  run /bin/bash -c "dokku logs:set $TEST_APP app-label-alias alt_name" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Writing updated vector config to /var/lib/dokku/data/logs/vector.json"

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-computed-app-label-alias" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "alt_name"

  run /bin/bash -c "jq -r '.sources[\"docker-source:$TEST_APP\"].include_labels[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "com.dokku.app-name=$TEST_APP"

  run /bin/bash -c "jq -r '.transforms[\"docker-relabel:$TEST_APP\"].source' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output '.label."alt_name" = del(.label."com.dokku.app-name")'

  run /bin/bash -c "dokku logs:set $TEST_APP app-label-alias" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(logs:set) app-label-alias rejects an unusable label key" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP app-label-alias 'not a label' 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid app-label-alias value"

  run /bin/bash -c "dokku logs:set --global app-label-alias '_leading_underscore' 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid app-label-alias value"
}

# an app whose own alias differs from the global one is collected by the global
# source too, so the global pipeline carries a branch for it
@test "(logs) vector.json global relabel honors a per-app alias" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set --global vector-sink console://?encoding[codec]=json"
  assert_success

  run /bin/bash -c "dokku logs:set --global app-label-alias global_alt_name"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP app-label-alias app_alt_name"
  assert_success

  run /bin/bash -c "jq -r '.transforms[\"docker-global-relabel\"].source' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "if app == \"$TEST_APP\""
  assert_output_contains '.label."app_alt_name" = del(.label."com.dokku.app-name")'
  assert_output_contains '.label."global_alt_name" = del(.label."com.dokku.app-name")'

  run /bin/bash -c "jq -r '.sinks[\"docker-global-sink\"].inputs[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker-global-relabel"

  # a per-app alias matching the global one needs no branch of its own
  run /bin/bash -c "dokku logs:set $TEST_APP app-label-alias global_alt_name"
  assert_success

  run /bin/bash -c "jq -r '.transforms[\"docker-global-relabel\"].source' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output '.label."global_alt_name" = del(.label."com.dokku.app-name")'

  run /bin/bash -c "dokku logs:set $TEST_APP app-label-alias"
  assert_success

  run /bin/bash -c "dokku logs:set --global app-label-alias"
  assert_success

  run /bin/bash -c "dokku logs:set --global vector-sink"
  assert_success
}

@test "(logs) vector.json drops a destroyed app" {
  local DESTROYED_APP="${TEST_APP}-destroyed"

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:create $DESTROYED_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink console://?encoding[codec]=json"
  assert_success

  run /bin/bash -c "dokku logs:set $DESTROYED_APP vector-sink console://?encoding[codec]=json"
  assert_success

  run /bin/bash -c "jq -e '.sources | has(\"docker-source:$DESTROYED_APP\")' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy $DESTROYED_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # an orphaned sink keeps pointing at an endpoint decommissioned with the app
  run /bin/bash -c "jq -e '(.sources | has(\"docker-source:$DESTROYED_APP\")) or (.sinks | has(\"docker-sink:$DESTROYED_APP\"))' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "jq -e '.sources | has(\"docker-source:$TEST_APP\")' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(logs) vector.json follows a renamed app" {
  local RENAMED_APP="${TEST_APP}-renamed"

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink console://?encoding[codec]=json"
  assert_success

  run /bin/bash -c "dokku apps:rename --skip-deploy $TEST_APP $RENAMED_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # the sink property follows the app, so a stale config leaves the app with a
  # sink and no source feeding it
  run /bin/bash -c "dokku logs:report $RENAMED_APP --logs-computed-vector-sink"
  echo "output: $output"
  echo "status: $status"
  assert_output "console://redacted"

  run /bin/bash -c "jq -r '.sources[\"docker-source:$RENAMED_APP\"].include_labels[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_output "com.dokku.app-name=$RENAMED_APP"

  run /bin/bash -c "jq -e '(.sources | has(\"docker-source:$TEST_APP\")) or (.sinks | has(\"docker-sink:$TEST_APP\"))' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  TEST_APP="$RENAMED_APP"
}

@test "(logs) vector.json adds a source for a cloned app" {
  local CLONE_APP="${TEST_APP}-clone"

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink console://?encoding[codec]=json"
  assert_success

  run /bin/bash -c "dokku apps:clone --skip-deploy $TEST_APP $CLONE_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "jq -r '.sources[\"docker-source:$CLONE_APP\"].include_labels[0]' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_output "com.dokku.app-name=$CLONE_APP"

  run /bin/bash -c "jq -e '.sinks | has(\"docker-sink:$CLONE_APP\")' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "jq -e '.sources | has(\"docker-source:$TEST_APP\")' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy $CLONE_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(logs) vector.json global relabel follows a renamed app" {
  local RENAMED_APP="${TEST_APP}-renamed"

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set --global vector-sink console://?encoding[codec]=json"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP app-label-alias app_alt_name"
  assert_success

  run /bin/bash -c "dokku apps:rename --skip-deploy $TEST_APP $RENAMED_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # the relabel branches are generated VRL with app names baked in, so a rename
  # leaves a branch naming an app that no longer exists
  run /bin/bash -c "jq -r '.transforms[\"docker-global-relabel\"].source' /var/lib/dokku/data/logs/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "if app == \"$RENAMED_APP\""
  assert_output_not_contains "if app == \"$TEST_APP\""

  TEST_APP="$RENAMED_APP"
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
  assert_output "--log-opt=max-size=20m --restart=on-failure:10"

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
  assert_output "--restart=on-failure:10"

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
  assert_output "--log-opt=max-size=20m --restart=on-failure:10"
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
  assert_output "--log-opt=max-size=20m --restart=on-failure:10"

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy --log-driver=local" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo "" | dokku plugin:trigger docker-args-process-deploy $TEST_APP 2>&1 | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--log-opt=max-size=20m --restart=on-failure:10"

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy --log-driver=json-file" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo "" | dokku plugin:trigger docker-args-process-deploy $TEST_APP 2>&1 | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--log-opt=max-size=20m --restart=on-failure:10"

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy --log-driver=journald" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo "" | dokku plugin:trigger docker-args-process-deploy $TEST_APP 2>&1 | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--restart=on-failure:10"
}

@test "(logs:report) emits new stripped JSON keys alongside legacy" {
  run create_app
  assert_success

  run /bin/bash -c "dokku logs:report $TEST_APP --format json | jq -r 'has(\"vector-sink\") and has(\"logs-vector-sink\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku logs:report $TEST_APP --format json | jq -r 'has(\"global-vector-sink\") and has(\"logs-global-vector-sink\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku logs:report $TEST_APP --format json | jq -r 'has(\"computed-vector-sink\") and has(\"logs-computed-vector-sink\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku logs:report $TEST_APP --format json | jq -r 'has(\"max-size\") and has(\"logs-max-size\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku logs:report --global --format json | jq -r 'has(\"global-vector-image\") and has(\"logs-global-vector-image\")'"
  assert_success
  assert_output "true"
}
