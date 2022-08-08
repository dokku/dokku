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
  if [[ -f ${DOKKU_ROOT}/ENV.bak ]]; then
    mv -f ${DOKKU_ROOT}/ENV.bak ${DOKKU_ROOT}/ENV
  fi
  global_teardown
}

@test "(config-oddities) set-local/get with multiple spaces and $" {
  run /bin/bash -c "dokku config:set --global double_quotes=\"hello  world$\" single_quotes='hello$  world'"
  echo "output: $output"
  echo "status: $status"
  run /bin/bash -c "dokku config:get --global double_quotes"
  echo "output: $output"
  echo "status: $status"
  assert_output 'hello  world$'
  run /bin/bash -c "dokku config:get --global single_quotes"
  echo "output: $output"
  echo "status: $status"
  assert_output 'hello$  world'
  assert_success
}

@test "(config-oddities) set-local/get with multiple lines" {
  multiline='line one
  line two'
  run /bin/bash -c "dokku config:set --global double_quotes=\"$multiline\""
  echo "output: $output"
  echo "status: $status"
  run /bin/bash -c "dokku config:get --global double_quotes"
  echo "output: $output"
  echo "status: $status"
  assert_output "$multiline"
  assert_success
}
