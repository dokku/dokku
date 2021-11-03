#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f ${DOKKU_ROOT}/ENV ]] && mv -f ${DOKKU_ROOT}/ENV ${DOKKU_ROOT}/ENV.bak
  sudo -H -u dokku /bin/bash -c "echo 'export global_test=true' > ${DOKKU_ROOT}/ENV"
  create_app
}

teardown() {
  destroy_app
  ls -la ${DOKKU_ROOT}
  if [[ -f ${DOKKU_ROOT}/ENV.bak ]];then
    mv -f ${DOKKU_ROOT}/ENV.bak ${DOKKU_ROOT}/ENV
  fi
  global_teardown
}

@test "(config) config:help" {
  run /bin/bash -c "dokku config:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage global and app-specific config vars"
}

@test "(config) config:set --global" {
  run ssh dokku@dokku.me config:set --global test_var=true test_var2=\"hello world\" test_var3='double\"quotes'
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(config) config:get --global" {
  run ssh dokku@dokku.me config:set --global test_var=true test_var2=\"hello world\" test_var3=\"with\\nnewline\" test_var4='double\"quotes'
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config:get --global test_var2"
  echo "output: $output"
  echo "status: $status"
  assert_output 'hello world'
  run /bin/bash -c "dokku config:get --global test_var3"
  echo "output: $output"
  echo "status: $status"
  assert_output 'with\nnewline'
  run /bin/bash -c "dokku config:get --global test_var4 | grep 'double\"quotes'"
  echo "output: $output"
  echo "status: $status"
  assert_output 'double"quotes'
}

@test "(config) config:unset --global" {
  run ssh dokku@dokku.me config:set --global test_var=true test_var2=\"hello world\"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config:get --global test_var"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config:unset --global test_var"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config:get --global test_var"
  echo "output: $output"
  echo "status: $status"
  assert_output ""
}

@test "(config) config:set/get" {
  run ssh dokku@dokku.me config:set $TEST_APP test_var1=true test_var2=\"hello world\" test_var3='double\"quotes'
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output " !     At least one env pair must be given"
  assert_failure

  run /bin/bash -c "dokku config:get $TEST_APP test_var1 | grep true"
  echo "output: $output"
  echo "status: $status"
  assert_output "true"
  run /bin/bash -c "dokku config:get $TEST_APP test_var2 | grep 'hello world'"
  echo "output: $output"
  echo "status: $status"
  assert_output "hello world"
  run /bin/bash -c "dokku config:get $TEST_APP test_var3 | grep 'double\"quotes'"
  echo "output: $output"
  echo "status: $status"
  assert_output 'double"quotes'
}

@test "(config) config:set/get (with --app)" {
  run /bin/bash -c "dokku --app $TEST_APP config:set test_var1=true test_var2=\"hello world\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku --app $TEST_APP config:set"
  echo "output: $output"
  echo "status: $status"
  assert_output " !     At least one env pair must be given"
  assert_failure

  run /bin/bash -c "dokku --app $TEST_APP config:get test_var1 | grep true"
  echo "output: $output"
  echo "status: $status"
  assert_output "true"
  run /bin/bash -c "dokku --app $TEST_APP config:get test_var2 | grep 'hello world'"
  echo "output: $output"
  echo "status: $status"
  assert_output "hello world"
}

@test "(config) config:clear" {
  run ssh dokku@dokku.me config:set $TEST_APP test_var=true test_var2=\"hello world\" test_var3=\"with\\nnewline\"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config:get $TEST_APP test_var"
  echo "output: $output"
  echo "status: $status"
  assert_output ""
}

@test "(config) config:unset" {
  run ssh dokku@dokku.me config:set $TEST_APP test_var=true test_var2=\"hello world\" test_var3=\"with\\nnewline\"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config:get $TEST_APP test_var"
  echo "output: $output"
  echo "status: $status"
  assert_output "true"
  run /bin/bash -c "dokku config:unset $TEST_APP test_var"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config:get $TEST_APP test_var"
  echo "output: $output"
  echo "status: $status"
  assert_output ""
  run /bin/bash -c "dokku config:get $TEST_APP test_var3"
  echo "output: $output"
  echo "status: $status"
  assert_output 'with\nnewline'
  run /bin/bash -c "dokku config:unset $TEST_APP test_var3"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku config:get $TEST_APP test_var3"
  echo "output: $output"
  echo "status: $status"
  assert_output ""
  run /bin/bash -c "dokku config:unset $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output " !     At least one key must be given"
  assert_failure
}

@test "(config) global config (herokuish)" {
  run /bin/bash -c "dokku config:set --global HASURA_GRAPHQL_JWT_SECRET='{ \"type\": \"HS256\", \"key\": \"347a4efd2dbb6b91aebf38db5dcf2c4e\" }'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set --global VAR='\$123*&456$'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku run $TEST_APP env | grep -E '^HASURA_GRAPHQL_JWT_SECRET='"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains '{ "type": "HS256", "key": "347a4efd2dbb6b91aebf38db5dcf2c4e" }'

  run /bin/bash -c "dokku run $TEST_APP env | grep -E '^VAR='"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains '123*&456$'

  run /bin/bash -c "dokku run $TEST_APP env | grep -E '^global_test=true'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(config) global config (dockerfile)" {
  run deploy_app dockerfile
  run /bin/bash -c "dokku run $TEST_APP env | grep -E '^global_test=true'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(config) config:show" {
  run /bin/bash -c "dokku --app $TEST_APP config:set zKey=true bKey=true BKEY=true aKey=true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --app $TEST_APP config:show"
  echo "output: $output"
  echo "status: $status"
  assert_output "=====> $TEST_APP env vars"$'\nBKEY:  true\naKey:  true\nbKey:  true\nzKey:  true'
}

@test "(config) config:export" {
  run /bin/bash -c "dokku --app $TEST_APP config:set zKey=true bKey=true BKEY=true aKey=true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:export --format docker-args $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output "--env=BKEY='true' --env=aKey='true' --env=bKey='true' --env=zKey='true'"

  run /bin/bash -c "dokku config:export --format shell $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output "BKEY='true' aKey='true' bKey='true' zKey='true' "

  run /bin/bash -c "dokku config:export --format json $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output '{"BKEY":"true","aKey":"true","bKey":"true","zKey":"true"}'

  run /bin/bash -c "dokku config:export --format json-list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output '[{"name":"BKEY","value":"true"},{"name":"aKey","value":"true"},{"name":"bKey","value":"true"},{"name":"zKey","value":"true"}]'
}
