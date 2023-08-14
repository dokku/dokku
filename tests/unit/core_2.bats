#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  rm -rf /home/dokku/$TEST_APP/tls
  destroy_app
  global_teardown
}

@test "(core) cleanup:help" {
  run /bin/bash -c "dokku cleanup:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Cleans up exited/dead Docker containers and removes dangling image"
}

@test "(core) unknown command" {
  run /bin/bash -c "dokku fakecommand"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku fakecommand 2>&1 | grep -q 'is not a dokku command'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps: 2>&1 | grep -q 'is not a dokku command'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(core) urls (non-ssl)" {
  assert_urls "http://${TEST_APP}.${DOKKU_DOMAIN}"

  run /bin/bash -c "dokku domains:add $TEST_APP test.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_urls "http://${TEST_APP}.${DOKKU_DOMAIN}" "http://test.${DOKKU_DOMAIN}"
}

@test "(core) urls (app ssl)" {
  assert_urls "http://${TEST_APP}.${DOKKU_DOMAIN}"

  run setup_test_tls
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_urls "https://${TEST_APP}.${DOKKU_DOMAIN}"

  run /bin/bash -c "dokku domains:add $TEST_APP test.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_urls "http://${TEST_APP}.${DOKKU_DOMAIN}" "https://${TEST_APP}.${DOKKU_DOMAIN}" "https://test.${DOKKU_DOMAIN}" "http://test.${DOKKU_DOMAIN}"
}

@test "(core) url (app ssl)" {
  run setup_test_tls
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_url "https://${TEST_APP}.${DOKKU_DOMAIN}"
}

@test "(core) urls (wildcard ssl)" {
  run setup_test_tls wildcard
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_urls "https://${TEST_APP}.${DOKKU_DOMAIN}"

  run /bin/bash -c "dokku domains:add $TEST_APP test.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_urls "http://${TEST_APP}.${DOKKU_DOMAIN}" "https://${TEST_APP}.${DOKKU_DOMAIN}" "https://test.${DOKKU_DOMAIN}" "http://test.${DOKKU_DOMAIN}"

  run /bin/bash -c "dokku domains:add $TEST_APP dokku.example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_urls "http://dokku.example.com" "http://${TEST_APP}.${DOKKU_DOMAIN}" "https://dokku.example.com" "https://${TEST_APP}.${DOKKU_DOMAIN}" "https://test.${DOKKU_DOMAIN}" "http://test.${DOKKU_DOMAIN}"
}

@test "(core) git-remote (off-port)" {
  run deploy_app nodejs-express ssh://dokku@127.0.0.1:22333/$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(core) git-remote (bad name)" {
  run deploy_app nodejs-express ssh://dokku@127.0.0.1:22333/home/dokku/$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_failure
}
