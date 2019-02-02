#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  [[ -f "$DOKKU_ROOT/HOSTNAME" ]] && cp -fp "$DOKKU_ROOT/HOSTNAME" "$DOKKU_ROOT/HOSTNAME.bak"
  create_app
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  [[ -f "$DOKKU_ROOT/HOSTNAME.bak" ]] && mv "$DOKKU_ROOT/HOSTNAME.bak" "$DOKKU_ROOT/HOSTNAME" && chown dokku:dokku "$DOKKU_ROOT/HOSTNAME"
  global_teardown
}

@test "(domains) domains" {
  dokku domains:setup $TEST_APP
  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null | grep ${TEST_APP}.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_output "${TEST_APP}.dokku.me"
}

@test "(domains) domains:add" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP 2.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP a--domain.with--hyphens"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line www.test.app.dokku.me
  assert_line test.app.dokku.me
  assert_line 2.app.dokku.me
  assert_line a--domain.with--hyphens
}

@test "(domains) domains:add (multiple)" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.dokku.me test.app.dokku.me 2.app.dokku.me a--domain.with--hyphens"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line www.test.app.dokku.me
  assert_line test.app.dokku.me
  assert_line 2.app.dokku.me
  assert_line a--domain.with--hyphens
}

@test "(domains) domains:add (duplicate)" {
  run /bin/bash -c "dokku domains:add $TEST_APP test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) domains:add (invalid)" {
  run /bin/bash -c "dokku domains:add $TEST_APP http://test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(domains) domains:remove" {
  run /bin/bash -c "dokku domains:add $TEST_APP test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:remove $TEST_APP test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  refute_line test.app.dokku.me
}

@test "(domains) domains:remove (multiple)" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.dokku.me test.app.dokku.me 2.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:remove $TEST_APP www.test.app.dokku.me test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  refute_line www.test.app.dokku.me
  refute_line test.app.dokku.me
  assert_line 2.app.dokku.me
}

@test "(domains) domains:remove (wildcard domain)" {
  run /bin/bash -c "dokku domains:add $TEST_APP *.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:remove $TEST_APP *.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  refute_line *.dokku.me
}

@test "(domains) domains:set" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.dokku.me test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP 2.app.dokku.me a--domain.with--hyphens"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  refute_line www.test.app.dokku.me
  refute_line test.app.dokku.me
  assert_line 2.app.dokku.me
  assert_line a--domain.with--hyphens
}

@test "(domains) domains:clear" {
  run /bin/bash -c "dokku domains:add $TEST_APP test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  refute_line test.app.dokku.me
}

@test "(domains) domains:add-global" {
  run /bin/bash -c "dokku domains:add-global global.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains 2>/dev/null | egrep -qw '^global.dokku.me\$'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) domains:add-global (multiple)" {
  run /bin/bash -c "dokku domains:add-global global1.dokku.me global2.dokku.me global3.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains 2>/dev/null | grep -q global1.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains 2>/dev/null | grep -q global2.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains 2>/dev/null | grep -q global3.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) domains:remove-global" {
  run /bin/bash -c "dokku domains:add-global global.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:remove-global global.dokku.me"
  echo "output: $output"
  echo "status: $status"
  refute_line "global.dokku.me"
}

@test "(domains) domains (multiple global domains)" {
  run /bin/bash -c "dokku domains:add-global global1.dokku.me global2.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku domains:setup $TEST_APP

  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_line ${TEST_APP}.global1.dokku.me
  assert_line ${TEST_APP}.global2.dokku.me
}

@test "(domains) domains:set-global" {
  run /bin/bash -c "dokku domains:add-global global1.dokku.me global2.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set-global global3.dokku.me global4.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  dokku domains:setup $TEST_APP

  run /bin/bash -c "dokku domains $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  refute_line ${TEST_APP}.global1.dokku.me
  refute_line ${TEST_APP}.global2.dokku.me
  assert_line ${TEST_APP}.global3.dokku.me
  assert_line ${TEST_APP}.global4.dokku.me
}
