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

@test "(docker-options:add) --process scopes option to one process type" {
  run /bin/bash -c "dokku docker-options:add --process web $TEST_APP deploy '-p 8080:5000'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:list $TEST_APP --process web --phase deploy"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "-p 8080:5000"

  run /bin/bash -c "dokku docker-options:list $TEST_APP --phase deploy"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-p 8080:5000" 0
}

@test "(docker-options:add) multiple --process flags add to each process" {
  run /bin/bash -c "dokku docker-options:add --process web --process worker $TEST_APP deploy '-v /shared:/shared'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:list $TEST_APP --process web --phase deploy"
  echo "output: $output"
  assert_output "-v /shared:/shared"

  run /bin/bash -c "dokku docker-options:list $TEST_APP --process worker --phase deploy"
  echo "output: $output"
  assert_output "-v /shared:/shared"
}

@test "(docker-options:add) without --process keeps default-scope behaviour" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy '-v /tmp/shared:/shared'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:list $TEST_APP --phase deploy"
  echo "output: $output"
  assert_output_contains "-v /tmp/shared:/shared"

  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy"
  echo "output: $output"
  assert_output_contains "-v /tmp/shared:/shared"
}

@test "(docker-options:add) rejects --process for non-deploy phases" {
  run /bin/bash -c "dokku docker-options:add --process web $TEST_APP build '--shm-size 256m'"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "deploy phase"

  run /bin/bash -c "dokku docker-options:add --process web $TEST_APP run '--shm-size 256m'"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku docker-options:add --process web $TEST_APP build,deploy '--shm-size 256m'"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(docker-options:add) rejects reserved _default_ process name" {
  run /bin/bash -c "dokku docker-options:add --process _default_ $TEST_APP deploy '-v /tmp:/tmp'"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "_default_"
}

@test "(docker-options:add) warns on process types missing from Procfile" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:add --process nonexistent $TEST_APP deploy '-v /tmp:/tmp'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "nonexistent"
}

@test "(docker-options:remove) --process removes from a single process" {
  run /bin/bash -c "dokku docker-options:add --process web $TEST_APP deploy '-p 8080:5000'"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku docker-options:add --process worker $TEST_APP deploy '-p 9000:9000'"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku docker-options:remove --process web $TEST_APP deploy '-p 8080:5000'"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku docker-options:list $TEST_APP --process web --phase deploy"
  echo "output: $output"
  assert_output ""

  run /bin/bash -c "dokku docker-options:list $TEST_APP --process worker --phase deploy"
  echo "output: $output"
  assert_output "-p 9000:9000"
}

@test "(docker-options:clear) --process clears only that process" {
  run /bin/bash -c "dokku docker-options:add --process web $TEST_APP deploy '-p 8080:5000'"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku docker-options:add --process worker $TEST_APP deploy '-p 9000:9000'"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku docker-options:clear --process web $TEST_APP deploy"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku docker-options:list $TEST_APP --process web --phase deploy"
  echo "output: $output"
  assert_output ""

  run /bin/bash -c "dokku docker-options:list $TEST_APP --process worker --phase deploy"
  echo "output: $output"
  assert_output "-p 9000:9000"
}

@test "(docker-options:list) --phase is required" {
  run /bin/bash -c "dokku docker-options:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "phase"
}

@test "(docker-options:report) exposes dynamic per-process keys" {
  run /bin/bash -c "dokku docker-options:add --process web $TEST_APP deploy '-p 8080:5000'"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy.web"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-p 8080:5000"

  run /bin/bash -c "dokku docker-options:report $TEST_APP"
  echo "output: $output"
  assert_success
  assert_output_contains "Docker options deploy web"
  assert_output_contains "-p 8080:5000"
}

@test "(docker-options:report) --format json carries process-scoped options" {
  run /bin/bash -c "dokku docker-options:add --process web $TEST_APP deploy '-p 8080:5000'"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy '-v /logs:/logs'"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku docker-options:report $TEST_APP --format json"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains '"docker-options-deploy.web"'
  assert_output_contains "-p 8080:5000"
  assert_output_contains '"docker-options-deploy"'
  assert_output_contains "-v /logs:/logs"
}

@test "(docker-options) clone copies default and per-process options" {
  local CLONE_APP="${TEST_APP}-clone"
  run /bin/bash -c "dokku docker-options:add --process web $TEST_APP deploy '-p 8080:5000'"
  assert_success
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy '-v /logs:/logs'"
  assert_success

  run /bin/bash -c "dokku apps:clone --skip-deploy $TEST_APP $CLONE_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:list $CLONE_APP --process web --phase deploy"
  echo "output: $output"
  assert_output "-p 8080:5000"

  run /bin/bash -c "dokku docker-options:list $CLONE_APP --phase deploy"
  echo "output: $output"
  assert_output_contains "-v /logs:/logs"

  dokku --force apps:destroy "$CLONE_APP" || true
}

@test "(docker-options) rename moves default and per-process options" {
  local RENAMED_APP="${TEST_APP}-renamed"
  run /bin/bash -c "dokku docker-options:add --process web $TEST_APP deploy '-p 8080:5000'"
  assert_success
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy '-v /logs:/logs'"
  assert_success

  run /bin/bash -c "dokku apps:rename $TEST_APP $RENAMED_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku docker-options:list $RENAMED_APP --process web --phase deploy"
  assert_output "-p 8080:5000"

  run /bin/bash -c "dokku docker-options:list $RENAMED_APP --phase deploy"
  assert_output_contains "-v /logs:/logs"

  [[ ! -d "/var/lib/dokku/config/docker-options/$TEST_APP" ]]

  TEST_APP="$RENAMED_APP"
}

@test "(docker-options) deploy with web-only port mapping does not break worker" {
  run /bin/bash -c "dokku docker-options:add --process web $TEST_APP deploy '-p 8080:5000'"
  echo "output: $output"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP web_worker_callback
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "App container failed to start" 0
  assert_output_contains "all ports are allocated" 0

  run /bin/bash -c "dokku ps:report $TEST_APP --ps-status-web-1"
  echo "output: $output"
  assert_output_contains "running"

  run /bin/bash -c "dokku ps:report $TEST_APP --ps-status-worker-1"
  echo "output: $output"
  assert_output_contains "running"
}

web_worker_callback() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  cat >"$APP_REPO_DIR/Procfile" <<EOF
web: python3 -u web.py
worker: python3 -u worker.py
EOF
}
