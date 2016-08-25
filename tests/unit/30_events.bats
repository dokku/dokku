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

@test "(events) check conffiles" {
  run bash -c "test -f /etc/logrotate.d/dokku"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "test -f /etc/rsyslog.d/99-dokku.conf"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "stat -c '%U:%G:%a' /var/log/dokku/"
  echo "output: "$output
  echo "status: "$status
  assert_output "syslog:dokku:775"
  run bash -c "stat -c '%U:%G:%a' /var/log/dokku/events.log"
  echo "output: "$output
  echo "status: "$status
  assert_output "syslog:dokku:664"
}

@test "(events) log commands" {
  run dokku events:on
  deploy_app
  run dokku events
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku events:off
}
