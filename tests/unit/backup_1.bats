#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  rm -f /tmp/dokku-backup-*.tar.gz 2>/dev/null || true
  global_teardown
}

@test "(backup) full app round-trip restores config, scale, code, and redeploys" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku config:set --no-restart $TEST_APP RTKEY=rtvalue
  dokku ps:scale $TEST_APP web=2
  dokku domains:add $TEST_APP backup-rt.example.com
  dokku docker-options:add $TEST_APP deploy "--memory=512m"

  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  local backup_file="$output"

  run /bin/bash -c "tar tzf '$backup_file'"
  echo "output: $output"
  assert_output_contains "apps/$TEST_APP/config/config.yml"
  assert_output_contains "apps/$TEST_APP/config/ps.yml"
  assert_output_contains "apps/$TEST_APP/data/repo.bundle"

  dokku --force apps:destroy $TEST_APP

  run /bin/bash -c "dokku backup:import '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:get $TEST_APP RTKEY"
  echo "output: $output"
  assert_output "rtvalue"

  run /bin/bash -c "dokku ps:scale $TEST_APP | grep -E '^web'"
  echo "output: $output"
  assert_output_contains "2"

  run /bin/bash -c "dokku ps:report $TEST_APP --deployed"
  echo "output: $output"
  assert_output_contains "true"

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-vhosts"
  echo "output: $output"
  assert_output_contains "backup-rt.example.com"

  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy"
  echo "output: $output"
  assert_output_contains "--memory=512m"
}
