#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  save_global_vhost
  create_app
}

teardown() {
  restore_global_vhost
  dokku storage:destroy --force backup-mode-data 2>/dev/null || true
  rm -rf /var/lib/dokku/data/storage/backup-mode-data 2>/dev/null || true
  destroy_app
  rm -f /tmp/dokku-backup-*.tar.gz 2>/dev/null || true
  global_teardown
}

@test "(backup) builder-pack stack is exported without the transient clear-cache flag" {
  dokku builder-pack:set $TEST_APP stack heroku/builder:24

  run /bin/bash -c "dokku builder-pack:report $TEST_APP --builder-pack-stack"
  echo "output: $output"
  assert_output "heroku/builder:24"

  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  run /bin/bash -c "tar xzf '$backup_file' -O apps/$TEST_APP/config/builder-pack.yml"
  echo "output: $output"
  assert_success
  assert_output_contains "property: stack"
  assert_output_contains "heroku/builder:24"
  assert_output_contains "clear-cache" 0
}

@test "(backup:import) migrates a legacy buildpacks stack to the matching builder" {
  mkdir -p "/var/lib/dokku/config/buildpacks/$TEST_APP"
  echo "gliderlabs/herokuish:latest-24" >"/var/lib/dokku/config/buildpacks/$TEST_APP/stack"
  chown -R dokku:dokku "/var/lib/dokku/config/buildpacks/$TEST_APP"

  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  run /bin/bash -c "tar xzf '$backup_file' -O apps/$TEST_APP/config/buildpacks.yml"
  echo "output: $output"
  assert_output_contains "property: stack"

  dokku --force apps:destroy $TEST_APP

  run /bin/bash -c "dokku backup:import --force '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku builder-herokuish:report $TEST_APP --builder-herokuish-stack"
  echo "output: $output"
  assert_output "gliderlabs/herokuish:latest-24"

  [[ ! -f "/var/lib/dokku/config/buildpacks/$TEST_APP/stack" ]]
}

@test "(backup:export) leaves registry credentials out of the archive" {
  dokku registry:set $TEST_APP server registry.example.com
  mkdir -p "/var/lib/dokku/config/registry/$TEST_APP"
  echo '{"auths":{}}' >"/var/lib/dokku/config/registry/$TEST_APP/config.json"
  chown -R dokku:dokku "/var/lib/dokku/config/registry/$TEST_APP"

  run /bin/bash -c "dokku backup:export --app $TEST_APP --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  run /bin/bash -c "tar xzf '$backup_file' -O apps/$TEST_APP/config/registry.yml"
  echo "output: $output"
  assert_success
  assert_output_contains "registry.example.com"
  assert_output_contains "config.json" 0
}

@test "(backup) --include-storage restores the entry directory mode" {
  dokku storage:create backup-mode-data --mode 0770
  dokku storage:mount $TEST_APP backup-mode-data --container-dir /app/storage
  echo "vol-payload" >"/var/lib/dokku/data/storage/backup-mode-data/sentinel.txt"

  run /bin/bash -c "dokku backup:export --app $TEST_APP --include-storage --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  dokku --force apps:destroy $TEST_APP
  dokku storage:destroy --force backup-mode-data 2>/dev/null || true
  rm -rf "/var/lib/dokku/data/storage/backup-mode-data"

  run /bin/bash -c "dokku backup:import --force '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  [[ "$(cat "/var/lib/dokku/data/storage/backup-mode-data/sentinel.txt")" == "vol-payload" ]]

  run /bin/bash -c "stat -c '%a' /var/lib/dokku/data/storage/backup-mode-data"
  echo "output: $output"
  assert_output "770"
}

@test "(backup) global domains round-trip" {
  dokku domains:set-global backup-global.example.com

  run /bin/bash -c "dokku backup:export --backup-dir /tmp 2>/dev/null"
  echo "output: $output"
  assert_success
  local backup_file="$output"

  run /bin/bash -c "tar xzf '$backup_file' -O global/config/domains.yml"
  echo "output: $output"
  assert_success
  assert_output_contains "global: true"
  assert_output_contains "backup-global.example.com"

  dokku domains:clear-global

  run /bin/bash -c "dokku backup:import --force '$backup_file'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report --global --domains-global-vhosts"
  echo "output: $output"
  assert_output "backup-global.example.com"
}

save_global_vhost() {
  declare desc="saves the global VHOST file so a test can change global domains"
  cp -a "$DOKKU_ROOT/VHOST" "$BATS_TMPDIR/backup-3-global-VHOST" 2>/dev/null || true
}

restore_global_vhost() {
  declare desc="restores the global VHOST file saved by save_global_vhost"
  if [[ -f "$BATS_TMPDIR/backup-3-global-VHOST" ]]; then
    cp -a "$BATS_TMPDIR/backup-3-global-VHOST" "$DOKKU_ROOT/VHOST"
    rm -f "$BATS_TMPDIR/backup-3-global-VHOST"
  else
    rm -f "$DOKKU_ROOT/VHOST"
  fi
}
