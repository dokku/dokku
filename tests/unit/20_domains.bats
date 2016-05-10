#!/usr/bin/env bats

load test_helper

setup() {
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  [[ -f "$DOKKU_ROOT/HOSTNAME" ]] && cp -fp "$DOKKU_ROOT/HOSTNAME" "$DOKKU_ROOT/HOSTNAME.bak"
  create_app
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  [[ -f "$DOKKU_ROOT/HOSTNAME.bak" ]] && mv "$DOKKU_ROOT/HOSTNAME.bak" "$DOKKU_ROOT/HOSTNAME" && chown dokku:dokku "$DOKKU_ROOT/HOSTNAME"
}

@test "(domains) domains" {
  dokku domains:setup $TEST_APP
  run bash -c "dokku domains $TEST_APP | grep ${TEST_APP}.dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_output "${TEST_APP}.dokku.me"
}

@test "(domains) domains:add" {
  run dokku domains:add $TEST_APP www.test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku domains:add $TEST_APP test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(domains) domains:add (multiple)" {
  run dokku domains:add $TEST_APP www.test.app.dokku.me test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(domains) domains:add (duplicate)" {
  run dokku domains:add $TEST_APP test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success

  run dokku domains:add $TEST_APP test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(domains) domains:remove" {
  run dokku domains:add $TEST_APP test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku domains:remove $TEST_APP test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  refute_line "test.app.dokku.me"
}

@test "(domains) domains:remove (wildcard domain)" {
  run dokku domains:add $TEST_APP *.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku domains:remove $TEST_APP *.dokku.me
  echo "output: "$output
  echo "status: "$status
  refute_line "*.dokku.me"
}

@test "(domains) domains:clear" {
  run dokku domains:add $TEST_APP test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku domains:clear $TEST_APP
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(domains) domains:set-global" {
  run dokku domains:set-global global.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success

  run bash -c "dokku domains | grep -q global.dokku.me"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
