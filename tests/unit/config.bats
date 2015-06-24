#!/usr/bin/env bats

load test_helper

setup() {
  sudo -H -u dokku /bin/bash -c "echo 'export global_test=true' > $DOKKU_ROOT/ENV"
  create_app
}

teardown() {
  destroy_app
  rm -f "$DOKKU_ROOT/ENV"
}

@test "(config) config:set" {
  run ssh dokku@dokku.me config:set $TEST_APP test_var=true test_var2=\"hello world\"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(config) config:get" {
  run ssh dokku@dokku.me config:set $TEST_APP test_var=true test_var2=\"hello world\" test_var3=\"with\\nnewline\"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku config:get $TEST_APP test_var2
  echo "output: "$output
  echo "status: "$status
  assert_output 'hello world'
  run dokku config:get $TEST_APP test_var3
  echo "output: "$output
  echo "status: "$status
  assert_output 'with\nnewline'
}

@test "(config) config:unset" {
  run ssh dokku@dokku.me config:set $TEST_APP test_var=true test_var2=\"hello world\"
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku config:get $TEST_APP test_var
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku config:unset $TEST_APP test_var
  echo "output: "$output
  echo "status: "$status
  assert_success
  run dokku config:get $TEST_APP test_var
  echo "output: "$output
  echo "status: "$status
  assert_output ""
}

@test "(config) global config (buildstep)" {
  deploy_app
  run bash -c "dokku run $TEST_APP env | egrep '^global_test=true'"
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(config) global config (dockerfile)" {
  deploy_app dockerfile
  run bash -c "dokku run $TEST_APP env | egrep '^global_test=true'"
  echo "output: "$output
  echo "status: "$status
  assert_success
}
