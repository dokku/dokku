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

@test "(run) run" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP echo $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker ps -a -f 'status=exited' --no-trunc=true | grep \"/exec echo $TEST_APP\""
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(run) run:detached" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  RANDOM_RUN_CID="$(dokku --label=com.dokku.test-label=value run:detached $TEST_APP sleep 300)"
  run /bin/bash -c "docker inspect -f '{{ .State.Status }}' $RANDOM_RUN_CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "running"

  run /bin/bash -c "docker inspect $RANDOM_RUN_CID --format '{{ index .Config.Labels \"com.dokku.test-label\" }}'"
  echo "output: $output"
  echo "status: $status"
  assert_output "value"

  run /bin/bash -c "docker stop $RANDOM_RUN_CID"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku cleanup"
  echo "output: $output"
  echo "status: $status"
  assert_success
  sleep 5 # wait for dokku cleanup to happen in the background
}

@test "(run) run (with tty)" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP ls /app/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(run) run (without tty)" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c ": |dokku run $TEST_APP ls /app/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(run) run command from Procfile" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP custom 'hi dokku' | tail -n 1"
  echo "output: $output"
  echo "status: $status"

  assert_success
  assert_output 'hi dokku'
}

@test "(run) list" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet run:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run:detached $TEST_APP sleep 300"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet run:list $TEST_APP | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "dokku run:list $TEST_APP --format json | jq '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "dokku run:list --format json $TEST_APP | jq '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "dokku run:detached $TEST_APP sleep 300"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet run:list $TEST_APP | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  run /bin/bash -c "dokku run:list $TEST_APP --format json | jq '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  run /bin/bash -c "dokku run:list --format json $TEST_APP | jq '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"
}

@test "(run) docker-options and -e flags are not eval-injected" {
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:add $TEST_APP run '--label=com.dokku.test=safe'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run --ttl-seconds=60 -e FOO=bar $TEST_APP env | grep -E '^FOO=bar'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run -e 'X=\$(id)' $TEST_APP env"
  echo "output: $output"
  echo "status: $status"
  assert_success
  [[ "$output" == *'X=$(id)'* ]] || flunk "expected literal X=\$(id) in container env"
  [[ "$output" != *"uid="* ]] || flunk "id command output leaked - env value was expanded"
}

@test "(run) --ttl-seconds rejects non-numeric input; docker-options stores it verbatim" {
  run /bin/bash -c "dokku run --ttl-seconds '\$(id)' $TEST_APP echo hi"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  [[ "$output" != *"uid="* ]] || flunk "id command output leaked - ttl value was expanded"

  run /bin/bash -c "dokku docker-options:add $TEST_APP run '--label=x=\$(id)'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-run"
  echo "output: $output"
  echo "status: $status"
  [[ "$output" == *'x=$(id)'* ]] || flunk "expected literal x=\$(id) stored in docker options"
  [[ "$output" != *"uid="* ]] || flunk "id command output leaked - docker option was expanded"
}
