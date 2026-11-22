#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  dokku datastore:unlink redis backup-redis $TEST_APP 2>/dev/null || true
  dokku datastore:destroy redis backup-redis --force 2>/dev/null || true
  destroy_app
  rm -f /tmp/dokku-backup-*.tar.gz 2>/dev/null || true
  global_teardown
}

@test "(backup) core datastore service round-trip keeps data and credentials" {
  run /bin/bash -c "dokku datastore:create redis backup-redis"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo 'SET backupkey backupvalue' | dokku datastore:connect redis backup-redis"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku datastore:link redis backup-redis $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku datastore:info redis backup-redis --dsn"
  echo "output: $output"
  assert_success
  local dsn="$output"

  run /bin/bash -c "dokku backup:export --service redis:backup-redis --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  local backup_file="$output"

  run /bin/bash -c "tar tzf '$backup_file'"
  echo "output: $output"
  assert_output_contains "services/redis/backup-redis/config/redis.yml"
  assert_output_contains "services/redis/backup-redis/data/service.json"
  assert_output_contains "services/redis/backup-redis/data/dumps/service.dump"

  run /bin/bash -c "dokku datastore:unlink redis backup-redis $TEST_APP"
  echo "output: $output"
  assert_success
  run /bin/bash -c "dokku datastore:destroy redis backup-redis --force"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku backup:import --service redis:backup-redis '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo 'GET backupkey' | dokku datastore:connect redis backup-redis"
  echo "output: $output"
  assert_output_contains "backupvalue"

  run /bin/bash -c "dokku datastore:info redis backup-redis --dsn"
  echo "output: $output"
  assert_output "$dsn"
}

@test "(backup:export) discovers core datastore services in a full export" {
  run /bin/bash -c "dokku datastore:create redis backup-redis"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku backup:export --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  local backup_file="$output"

  run /bin/bash -c "tar tzf '$backup_file'"
  echo "output: $output"
  assert_output_contains "services/redis/backup-redis/data/service.json"
}

@test "(backup:import) --force replaces an existing core datastore service" {
  run /bin/bash -c "dokku datastore:create redis backup-redis"
  echo "output: $output"
  assert_success

  run /bin/bash -c "echo 'SET backupkey original' | dokku datastore:connect redis backup-redis"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku backup:export --service redis:backup-redis --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  run /bin/bash -c "echo 'SET backupkey changed' | dokku datastore:connect redis backup-redis"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku backup:import --service redis:backup-redis '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku backup:import --force --service redis:backup-redis '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "echo 'GET backupkey' | dokku datastore:connect redis backup-redis"
  echo "output: $output"
  assert_output_contains "original"
}

@test "(backup:export) backs up a stopped core datastore service without its data" {
  run /bin/bash -c "dokku datastore:create redis backup-redis"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku datastore:stop redis backup-redis"
  echo "output: $output"
  assert_success

  run /bin/bash -c "dokku backup:export --service redis:backup-redis --backup-dir /tmp 2>&1 >/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "is not running"

  run /bin/bash -c "ls -t /tmp/dokku-backup-*.tar.gz | head -n1"
  local backup_file="$output"

  run /bin/bash -c "tar tzf '$backup_file'"
  echo "output: $output"
  assert_output_contains "services/redis/backup-redis/data/service.json"
  assert_output_contains "services/redis/backup-redis/data/dumps/" 0
}
