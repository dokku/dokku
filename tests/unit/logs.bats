#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  destroy_app
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
  assert_output_contains "Invalid flag passed, valid flags: --logs-computed-max-size, --logs-global-max-size, --logs-global-vector-sink, --logs-max-size, --logs-vector-sink"

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
  assert_output_contains "Invalid property specified, valid properties include: max-size, vector-sink"

  run /bin/bash -c "dokku logs:set $TEST_APP invalid value" 2>&1
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid property specified, valid properties include: max-size, vector-sink"
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
  assert_output ""

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
  assert_output ""
}

@test "(logs) logs:set escaped uri" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink http://?uri=https%3A//loggerservice.com%3A1234/%3Ftoken%3Dabc1234%26type%3Dvector%26key%3Dvalue%2Bvalue2" 2>&1
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

  run /bin/bash -c "dokku logs:report $TEST_APP --logs-global-max-size 2>&1"
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
  assert_output "10m"
}

@test "(logs) logs:set max-size with alternate log-driver daemon " {
  if [[ "$REMOTE_CONTAINERS" == "true" ]]; then
    skip "skipping due non-existent docker service in remote dev container"
  fi

  if [[ ! -f /etc/docker/daemon.json ]]; then
    echo "{}" >/etc/docker/daemon.json
  fi

  driver="$(jq -r '."log-driver"' /etc/docker/daemon.json)"
  local TMP_FILE=$(mktemp "/tmp/dokku.me.XXXX")

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
  assert_output ""
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

  run /bin/bash -c "sudo docker inspect --format='{{.HostConfig.RestartPolicy.Name}}' vector"
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

  run /bin/bash -c "docker stop vector"
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
  assert_output_contains "Stopping and removing vector container"
}
