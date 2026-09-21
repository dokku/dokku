#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  rm -rf "$DOKKU_LIB_ROOT/data/storage/rdmtestapp*"
}

teardown() {
  destroy_app
  global_teardown
}

@test "(storage) [process-type] docker-args-process-deploy emits default and matching mounts" {
  run /bin/bash -c "dokku storage:create rdmtest-shared"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-web"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-shared --container-dir /shared"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-web --container-dir /web --process-type web"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo '' | dokku plugin:trigger docker-args-process-deploy $TEST_APP herokuish latest web | grep -o -- '-v [^ ]*' | cut -d: -f2 | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/shared /web"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-shared --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-web --force"
  assert_success
}

@test "(storage) [process-type] docker-args-process-deploy omits mounts scoped to another process" {
  run /bin/bash -c "dokku storage:create rdmtest-shared"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-web"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-shared --container-dir /shared"
  assert_success
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-web --container-dir /web --process-type web"
  assert_success

  run /bin/bash -c "echo '' | dokku plugin:trigger docker-args-process-deploy $TEST_APP herokuish latest worker | grep -o -- '-v [^ ]*' | cut -d: -f2 | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/shared"

  # An ephemeral container with no process type of its own - the app.json
  # deploy-task path invokes the trigger this way - still gets the default
  # scope, and only the default scope.
  run /bin/bash -c "echo '' | dokku plugin:trigger docker-args-process-deploy $TEST_APP herokuish latest | grep -o -- '-v [^ ]*' | cut -d: -f2 | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/shared"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-shared --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-web --force"
  assert_success
}

@test "(storage) [process-type] docker-args-run emits only default-scoped mounts" {
  run /bin/bash -c "dokku storage:create rdmtest-shared"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-web"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-shared --container-dir /shared"
  assert_success
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-web --container-dir /web --process-type web"
  assert_success

  run /bin/bash -c "echo '' | dokku plugin:trigger docker-args-run $TEST_APP latest | grep -o -- '-v [^ ]*' | cut -d: -f2 | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/shared"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-shared --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-web --force"
  assert_success
}

@test "(storage:mount) rejects a container path already mounted in an overlapping process type" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-data --container-dir /app/storage"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # The default scope also applies to web, so both mounts would target
  # /app/storage on the same container.
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-cache --container-dir /app/storage --process-type web"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "is already mounted by storage entry rdmtest-data for process type _default_"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  # Two named scopes never share a container, so they may each claim the path.
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-data --container-dir /scoped --process-type web"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-cache --container-dir /scoped --process-type worker"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-data --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-cache --force"
  assert_success
}

@test "(storage:migrate) drains process-scoped docker-options mounts" {
  run /bin/bash -c "mkdir -p /tmp/rdmtest-spool"
  assert_success

  run /bin/bash -c "dokku docker-options:add --process worker $TEST_APP deploy '-v /tmp/rdmtest-spool:/spool'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:migrate $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.container-path\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/spool"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.process-type\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "worker"

  run /bin/bash -c "dokku docker-options:report $TEST_APP --format json | jq -r '.. | strings' | grep -c 'rdmtest-spool' || true"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "rm -rf /tmp/rdmtest-spool"
  assert_success
}

@test "(storage) [process-type] a process-scoped mount reaches only its own process" {
  run /bin/bash -c "dokku storage:create rdmtest-shared"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-web"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-shared --container-dir /shared"
  assert_success
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-web --container-dir /web --process-type web"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP web_worker_callback
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Only web is scaled up by default, so the worker needs a container of its
  # own before there is anything to compare against.
  run /bin/bash -c "dokku ps:scale $TEST_APP worker=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect -f '{{range .HostConfig.Binds}}{{println .}}{{end}}' \$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1) | cut -d: -f2 | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/shared /web"

  run /bin/bash -c "docker inspect -f '{{range .HostConfig.Binds}}{{println .}}{{end}}' \$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.worker.1) | cut -d: -f2 | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/shared"

  run /bin/bash -c "dokku storage:unmount --all $TEST_APP"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-shared --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-web --force"
  assert_success
}

web_worker_callback() {
  local APP="$1"
  local APP_REPO_DIR="$2"
  cat >"$APP_REPO_DIR/Procfile" <<EOF
web: python3 -u web.py
worker: python3 -u worker.py
EOF
}
