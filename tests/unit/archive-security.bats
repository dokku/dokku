#!/usr/bin/env bats

load test_helper

ARCHIVE_TMP_DIR="${BATS_TMPDIR}/archive-security"

setup() {
  global_setup
  create_app
  mkdir -p "$ARCHIVE_TMP_DIR"
}

teardown() {
  rm -rf "$ARCHIVE_TMP_DIR"
  rm -f /tmp/dokku-archive-security-canary.txt
  dokku git:set --global archive-max-size "" >/dev/null 2>&1 || true
  dokku git:set --global archive-max-files "" >/dev/null 2>&1 || true
  destroy_app
  global_teardown
}

create_absolute_symlink_tar() {
  local OUTPUT="$1" FORMAT="${2:-tar}"
  python3 - "$OUTPUT" "$FORMAT" <<'PY'
import io, sys, tarfile
output, fmt = sys.argv[1], sys.argv[2]
mode = "w:gz" if fmt == "tar.gz" else "w"
with tarfile.open(output, mode) as t:
    link = tarfile.TarInfo("pwn")
    link.type = tarfile.SYMTYPE
    link.linkname = "/tmp"
    t.addfile(link)
    payload = b"canary content\n"
    fi = tarfile.TarInfo("pwn/dokku-archive-security-canary.txt")
    fi.size = len(payload)
    t.addfile(fi, io.BytesIO(payload))
    readme = b"# dummy\n"
    ri = tarfile.TarInfo("README.md")
    ri.size = len(readme)
    t.addfile(ri, io.BytesIO(readme))
PY
}

create_relative_traversal_symlink_tar() {
  local OUTPUT="$1" FORMAT="${2:-tar}"
  python3 - "$OUTPUT" "$FORMAT" <<'PY'
import io, sys, tarfile
output, fmt = sys.argv[1], sys.argv[2]
mode = "w:gz" if fmt == "tar.gz" else "w"
with tarfile.open(output, mode) as t:
    link = tarfile.TarInfo("pwn")
    link.type = tarfile.SYMTYPE
    link.linkname = "../../../../tmp"
    t.addfile(link)
    payload = b"canary content\n"
    fi = tarfile.TarInfo("pwn/dokku-archive-security-canary.txt")
    fi.size = len(payload)
    t.addfile(fi, io.BytesIO(payload))
PY
}

create_absolute_path_tar() {
  local OUTPUT="$1" FORMAT="${2:-tar}"
  python3 - "$OUTPUT" "$FORMAT" <<'PY'
import io, sys, tarfile
output, fmt = sys.argv[1], sys.argv[2]
mode = "w:gz" if fmt == "tar.gz" else "w"
with tarfile.open(output, mode) as t:
    payload = b"absolute path payload\n"
    fi = tarfile.TarInfo("/tmp/dokku-archive-security-canary.txt")
    fi.size = len(payload)
    t.addfile(fi, io.BytesIO(payload))
PY
}

create_traversal_path_tar() {
  local OUTPUT="$1" FORMAT="${2:-tar}"
  python3 - "$OUTPUT" "$FORMAT" <<'PY'
import io, sys, tarfile
output, fmt = sys.argv[1], sys.argv[2]
mode = "w:gz" if fmt == "tar.gz" else "w"
with tarfile.open(output, mode) as t:
    payload = b"traversal payload\n"
    fi = tarfile.TarInfo("../../../tmp/dokku-archive-security-canary.txt")
    fi.size = len(payload)
    t.addfile(fi, io.BytesIO(payload))
PY
}

create_safe_multi_file_tar() {
  local OUTPUT="$1"
  python3 - "$OUTPUT" <<'PY'
import io, sys, tarfile
output = sys.argv[1]
with tarfile.open(output, "w") as t:
    for name in ("README.md", "Procfile", "app.py", "requirements.txt"):
        payload = (name + " contents\n").encode()
        ti = tarfile.TarInfo(name)
        ti.size = len(payload)
        t.addfile(ti, io.BytesIO(payload))
PY
}

create_evil_certs_tar() {
  local OUTPUT="$1"
  python3 - "$OUTPUT" <<'PY'
import io, sys, tarfile
output = sys.argv[1]
with tarfile.open(output, "w") as t:
    link = tarfile.TarInfo("pwn")
    link.type = tarfile.SYMTYPE
    link.linkname = "/tmp"
    t.addfile(link)
    payload = b"canary content\n"
    fi = tarfile.TarInfo("pwn/dokku-archive-security-canary.txt")
    fi.size = len(payload)
    t.addfile(fi, io.BytesIO(payload))
    crt = b"-----BEGIN CERTIFICATE-----\nfake\n-----END CERTIFICATE-----\n"
    key = b"-----BEGIN PRIVATE KEY-----\nfake\n-----END PRIVATE KEY-----\n"
    tcrt = tarfile.TarInfo("server.crt")
    tcrt.size = len(crt)
    t.addfile(tcrt, io.BytesIO(crt))
    tkey = tarfile.TarInfo("server.key")
    tkey.size = len(key)
    t.addfile(tkey, io.BytesIO(key))
PY
}

@test "(archive-security) git:from-archive rejects tar with absolute symlink target" {
  create_absolute_symlink_tar "$ARCHIVE_TMP_DIR/evil.tar" "tar"
  run /bin/bash -c "cat $ARCHIVE_TMP_DIR/evil.tar | dokku git:from-archive $TEST_APP --"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "absolute targets"
  [[ ! -f /tmp/dokku-archive-security-canary.txt ]]
}

@test "(archive-security) git:from-archive rejects tar.gz with absolute symlink target" {
  create_absolute_symlink_tar "$ARCHIVE_TMP_DIR/evil.tar.gz" "tar.gz"
  run /bin/bash -c "cat $ARCHIVE_TMP_DIR/evil.tar.gz | dokku git:from-archive --archive-type tar.gz $TEST_APP --"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "absolute targets"
  [[ ! -f /tmp/dokku-archive-security-canary.txt ]]
}

@test "(archive-security) git:from-archive rejects tar with relative traversal symlink" {
  create_relative_traversal_symlink_tar "$ARCHIVE_TMP_DIR/evil.tar" "tar"
  run /bin/bash -c "cat $ARCHIVE_TMP_DIR/evil.tar | dokku git:from-archive $TEST_APP --"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "symlinks pointing outside extraction directory"
  [[ ! -f /tmp/dokku-archive-security-canary.txt ]]
}

@test "(archive-security) git:from-archive rejects tar with absolute paths" {
  create_absolute_path_tar "$ARCHIVE_TMP_DIR/evil.tar" "tar"
  run /bin/bash -c "cat $ARCHIVE_TMP_DIR/evil.tar | dokku git:from-archive $TEST_APP --"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "absolute paths"
  [[ ! -f /tmp/dokku-archive-security-canary.txt ]]
}

@test "(archive-security) git:from-archive rejects tar with parent traversal" {
  create_traversal_path_tar "$ARCHIVE_TMP_DIR/evil.tar" "tar"
  run /bin/bash -c "cat $ARCHIVE_TMP_DIR/evil.tar | dokku git:from-archive $TEST_APP --"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Archive contains entries with parent directory traversal"
  [[ ! -f /tmp/dokku-archive-security-canary.txt ]]
}

@test "(archive-security) certs:add rejects tar with absolute symlink target" {
  create_evil_certs_tar "$ARCHIVE_TMP_DIR/evil-certs.tar"
  run /bin/bash -c "cat $ARCHIVE_TMP_DIR/evil-certs.tar | dokku certs:add $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "absolute targets"
  [[ ! -f /tmp/dokku-archive-security-canary.txt ]]
}

@test "(archive-security) certs:add still works with legitimate tarball" {
  run /bin/bash -c "dokku certs:add $TEST_APP < $BATS_TEST_DIRNAME/server_ssl.tar"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(archive-security) git:from-archive enforces archive-max-size property" {
  create_safe_multi_file_tar "$ARCHIVE_TMP_DIR/safe.tar"
  run /bin/bash -c "dokku git:set --global archive-max-size 1"
  assert_success

  run /bin/bash -c "cat $ARCHIVE_TMP_DIR/safe.tar | dokku git:from-archive $TEST_APP --"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Archive exceeds maximum allowed size of 1 bytes"

  run /bin/bash -c "dokku git:set --global archive-max-size"
  assert_success
}

@test "(archive-security) git:from-archive enforces archive-max-files property" {
  create_safe_multi_file_tar "$ARCHIVE_TMP_DIR/safe.tar"
  run /bin/bash -c "dokku git:set --global archive-max-files 1"
  assert_success

  run /bin/bash -c "cat $ARCHIVE_TMP_DIR/safe.tar | dokku git:from-archive $TEST_APP --"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "exceeds the maximum of 1"

  run /bin/bash -c "dokku git:set --global archive-max-files"
  assert_success
}

@test "(archive-security) git:report exposes archive-max-size and archive-max-files" {
  run /bin/bash -c "dokku git:report --global --git-global-archive-max-size"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku git:report --global --git-computed-archive-max-size"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1073741824"

  run /bin/bash -c "dokku git:report --global --git-global-archive-max-files"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku git:report --global --git-computed-archive-max-files"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "10000"

  run /bin/bash -c "dokku git:set --global archive-max-size 2147483648"
  assert_success
  run /bin/bash -c "dokku git:report --global --git-global-archive-max-size"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2147483648"

  run /bin/bash -c "dokku git:report --global --git-computed-archive-max-size"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2147483648"

  run /bin/bash -c "dokku git:set --global archive-max-size"
  assert_success
}
