#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  rm -rf /home/dokku/$TEST_APP/tls /home/dokku/tls
  destroy_app
}

@test "build (long filenames)" {
  deploy_longfilename_app
  run bash -c "response=\"$(curl -s -S my-cool-guy-test-app.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "urls (non-ssl)" {
  run bash -c "dokku urls $TEST_APP | grep dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_output "http://dokku.me"
}

@test "urls (app ssl)" {
  mkdir -p /home/dokku/$TEST_APP/tls
  touch /home/dokku/$TEST_APP/tls/server.crt /home/dokku/$TEST_APP/tls/server.key
  run bash -c "dokku urls $TEST_APP | grep dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_output "https://dokku.me"
}

@test "urls (wildcard ssl)" {
  mkdir -p /home/dokku/tls
  touch /home/dokku/tls/server.crt /home/dokku/tls/server.key
  run bash -c "dokku urls $TEST_APP | grep dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_output "https://dokku.me"
}
