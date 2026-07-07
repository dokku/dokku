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

@test "(storage) storage:help" {
  run /bin/bash -c "dokku storage"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage mounted volumes"
  help_output="$output"

  run /bin/bash -c "dokku storage:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage mounted volumes"
  assert_output "$help_output"
}

@test "(storage) storage:ensure-directory" {
  run /bin/bash -c "test -d $DOKKU_LIB_ROOT/data/storage/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku storage:ensure-directory @"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP/"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 0
  assert_output_contains "Setting directory ownership to 2000:2000" 0
  assert_output_contains "Setting directory ownership to 32767:32767" 1

  run /bin/bash -c "test -d $DOKKU_LIB_ROOT/data/storage/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:ensure-directory --chown false $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 0
  assert_output_contains "Setting directory ownership to 2000:2000" 0
  assert_output_contains "Setting directory ownership to 32767:32767" 0

  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP --chown false"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 0
  assert_output_contains "Setting directory ownership to 2000:2000" 0
  assert_output_contains "Setting directory ownership to 32767:32767" 0

  run /bin/bash -c "dokku storage:ensure-directory --chown heroku $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 1
  assert_output_contains "Setting directory ownership to 2000:2000" 0
  assert_output_contains "Setting directory ownership to 32767:32767" 0

  run /bin/bash -c "dokku storage:ensure-directory --chown paketo $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 0
  assert_output_contains "Setting directory ownership to 2000:2000" 1
  assert_output_contains "Setting directory ownership to 32767:32767" 0

  run /bin/bash -c "dokku storage:ensure-directory --chown herokuish $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting directory ownership to 1000:1000" 0
  assert_output_contains "Setting directory ownership to 2000:2000" 0
  assert_output_contains "Setting directory ownership to 32767:32767" 1
}

@test "(storage) storage:mount, storage:list, storage:umount" {
  run /bin/bash -c "dokku storage:mount $TEST_APP /tmp/mount:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet storage:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/tmp/mount:/mount"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].host_path'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/tmp/mount"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].container_path'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/mount"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r 'map(has(\"volume_options\") or has(\"readonly\")) | any'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku storage:mount $TEST_APP /tmp/mount:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_output " !     Mount path already exists."
  assert_failure

  run /bin/bash -c "dokku storage:unmount $TEST_APP /tmp/mount:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet storage:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku storage:unmount $TEST_APP /tmp/mount:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_output " !     Mount path does not exist."
  assert_failure

  run /bin/bash -c "dokku storage:mount $TEST_APP mount_volume:/mount"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet storage:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "mount_volume:/mount"
}

@test "(storage:report) build/deploy/run mounts raw" {
  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"storage-build-mounts\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"storage-deploy-mounts\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"storage-run-mounts\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '[keys[] | select(startswith(\"attachment.\"))] | length'"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku storage:mount $TEST_APP /tmp/storage-mount:/mount"
  assert_success

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"storage-deploy-mounts\"'"
  assert_success
  assert_output "-v /tmp/storage-mount:/mount"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"storage-run-mounts\"'"
  assert_success
  assert_output "-v /tmp/storage-mount:/mount"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"storage-build-mounts\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '[keys[] | select(startswith(\"attachment.\"))] | length'"
  assert_success
  assert_output "9"

  run /bin/bash -c "dokku storage:unmount $TEST_APP /tmp/storage-mount:/mount"
  assert_success
}

@test "(storage:report) emits per-attachment dotted keys" {
  run /bin/bash -c "dokku storage:create rdmtest-rpt"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-rpt --container-dir /data --phase deploy --phase run --volume-options Z --volume-chown herokuish --volume-subpath uploads"
  assert_success

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.entry-name\"'"
  assert_success
  assert_output "rdmtest-rpt"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.host-path\"'"
  assert_success
  assert_output "$DOKKU_LIB_ROOT/data/storage/rdmtest-rpt"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.container-path\"'"
  assert_success
  assert_output "/data"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.phases\"'"
  assert_success
  assert_output "deploy,run"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.process-type\"'"
  assert_success
  assert_output "_default_"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.subpath\"'"
  assert_success
  assert_output "uploads"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.readonly\"'"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.volume-options\"'"
  assert_success
  assert_output "Z"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"attachment.1.volume-chown\"'"
  assert_success
  assert_output "herokuish"

  # stdout output gains one line per field; verify a representative subset
  run /bin/bash -c "dokku storage:report $TEST_APP"
  assert_success
  assert_output_contains "Storage attachment 1 entry name"
  assert_output_contains "rdmtest-rpt" -1
  assert_output_contains "Storage attachment 1 volume options"
  assert_output_contains "Storage attachment 1 volume chown"
  assert_output_contains "herokuish"

  # info-flag lookup returns just the value
  run /bin/bash -c "dokku storage:report $TEST_APP --storage-attachment.1.volume-options"
  assert_success
  assert_output "Z"

  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-rpt --container-dir /data"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-rpt --force"
  assert_success
}

@test "(storage:report) multiple attachments cluster by index" {
  run /bin/bash -c "dokku storage:create rdmtest-rpt-a"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-rpt-b"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-rpt-a --container-dir /data --volume-options Z"
  assert_success
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-rpt-b --container-dir /cache --volume-options noexec,nosuid"
  assert_success

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '[.\"attachment.1.entry-name\", .\"attachment.2.entry-name\"] | sort | join(\"|\")'"
  assert_success
  assert_output "rdmtest-rpt-a|rdmtest-rpt-b"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '[.\"attachment.1.volume-options\", .\"attachment.2.volume-options\"] | sort | join(\"|\")'"
  assert_success
  assert_output "Z|noexec,nosuid"

  # info-flag lookup against the second attachment
  run /bin/bash -c "dokku storage:report $TEST_APP --storage-attachment.2.entry-name"
  assert_success
  assert_output_contains "rdmtest-rpt"

  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-rpt-a --container-dir /data"
  assert_success
  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-rpt-b --container-dir /cache"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-rpt-a --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-rpt-b --force"
  assert_success
}

@test "(storage:report) degrades gracefully when an attachment entry is missing" {
  run /bin/bash -c "dokku storage:create rdmtest-rpt-missing"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-rpt-missing --container-dir /data --volume-options Z"
  assert_success

  # delete the entry on disk so LoadEntry fails for the existing attachment
  run /bin/bash -c "sudo rm -f $DOKKU_LIB_ROOT/data/storage-registry/entries/rdmtest-rpt-missing.json"
  assert_success

  # discard dokku's stderr so the LogWarn line doesn't pollute the jq output capture
  run /bin/bash -c "dokku storage:report $TEST_APP --format json 2>/dev/null | jq -r '[keys[] | select(startswith(\"attachment.\"))] | length'"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json 2>/dev/null | jq -r 'has(\"storage-build-mounts\") and has(\"storage-deploy-mounts\") and has(\"storage-run-mounts\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku storage:report $TEST_APP"
  assert_success
  assert_output_contains "Skipping attachment"
  assert_output_contains "rdmtest-rpt-missing" -1
  assert_output_contains "Storage build mounts"
  assert_output_contains "Storage deploy mounts"
  assert_output_contains "Storage run mounts"

  # clean up the leftover attachment record so subsequent tests don't trip
  run /bin/bash -c "sudo rm -f $DOKKU_LIB_ROOT/config/storage/$TEST_APP/attachments"
  assert_success
}

@test "(storage) storage:create / storage:list-entries / storage:destroy" {
  run /bin/bash -c "dokku storage:create rdmtest-entry"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-entry$'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-entry"

  run /bin/bash -c "dokku storage:info rdmtest-entry --format json | jq -r '.scheduler'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker-local"

  run /bin/bash -c "dokku storage:destroy rdmtest-entry --force"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(storage:create) --chown sets directory ownership" {
  run /bin/bash -c "dokku storage:create --chown herokuish rdmtest-chown"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "stat -c '%u:%g' $DOKKU_LIB_ROOT/data/storage/rdmtest-chown"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "32767:32767"

  run /bin/bash -c "dokku storage:destroy rdmtest-chown --force"
  assert_success
}

@test "(storage:create) --chown accepts a custom numeric uid" {
  run /bin/bash -c "dokku storage:create --chown 1500 rdmtest-chown-numeric"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "stat -c '%u:%g' $DOKKU_LIB_ROOT/data/storage/rdmtest-chown-numeric"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1500:1500"

  run /bin/bash -c "dokku storage:destroy rdmtest-chown-numeric --force"
  assert_success
}

@test "(storage:create) --chown rejects a non-default host path" {
  custom_path="/tmp/rdmtest-chown-custom"
  rm -rf "$custom_path"

  run /bin/bash -c "dokku storage:create --chown herokuish rdmtest-chown-custom $custom_path"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "--chown is only supported when the storage entry uses the default host path"

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-chown-custom$' || true"
  assert_output ""

  rm -rf "$custom_path"
}

@test "(storage) storage:create rejects invalid names" {
  # underscore: rejected
  run /bin/bash -c "dokku storage:create rdmtest_invalid"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  # uppercase: rejected
  run /bin/bash -c "dokku storage:create RdmTest"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  # 46 chars: too long
  long_name=$(printf 'a%.0s' {1..46})
  run /bin/bash -c "dokku storage:create $long_name"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  # legacy- prefix: reserved
  run /bin/bash -c "dokku storage:create legacy-foo"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(storage) storage:create + storage:mount with named entry attaches multiple entries to one app" {
  run /bin/bash -c "dokku storage:create rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:create rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-data --container-dir /data"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-cache --container-dir /cache"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].entry_name' | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-cache rdmtest-data"

  # cleanup
  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-data"
  assert_success
  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-cache"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku storage:destroy rdmtest-data --force"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-cache --force"
  assert_success
}

@test "(storage:mount) --volume-options sets option on named-entry attachment" {
  run /bin/bash -c "dokku storage:create rdmtest-opts"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-opts --container-dir /data --volume-options Z"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].volume_options'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "Z"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].container_path'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "/data"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"storage-deploy-mounts\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains ":Z"

  # combine with --volume-readonly: rendered as ro,<options>
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-opts --container-dir /ro --volume-options noexec,nosuid --volume-readonly"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r '.\"storage-deploy-mounts\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains ":ro,noexec,nosuid"

  # The JSON output of storage:list exposes readonly and volume_options
  # as separate keys so external drift-detection tooling can compare
  # them against the underlying attachment fields one for one.
  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[] | select(.container_path == \"/ro\") | .readonly'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[] | select(.container_path == \"/ro\") | .volume_options'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "noexec,nosuid"

  # The /data mount has no readonly flag, so the key is absent.
  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[] | select(.container_path == \"/data\") | has(\"readonly\")'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  # cleanup
  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-opts --container-dir /data"
  assert_success
  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-opts --container-dir /ro"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-opts --force"
  assert_success
}

@test "(storage:mount) re-running against an existing (entry, container-dir) updates in place" {
  run /bin/bash -c "dokku storage:create rdmtest-upsert"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-upsert --container-dir /data --volume-options Z"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Re-mount the same (entry, container-dir) with a different --volume-options.
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-upsert --container-dir /data --volume-options noexec,nosuid"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Single attachment, not two.
  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"

  # Latest --volume-options wins.
  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].volume_options'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "noexec,nosuid"

  # Re-mount with --volume-readonly added; readonly toggles on.
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-upsert --container-dir /data --volume-options noexec,nosuid --volume-readonly"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].readonly'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  # Re-mount without --volume-readonly: mount-time fields are rewritten,
  # not merged, so the readonly key disappears from the JSON output.
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-upsert --container-dir /data --volume-options noexec,nosuid"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[] | has(\"readonly\")'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  # A different --container-dir on the same entry is still a distinct attachment.
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-upsert --container-dir /other"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  # cleanup
  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-upsert --container-dir /data"
  assert_success
  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-upsert --container-dir /other"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-upsert --force"
  assert_success
}

@test "(storage) storage:destroy refuses to remove a still-mounted entry" {
  run /bin/bash -c "dokku storage:create rdmtest-busy"
  assert_success
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-busy --container-dir /data"
  assert_success

  run /bin/bash -c "dokku storage:destroy rdmtest-busy"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "still mounted"

  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-busy"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-busy --force"
  assert_success
}

@test "(storage:destroy) requires confirmation without --force" {
  run /bin/bash -c "dokku storage:create rdmtest-confirm"
  assert_success

  # No --force and no matching stdin: aborts, entry remains.
  run /bin/bash -c "dokku storage:destroy rdmtest-confirm < /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-confirm$'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-confirm"

  # Matching confirmation via stdin: succeeds and removes the entry.
  run /bin/bash -c "echo rdmtest-confirm | dokku storage:destroy rdmtest-confirm"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-confirm$' || true"
  echo "output: $output"
  echo "status: $status"
  assert_output ""
}

@test "(storage:destroy) --force skips confirmation" {
  run /bin/bash -c "dokku storage:create rdmtest-force"
  assert_success

  run /bin/bash -c "dokku storage:destroy rdmtest-force --force < /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-force$' || true"
  assert_output ""
}

@test "(storage:destroy) global --force skips confirmation" {
  run /bin/bash -c "dokku storage:create rdmtest-gforce"
  assert_success

  run /bin/bash -c "dokku --force storage:destroy rdmtest-gforce < /dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-gforce$' || true"
  assert_output ""
}

@test "(storage) storage:exec runs a non-interactive command and propagates exit code" {
  run /bin/bash -c "dokku storage:create rdmtest-exec"
  assert_success

  # docker pull image up-front so the test isn't gated on registry latency.
  run /bin/bash -c "docker pull alpine:3 >/dev/null"
  assert_success

  run /bin/bash -c "dokku storage:exec rdmtest-exec -- /bin/sh -c 'touch /data/marker && ls /data/marker'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "/data/marker"

  run /bin/bash -c "dokku storage:exec rdmtest-exec -- /bin/sh -c 'exit 42'"
  echo "output: $output"
  echo "status: $status"
  assert_equal "$status" 42

  run /bin/bash -c "dokku storage:destroy rdmtest-exec --force"
  assert_success
}

@test "(storage) storage:migrate converts legacy docker-options -v lines into attachments" {
  # Stage a legacy `-v` line directly via docker-options:add. This
  # bypasses the new storage:mount path so the line lives in
  # docker-options without a corresponding attachment, simulating an
  # app upgraded from a pre-PR Dokku version.
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy,run \"-v /tmp/legacy-mount:/legacy\""
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Before migration, storage:list shows nothing (attachment-only) and
  # docker-options has the -v line.
  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy"
  assert_success
  assert_output_contains "-v /tmp/legacy-mount:/legacy"

  # Run the migration for this app.
  run /bin/bash -c "dokku storage:migrate $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # storage:list now surfaces the synthesized colon form.
  run /bin/bash -c "dokku --quiet storage:list $TEST_APP"
  assert_success
  assert_output "/tmp/legacy-mount:/legacy"

  # The synthesized entry shows up under the legacy- prefix.
  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^legacy-' | head -1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # Regression for #8557: the legacy-*.json file must be readable by the
  # dokku user. The migration runs as root via the install trigger, so
  # without an explicit chown the file lands as root:root and ps:rebuild
  # fails with permission denied.
  legacy_entry_name=$(dokku storage:list-entries --format json | jq -r '.[] | select(.name | startswith("legacy-")) | .name' | head -1)
  run /bin/bash -c "stat -c '%U:%G' /var/lib/dokku/data/storage-registry/entries/${legacy_entry_name}.json"
  assert_success
  assert_output "dokku:dokku"

  # docker-options no longer holds the -v line on either phase.
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy"
  assert_success
  assert_output_not_contains "-v /tmp/legacy-mount:/legacy"

  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-run"
  assert_success
  assert_output_not_contains "-v /tmp/legacy-mount:/legacy"

  # Idempotency: re-running storage:migrate is a no-op (still one
  # attachment, no errors).
  run /bin/bash -c "dokku storage:migrate $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '. | length'"
  assert_success
  assert_output "1"

  # Property marker is set after a successful migration with mounts.
  run /bin/bash -c "sudo test -f /var/lib/dokku/config/storage/$TEST_APP/legacy-mounts-migrated"
  assert_success
  run /bin/bash -c "sudo cat /var/lib/dokku/config/storage/$TEST_APP/legacy-mounts-migrated"
  assert_success
  assert_output "true"

  # Cleanup: unmount + destroy the synthesized legacy entry. We need
  # the entry name from list-entries since legacy-<hash> is content-derived.
  legacy_name=$(dokku storage:list-entries --format json | jq -r '.[] | select(.name | startswith("legacy-")) | .name' | head -1)
  if [[ -n "$legacy_name" ]]; then
    run /bin/bash -c "dokku storage:unmount $TEST_APP $legacy_name --container-dir /legacy"
    assert_success
    run /bin/bash -c "dokku storage:destroy $legacy_name --force"
    assert_success
  fi
}

@test "(storage) storage:ensure-directory emits deprecation warning" {
  run /bin/bash -c "dokku storage:ensure-directory $TEST_APP 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Deprecated:"
}

@test "(storage:migrate) leaves legacy-mounts-migrated unset for apps with no legacy -v lines" {
  # Fresh app with no -v lines: storage:migrate must NOT write the
  # legacy-mounts-migrated property. Distinguishes "never had legacy
  # state" from "had legacy state, drained".
  sudo rm -f "/var/lib/dokku/config/storage/$TEST_APP/legacy-mounts-migrated"

  run /bin/bash -c "dokku storage:migrate $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sudo test -f /var/lib/dokku/config/storage/$TEST_APP/legacy-mounts-migrated"
  assert_failure
}

@test "(storage:migrate) converts legacy migration flag file into property" {
  # Stage an upgrade-from-old-code scenario: a leftover flag file under
  # data/storage-registry/migrations/<app>. storage:migrate (which runs
  # convertLegacyMigrationFlag before migrateApp) must drain it into
  # the property and remove the file.
  flag_dir="/var/lib/dokku/data/storage-registry/migrations"
  sudo mkdir -p "$flag_dir"
  sudo touch "$flag_dir/$TEST_APP"
  sudo chown dokku:dokku "$flag_dir/$TEST_APP"
  sudo rm -f "/var/lib/dokku/config/storage/$TEST_APP/legacy-mounts-migrated"

  run /bin/bash -c "dokku storage:migrate $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sudo test -f /var/lib/dokku/config/storage/$TEST_APP/legacy-mounts-migrated"
  assert_success
  run /bin/bash -c "sudo cat /var/lib/dokku/config/storage/$TEST_APP/legacy-mounts-migrated"
  assert_success
  assert_output "true"

  run /bin/bash -c "sudo test -e $flag_dir/$TEST_APP"
  assert_failure
}

@test "(storage:report) emits new stripped JSON keys alongside legacy" {
  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r 'has(\"build-mounts\") and has(\"storage-build-mounts\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r 'has(\"deploy-mounts\") and has(\"storage-deploy-mounts\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r 'has(\"run-mounts\") and has(\"storage-run-mounts\")'"
  assert_success
  assert_output "true"

  # per-attachment dotted keys are emitted in both stripped and legacy forms too
  run /bin/bash -c "dokku storage:create rdmtest-rpt-keys"
  assert_success
  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-rpt-keys --container-dir /data"
  assert_success

  run /bin/bash -c "dokku storage:report $TEST_APP --format json | jq -r 'has(\"attachment.1.entry-name\") and has(\"storage-attachment.1.entry-name\")'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku storage:unmount $TEST_APP rdmtest-rpt-keys --container-dir /data"
  assert_success
  run /bin/bash -c "dokku storage:destroy rdmtest-rpt-keys --force"
  assert_success
}
