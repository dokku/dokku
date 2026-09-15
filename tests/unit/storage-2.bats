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

@test "(storage:create) --chown rejects an out-of-bounds numeric uid" {
  run /bin/bash -c "dokku storage:create --chown 65536 rdmtest-chown-oob"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Unsupported chown permissions"

  run /bin/bash -c "dokku storage:create --chown -1 rdmtest-chown-oob"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Unsupported chown permissions"

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-chown-oob$' || true"
  assert_output ""
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

@test "(storage:create) --mode sets directory permissions" {
  run /bin/bash -c "dokku storage:create --mode 0777 rdmtest-mode"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "stat -c '%a' $DOKKU_LIB_ROOT/data/storage/rdmtest-mode"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "777"

  # the mode is stored on the entry in its canonical 4-digit form
  run /bin/bash -c "dokku storage:info rdmtest-mode --format json | jq -r '.mode'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0777"

  run /bin/bash -c "dokku storage:destroy rdmtest-mode --destroy-host-dir --force"
  assert_success
}

@test "(storage:create) --mode re-applies on an existing directory" {
  # the default mode assertion below only holds for a freshly created directory
  rm -rf "$DOKKU_LIB_ROOT/data/storage/rdmtest-mode-converge"

  run /bin/bash -c "dokku storage:create rdmtest-mode-converge"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "stat -c '%a' $DOKKU_LIB_ROOT/data/storage/rdmtest-mode-converge"
  assert_success
  assert_output "755"

  # re-running create against the existing entry converges the directory
  run /bin/bash -c "dokku storage:create --mode 0700 rdmtest-mode-converge"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "stat -c '%a' $DOKKU_LIB_ROOT/data/storage/rdmtest-mode-converge"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "700"

  run /bin/bash -c "dokku storage:destroy rdmtest-mode-converge --destroy-host-dir --force"
  assert_success
}

@test "(storage:create) --mode accepts a 3 digit octal mode" {
  run /bin/bash -c "dokku storage:create --mode 750 rdmtest-mode-short"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "stat -c '%a' $DOKKU_LIB_ROOT/data/storage/rdmtest-mode-short"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "750"

  run /bin/bash -c "dokku storage:info rdmtest-mode-short --format json | jq -r '.mode'"
  assert_success
  assert_output "0750"

  run /bin/bash -c "dokku storage:destroy rdmtest-mode-short --destroy-host-dir --force"
  assert_success
}

@test "(storage:create) --mode rejects an invalid value" {
  run /bin/bash -c "dokku storage:create --mode 0888 rdmtest-mode-bad"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Unsupported directory mode"

  run /bin/bash -c "dokku storage:create --mode u+rwx rdmtest-mode-bad"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Unsupported directory mode"

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-mode-bad$' || true"
  assert_output ""
}

@test "(storage:create) --mode rejects a non-default host path" {
  custom_path="/tmp/rdmtest-mode-custom"
  rm -rf "$custom_path"

  run /bin/bash -c "dokku storage:create --mode 0777 rdmtest-mode-custom $custom_path"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "--mode is only supported when the storage entry uses the default host path"

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-mode-custom$' || true"
  assert_output ""

  rm -rf "$custom_path"
}

@test "(storage:create) --mode is rejected on a k3s entry" {
  run /bin/bash -c "dokku storage:create --scheduler k3s --size 1Gi --mode 0777 rdmtest-mode-k3s"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "does not accept --mode"

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-mode-k3s$' || true"
  assert_output ""
}

@test "(storage:set) mode converges the directory via the positional form" {
  run /bin/bash -c "dokku storage:create --mode 0755 rdmtest-mode-set"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:set rdmtest-mode-set mode 0770"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "stat -c '%a' $DOKKU_LIB_ROOT/data/storage/rdmtest-mode-set"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "770"

  run /bin/bash -c "dokku storage:info rdmtest-mode-set --format json | jq -r '.mode'"
  assert_success
  assert_output "0770"

  # the text renderer surfaces it too
  run /bin/bash -c "dokku storage:info rdmtest-mode-set"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Mode:"
  assert_output_contains "0770"

  run /bin/bash -c "dokku storage:destroy rdmtest-mode-set --destroy-host-dir --force"
  assert_success
}

@test "(storage:set) the deprecated flag form still works and warns" {
  run /bin/bash -c "dokku storage:create --mode 0755 rdmtest-set-flags"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:set rdmtest-set-flags --mode 0770 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Deprecated:"

  run /bin/bash -c "stat -c '%a' $DOKKU_LIB_ROOT/data/storage/rdmtest-set-flags"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "770"

  run /bin/bash -c "dokku storage:destroy rdmtest-set-flags --destroy-host-dir --force"
  assert_success
}

@test "(storage:set) sets a property via the positional form" {
  run /bin/bash -c "dokku storage:create rdmtest-set-prop"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:set rdmtest-set-prop chown herokuish"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:info rdmtest-set-prop --format json | jq -r '.chown'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "herokuish"

  run /bin/bash -c "dokku storage:destroy rdmtest-set-prop --destroy-host-dir --force"
  assert_success
}

@test "(storage:set) unsets a property when the value is omitted" {
  run /bin/bash -c "dokku storage:create --mode 0770 rdmtest-set-unset"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:info rdmtest-set-unset --format json | jq -r '.mode'"
  assert_success
  assert_output "0770"

  run /bin/bash -c "dokku storage:set rdmtest-set-unset mode"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:info rdmtest-set-unset --format json | jq -r '.mode // empty'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku storage:destroy rdmtest-set-unset --destroy-host-dir --force"
  assert_success
}

@test "(storage:set) rejects an unknown property" {
  run /bin/bash -c "dokku storage:create rdmtest-set-bad"
  assert_success

  run /bin/bash -c "dokku storage:set rdmtest-set-bad bogus value"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid property specified, valid properties include:"

  run /bin/bash -c "dokku storage:set rdmtest-set-bad"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "No property specified"

  run /bin/bash -c "dokku storage:destroy rdmtest-set-bad --destroy-host-dir --force"
  assert_success
}

@test "(storage:set) rejects mixing a property with flags" {
  run /bin/bash -c "dokku storage:create rdmtest-set-mixed"
  assert_success

  run /bin/bash -c "dokku storage:set rdmtest-set-mixed mode 0770 --chown herokuish"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "either a property and a value or flags, not both"

  run /bin/bash -c "dokku storage:destroy rdmtest-set-mixed --destroy-host-dir --force"
  assert_success
}

@test "(storage:set) refuses an in-place access-mode or storage-class change" {
  run /bin/bash -c "dokku storage:create rdmtest-set-inplace"
  assert_success

  run /bin/bash -c "dokku storage:set rdmtest-set-inplace access-mode ReadWriteMany"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "cannot change access-mode in place"

  run /bin/bash -c "dokku storage:set rdmtest-set-inplace storage-class-name longhorn"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "cannot change storage-class-name in place"

  run /bin/bash -c "dokku storage:destroy rdmtest-set-inplace --destroy-host-dir --force"
  assert_success
}

@test "(storage:destroy) leaves the host directory in place by default" {
  run /bin/bash -c "dokku storage:create rdmtest-keep-dir"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:destroy rdmtest-keep-dir --force"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -d $DOKKU_LIB_ROOT/data/storage/rdmtest-keep-dir"
  echo "output: $output"
  echo "status: $status"
  assert_success

  rm -rf "$DOKKU_LIB_ROOT/data/storage/rdmtest-keep-dir"
}

@test "(storage:destroy) --destroy-host-dir removes a non-empty host directory" {
  run /bin/bash -c "dokku storage:create rdmtest-drop-dir"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sudo touch $DOKKU_LIB_ROOT/data/storage/rdmtest-drop-dir/payload"
  assert_success

  run /bin/bash -c "dokku storage:destroy rdmtest-drop-dir --destroy-host-dir --force"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -d $DOKKU_LIB_ROOT/data/storage/rdmtest-drop-dir"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-drop-dir$' || true"
  assert_output ""
}

@test "(storage:destroy) --reclaim-policy Delete removes the host directory" {
  run /bin/bash -c "dokku storage:create --reclaim-policy Delete rdmtest-reclaim"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:destroy rdmtest-reclaim --force"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -d $DOKKU_LIB_ROOT/data/storage/rdmtest-reclaim"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(storage:create) --reclaim-policy Delete rejects a non-default host path" {
  custom_path="/tmp/rdmtest-reclaim-custom"
  rm -rf "$custom_path"

  run /bin/bash -c "dokku storage:create --reclaim-policy Delete rdmtest-reclaim-custom $custom_path"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "default host path"

  run /bin/bash -c "dokku storage:list-entries --format json | jq -r '.[].name' | grep '^rdmtest-reclaim-custom$' || true"
  assert_output ""

  rm -rf "$custom_path"
}

@test "(storage:destroy) --destroy-host-dir refuses a non-default host path" {
  custom_path="/tmp/rdmtest-drop-custom"
  rm -rf "$custom_path"
  mkdir -p "$custom_path"

  run /bin/bash -c "dokku storage:create rdmtest-drop-custom $custom_path"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:destroy rdmtest-drop-custom --destroy-host-dir --force"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "--destroy-host-dir is only supported when the storage entry uses the default host path"

  run /bin/bash -c "test -d $custom_path"
  assert_success

  run /bin/bash -c "dokku storage:destroy rdmtest-drop-custom --force"
  assert_success

  rm -rf "$custom_path"
}

@test "(storage:annotations:set) sets and clears a single key" {
  run /bin/bash -c "dokku storage:create rdmtest-annot"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:annotations:set rdmtest-annot first one"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:info rdmtest-annot --format json | jq -r '.annotations.first'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "one"

  run /bin/bash -c "dokku storage:annotations:set rdmtest-annot first"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:info rdmtest-annot --format json | jq -r '.annotations // empty'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku storage:destroy rdmtest-annot --destroy-host-dir --force"
  assert_success
}

@test "(storage:annotations:set) leaves other keys untouched" {
  run /bin/bash -c "dokku storage:create rdmtest-annot-multi"
  assert_success

  run /bin/bash -c "dokku storage:annotations:set rdmtest-annot-multi first one"
  assert_success
  run /bin/bash -c "dokku storage:annotations:set rdmtest-annot-multi second two"
  assert_success

  run /bin/bash -c "dokku storage:annotations:set rdmtest-annot-multi first"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:info rdmtest-annot-multi --format json | jq -r '.annotations.second'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "two"

  run /bin/bash -c "dokku storage:info rdmtest-annot-multi --format json | jq -r '.annotations.first // empty'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku storage:destroy rdmtest-annot-multi --destroy-host-dir --force"
  assert_success
}

@test "(storage:annotations:set) accepts a kubernetes-style key containing a slash" {
  run /bin/bash -c "dokku storage:create rdmtest-annot-slash"
  assert_success

  run /bin/bash -c "dokku storage:annotations:set rdmtest-annot-slash backup.velero.io/backup-volumes rdmtest-annot-slash"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:info rdmtest-annot-slash --format json | jq -r '.annotations.\"backup.velero.io/backup-volumes\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "rdmtest-annot-slash"

  run /bin/bash -c "dokku storage:destroy rdmtest-annot-slash --destroy-host-dir --force"
  assert_success
}

@test "(storage:annotations:report) reports one entry and every entry" {
  run /bin/bash -c "dokku storage:create rdmtest-annot-rpt"
  assert_success
  run /bin/bash -c "dokku storage:annotations:set rdmtest-annot-rpt team billing"
  assert_success

  run /bin/bash -c "dokku storage:annotations:report rdmtest-annot-rpt"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "rdmtest-annot-rpt annotations information"
  assert_output_contains "Annotation team:"
  assert_output_contains "billing"

  # without a name, every registered entry is covered
  run /bin/bash -c "dokku storage:annotations:report"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "rdmtest-annot-rpt annotations information"

  run /bin/bash -c "dokku storage:destroy rdmtest-annot-rpt --destroy-host-dir --force"
  assert_success
}

@test "(storage:annotations:report) emits json and answers a single info flag" {
  run /bin/bash -c "dokku storage:create rdmtest-annot-json"
  assert_success
  run /bin/bash -c "dokku storage:annotations:set rdmtest-annot-json team billing"
  assert_success

  run /bin/bash -c "dokku storage:annotations:report rdmtest-annot-json --format json | jq -r '.team'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "billing"

  run /bin/bash -c "dokku storage:annotations:report rdmtest-annot-json --storage-annotations.team"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "billing"

  run /bin/bash -c "dokku storage:annotations:report rdmtest-annot-json --storage-annotations.absent"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid flag passed, valid flags:"

  run /bin/bash -c "dokku storage:destroy rdmtest-annot-json --destroy-host-dir --force"
  assert_success
}

@test "(storage:labels:set) sets and clears a single key" {
  run /bin/bash -c "dokku storage:create rdmtest-label"
  assert_success

  run /bin/bash -c "dokku storage:labels:set rdmtest-label app.kubernetes.io/part-of billing"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:info rdmtest-label --format json | jq -r '.labels.\"app.kubernetes.io/part-of\"'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "billing"

  # labels and annotations are stored separately
  run /bin/bash -c "dokku storage:info rdmtest-label --format json | jq -r '.annotations // empty'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku storage:labels:set rdmtest-label app.kubernetes.io/part-of"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku storage:info rdmtest-label --format json | jq -r '.labels // empty'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku storage:destroy rdmtest-label --destroy-host-dir --force"
  assert_success
}

@test "(storage:labels:report) reports one entry" {
  run /bin/bash -c "dokku storage:create rdmtest-label-rpt"
  assert_success
  run /bin/bash -c "dokku storage:labels:set rdmtest-label-rpt tier cache"
  assert_success

  run /bin/bash -c "dokku storage:labels:report rdmtest-label-rpt"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "rdmtest-label-rpt labels information"
  assert_output_contains "Label tier:"
  assert_output_contains "cache"

  run /bin/bash -c "dokku storage:labels:report rdmtest-label-rpt --format json | jq -r '.tier'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "cache"

  run /bin/bash -c "dokku storage:destroy rdmtest-label-rpt --destroy-host-dir --force"
  assert_success
}

@test "(storage:annotations:set) fails on a missing entry or key" {
  run /bin/bash -c "dokku storage:annotations:set rdmtest-annot-absent key value"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "does not exist"

  run /bin/bash -c "dokku storage:create rdmtest-annot-nokey"
  assert_success

  run /bin/bash -c "dokku storage:annotations:set rdmtest-annot-nokey"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "No annotation key specified"

  run /bin/bash -c "dokku storage:destroy rdmtest-annot-nokey --destroy-host-dir --force"
  assert_success
}

@test "(storage:destroy) --destroy-host-dir warns before the confirmation prompt" {
  run /bin/bash -c "dokku storage:create rdmtest-drop-confirm"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sudo touch $DOKKU_LIB_ROOT/data/storage/rdmtest-drop-confirm/payload"
  assert_success

  run /bin/bash -c "echo rdmtest-drop-confirm | dokku storage:destroy rdmtest-drop-confirm --destroy-host-dir 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "which will be removed along with its contents"
  assert_output_contains "WARNING: Potentially Destructive Action"

  run /bin/bash -c "test -d $DOKKU_LIB_ROOT/data/storage/rdmtest-drop-confirm"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(storage:destroy) --destroy-host-dir is rejected on a k3s entry" {
  # a k3s entry cannot be created without a cluster, since storage:create
  # rolls the entry back when the scheduler trigger fails
  entry_path="$DOKKU_LIB_ROOT/data/storage-registry/entries/rdmtest-drop-k3s.json"
  run /bin/bash -c "echo '{\"name\":\"rdmtest-drop-k3s\",\"scheduler\":\"k3s\",\"size\":\"1Gi\",\"schema_version\":1}' | sudo tee $entry_path >/dev/null"
  assert_success
  run /bin/bash -c "sudo chown dokku:dokku $entry_path"
  assert_success

  run /bin/bash -c "dokku storage:destroy rdmtest-drop-k3s --destroy-host-dir --force"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "--destroy-host-dir only applies to docker-local storage entries"

  run /bin/bash -c "sudo rm -f $entry_path"
  assert_success
}

@test "(storage:set) the deprecated --annotation flag warns toward annotations:set" {
  run /bin/bash -c "dokku storage:create rdmtest-set-annot"
  assert_success

  run /bin/bash -c "dokku storage:set rdmtest-set-annot --annotation team=billing 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Deprecated:"
  assert_output_contains "storage:annotations:set"

  run /bin/bash -c "dokku storage:info rdmtest-set-annot --format json | jq -r '.annotations.team'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "billing"

  run /bin/bash -c "dokku storage:destroy rdmtest-set-annot --destroy-host-dir --force"
  assert_success
}

@test "(storage:mount) rejects k3s storage entry on docker-local app" {
  entry_path="$DOKKU_LIB_ROOT/data/storage-registry/entries/rdmtest-k3s-entry.json"
  run /bin/bash -c "echo '{\"name\":\"rdmtest-k3s-entry\",\"scheduler\":\"k3s\",\"size\":\"1Gi\",\"schema_version\":1}' | sudo tee $entry_path >/dev/null"
  assert_success
  run /bin/bash -c "sudo chown dokku:dokku $entry_path"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-k3s-entry --container-dir /data"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "storage entry \"rdmtest-k3s-entry\" is scheduler=k3s but cannot be mounted on a docker-local app; recreate it with --scheduler docker-local"

  run /bin/bash -c "dokku storage:list $TEST_APP --format json | jq -r '.[].entry_name' | grep '^rdmtest-k3s-entry$' || true"
  assert_output ""

  run /bin/bash -c "sudo rm -f $entry_path"
  assert_success
}

@test "(storage:mount) rejects docker-local storage entry on k3s app" {
  run /bin/bash -c "dokku storage:create rdmtest-local-entry"
  assert_success

  run /bin/bash -c "dokku scheduler:set $TEST_APP selected k3s"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP rdmtest-local-entry --container-dir /data"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "storage entry \"rdmtest-local-entry\" is scheduler=docker-local but cannot be mounted on a k3s app; recreate it with --scheduler k3s"

  run /bin/bash -c "dokku scheduler:set $TEST_APP selected docker-local"
  assert_success

  run /bin/bash -c "dokku storage:destroy rdmtest-local-entry --destroy-host-dir --force"
  assert_success
}

@test "(storage:mount) rejects colon-form mount on k3s app" {
  run /bin/bash -c "dokku scheduler:set $TEST_APP selected k3s"
  assert_success

  run /bin/bash -c "dokku storage:mount $TEST_APP /tmp/custom-mount:/data"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "colon-form mounts are only supported on docker-local apps"

  run /bin/bash -c "dokku scheduler:set $TEST_APP selected docker-local"
  assert_success
}
