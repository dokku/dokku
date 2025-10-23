#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  rm -f /tmp/fake-docker-bin
}

teardown() {
  rm -f /tmp/fake-docker-bin
  destroy_app || true
  global_teardown
}

@test "(report) report" {
  deploy_app

  run /bin/bash -c "dokku report"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "herokuish version:"
  assert_success

  run /bin/bash -c "dokku report $TEST_APP 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "herokuish version:"
  assert_output_contains "Deployed:                      false" "0"
  assert_success

  run /bin/bash -c "dokku report fake-app-name"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku report fake-app-name 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "App fake-app-name does not exist"
  assert_failure

  run /bin/bash -c "dokku apps:create ${TEST_APP}-2"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku report ${TEST_APP}-2 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Deployed:                      false"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy ${TEST_APP}-2"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(report) custom docker bin" {
  deploy_app

  export DOCKER_BIN="docker"
  run /bin/bash -c "dokku report"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "herokuish version:"
  assert_success

  export DOCKER_BIN="/usr/bin/docker"
  run /bin/bash -c "dokku report"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "herokuish version:"
  assert_success

  touch /tmp/fake-docker-bin
  echo '#!/usr/bin/env bash' >/tmp/fake-docker-bin
  echo '/usr/bin/docker "$@"' >>/tmp/fake-docker-bin
  chmod +x /tmp/fake-docker-bin

  export DOCKER_BIN="/tmp/fake-docker-bin"
  run /bin/bash -c "dokku report"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "herokuish version:"
  assert_success

  unset DOCKER_BIN
}

@test "(report) exit 0 when no apps exist" {
  run /bin/bash -c "dokku report"
  echo "output: $output"
  echo "status: $status"
  assert_success

  plugins=(
    app-json
    apps
    builder
    builder-dockerfile
    builder-herokuish
    builder-lambda
    builder-nixpacks
    builder-railpack
    builder-pack
    buildpacks
    caddy
    certs
    checks
    cron
    docker-options
    domains
    git
    haproxy
    logs
    network
    nginx
    openresty
    ports
    proxy
    ps
    registry
    resource
    scheduler
    scheduler-docker-local
    scheduler-k3s
    storage
    traefik
  )

  for plugin in "${plugins[@]}"; do
    run /bin/bash -c "dokku $plugin:report 2>&1"
    echo "output: $output"
    echo "status: $status"
    assert_success
    assert_output_contains "You haven't deployed any applications yet"
  done
}
