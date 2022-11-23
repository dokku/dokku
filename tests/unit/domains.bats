#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(domains) domains:help" {
  run /bin/bash -c "dokku domains"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage domains used by the proxy"
  help_output="$output"

  run /bin/bash -c "dokku domains:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage domains used by the proxy"
  assert_output "$help_output"
}

@test "(domains) domains" {
  dokku domains:setup $TEST_APP
  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null | grep ${TEST_APP}.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "${TEST_APP}.dokku.me"
}

@test "(domains) domains:add" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.dokku.me"
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

  run /bin/bash -c "dokku domains:add $TEST_APP .dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains www.test.app.dokku.me
  assert_output_contains 2.app.dokku.me
  assert_output_contains a--domain.with--hyphens
}

@test "(domains) domains:add (multiple)" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.dokku.me 2.app.dokku.me a--domain.with--hyphens"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains www.test.app.dokku.me
  assert_output_contains 2.app.dokku.me
  assert_output_contains a--domain.with--hyphens
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

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains test.app.dokku.me 0
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

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains www.test.app.dokku.me 0
  assert_output_contains test.app.dokku.me 0
  assert_output_contains 2.app.dokku.me
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

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains *.dokku.me 0
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

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains www.test.app.dokku.me 0
  assert_output_contains test.app.dokku.me 0
  assert_output_contains 2.app.dokku.me
  assert_output_contains a--domain.with--hyphens
}

@test "(domains) domains:clear" {
  run /bin/bash -c "dokku domains:add $TEST_APP test.app.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains test.app.dokku.me 0
}

@test "(domains) domains:add-global" {
  run /bin/bash -c "dokku domains:add-global global.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report --global"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report 2>/dev/null | grep -qw 'global.dokku.me'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) domains:add-global (multiple)" {
  run /bin/bash -c "dokku domains:add-global global1.dokku.me global2.dokku.me global3.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report 2>/dev/null | grep -q global1.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report 2>/dev/null | grep -q global2.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report 2>/dev/null | grep -q global3.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) domains:clear-global" {
  run /bin/bash -c "dokku domains:add-global global.dokku.invalid"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add-global global.dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:clear-global"
  echo "output: $output"
  echo "status: $status"
  assert_success
  refute_line "global.dokku.invalid"
  refute_line "global.dokku.me"
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

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains ${TEST_APP}.global1.dokku.me
  assert_output_contains ${TEST_APP}.global2.dokku.me
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

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains ${TEST_APP}.global1.dokku.me 0
  assert_output_contains ${TEST_APP}.global2.dokku.me 0
  assert_output_contains ${TEST_APP}.global3.dokku.me
  assert_output_contains ${TEST_APP}.global4.dokku.me
}

@test "(domains) app name overlaps with global domain.tld" {
  run /bin/bash -c "dokku domains:set-global dokku.test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # run domains:clear in order to invoke default vhost creation
  dokku --quiet apps:create test.dokku.test
  dokku --quiet domains:clear test.dokku.test

  run /bin/bash -c "dokku domains:report test.dokku.test --domains-app-vhosts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test.dokku.test"

  dokku --force apps:destroy test.dokku.test
}

@test "(domains) app rename only renames domains associated with global domains" {
  run /bin/bash -c "dokku domains:set-global dokku.me"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.dokku.me $TEST_APP.dokku.test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:rename $TEST_APP other-name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report other-name --domains-app-vhosts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "other-name.dokku.me $TEST_APP.dokku.test"

  run /bin/bash -c "dokku apps:rename other-name $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) verify warning on ipv4/ipv6 domain name" {
  touch /etc/nginx/sites-enabled/default
  rm "$DOKKU_ROOT/VHOST"
  echo "127.0.0.1" >"$DOKKU_ROOT/VHOST"
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Detected IPv4 domain name with nginx proxy enabled."

  rm -f /etc/nginx/sites-enabled/default
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Detected IPv4 domain name with nginx proxy enabled." 0
}
