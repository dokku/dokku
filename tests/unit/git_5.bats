#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  touch /home/dokku/.ssh/known_hosts
  chown dokku:dokku /home/dokku/.ssh/known_hosts
}

teardown() {
  rm -f /home/dokku/.ssh/id_rsa.pub || true
  destroy_app
  global_teardown
}

@test "(git) git:from-archive [missing]" {
  run /bin/bash -c "dokku git:from-archive $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(git) git:from-archive [invalid archive type]" {
  run /bin/bash -c "dokku git:from-archive --archive-type tarball $TEST_APP http://example.com/src.tar"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(git) git:from-archive [tar]" {
  run /bin/bash -c "dokku git:from-archive $TEST_APP https://github.com/dokku/smoke-test-app/releases/download/2.0.0/smoke-test-app.tar"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success http://${TEST_APP}.dokku.me
}

@test "(git) git:from-archive [tar.gz]" {
  run /bin/bash -c "dokku git:from-archive --archive-type tar.gz $TEST_APP https://github.com/dokku/smoke-test-app/archive/2.0.0.tar.gz"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success http://${TEST_APP}.dokku.me
}

@test "(git) git:from-archive [zip]" {
  run /bin/bash -c "dokku git:from-archive --archive-type zip $TEST_APP https://github.com/dokku/smoke-test-app/archive/2.0.0.zip"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success http://${TEST_APP}.dokku.me
}

@test "(git) git:from-archive [stdin-file]" {
  run /bin/bash -c "curl -sL https://github.com/dokku/smoke-test-app/releases/download/2.0.0/smoke-test-app.tar -o /tmp/smoke-test-app.tar"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /tmp/smoke-test-app.tar | dokku git:from-archive $TEST_APP  --"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success http://${TEST_APP}.dokku.me
}

@test "(git) git:from-archive [stdin-curl]" {
  run /bin/bash -c "curl -sL https://github.com/dokku/smoke-test-app/releases/download/2.0.0/smoke-test-app.tar | dokku git:from-archive $TEST_APP  --"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_http_success http://${TEST_APP}.dokku.me
}
