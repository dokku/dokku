#!/usr/bin/env bats

load test_helper

SMOKE_TEST_APP_1_0_0_SHA=9cf71bba639c4f1671dfd42685338b762d3354f2
SMOKE_TEST_APP_2_0_0_SHA=5c8a5e42bbd7fae98bd657fb17f41c6019b303f9
SMOKE_TEST_APP_ANOTHER_BRANCH_SHA=5c8a5e42bbd7fae98bd657fb17f41c6019b303f9
SMOKE_TEST_APP_MASTER_SHA=af1b02052199b8ca8115f80ff8676a7c7744a45f
SMOKE_TEST_APP_COMMIT_SHA=5c8a5e42bbd7fae98bd657fb17f41c6019b303f9

setup() {
  global_setup
  create_app
  [[ -f "$DOKKU_ROOT/.netrc" ]] && cp -fp "$DOKKU_ROOT/.netrc" "$DOKKU_ROOT/.netrc.bak"
  touch /home/dokku/.netrc
  chown dokku:dokku /home/dokku/.netrc
  touch /home/dokku/.ssh/known_hosts
  chown dokku:dokku /home/dokku/.ssh/known_hosts
}

teardown() {
  rm -f /home/dokku/.ssh/id_rsa.pub || true
  [[ -f "$DOKKU_ROOT/.netrc.bak" ]] && mv "$DOKKU_ROOT/.netrc.bak" "$DOKKU_ROOT/.netrc" && chown dokku:dokku "$DOKKU_ROOT/.netrc"
  destroy_app
  global_teardown
}

@test "(git) git:allow-host" {
  run /bin/bash -c "dokku git:allow-host"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "cat /home/dokku/.ssh/known_hosts | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  start_lines=$output

  run /bin/bash -c "dokku git:allow-host github.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/.ssh/known_hosts | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -f /home/dokku/.ssh/known_hosts"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/.ssh/known_hosts | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "$((start_lines + 1))"

  run /bin/bash -c "dokku git:allow-host github.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/.ssh/known_hosts | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_equal "$output" "$((start_lines + 2))"
}

@test "(git) git:auth" {
  run /bin/bash -c "dokku git:auth"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku git:auth github.com"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Removing netrc auth entry for host github.com"

  run /bin/bash -c "dokku git:auth github.com username"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Missing password for netrc auth entry"

  run /bin/bash -c "dokku git:auth github.com username password"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting netrc auth entry for host github.com"
}

@test "(git) git:sync new [errors]" {
  run /bin/bash -c "dokku git:sync"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku git:sync $TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run create_app "$TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:sync $TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run destroy_app 0 "$TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:sync new [--no-build noarg]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_MASTER_SHA"
}

@test "(git) git:sync new [--no-build branch]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git another-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/another-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_ANOTHER_BRANCH_SHA"
}

@test "(git) git:sync new [--no-build tag]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_1_0_0_SHA"
}

@test "(git) git:sync new [--no-build commit]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git $SMOKE_TEST_APP_COMMIT_SHA"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_COMMIT_SHA"
}

@test "(git) git:sync new [--build noarg]" {
  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run /bin/bash -c "dokku config:get $TEST_APP GIT_REV"
  echo "output: $output"
  echo "status: $status"
  assert_output_exists

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_MASTER_SHA"

  run /bin/bash -c "dokku git:sync --build-if-changes $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Skipping build as no changes were detected"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_MASTER_SHA"
}

@test "(git) git:sync new [--build branch]" {
  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git another-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/another-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_ANOTHER_BRANCH_SHA"

  run /bin/bash -c "dokku git:sync --build-if-changes $TEST_APP https://github.com/dokku/smoke-test-app.git another-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Skipping build as no changes were detected"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/another-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_ANOTHER_BRANCH_SHA"
}

@test "(git) git:sync new [--build tag]" {
  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_1_0_0_SHA"

  run /bin/bash -c "dokku git:sync --build-if-changes $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Skipping build as no changes were detected"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_1_0_0_SHA"
}

@test "(git) git:sync new [--build commit]" {
  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git $SMOKE_TEST_APP_COMMIT_SHA"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_COMMIT_SHA"

  run /bin/bash -c "dokku git:sync --build-if-changes $TEST_APP https://github.com/dokku/smoke-test-app.git $SMOKE_TEST_APP_COMMIT_SHA"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Skipping build as no changes were detected"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_COMMIT_SHA"
}

@test "(git) git:sync existing [errors]" {
  run /bin/bash -c "dokku git:sync"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku git:sync $TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run create_app "$TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:sync $TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run destroy_app 0 "$TEST_APP-non-existent"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(git) git:sync existing [--no-build noarg]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_1_0_0_SHA"

  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Fetching remote code for"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_MASTER_SHA"
}

@test "(git) git:sync existing [--no-build branch]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_1_0_0_SHA"

  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git another-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/another-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_ANOTHER_BRANCH_SHA"
}

@test "(git) git:sync existing [--no-build tag]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_1_0_0_SHA"

  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 2.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_2_0_0_SHA"
}

@test "(git) git:sync existing [--no-build annotated-tag]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_1_0_0_SHA"

  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 2.0.0-annotated"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_2_0_0_SHA"
}

@test "(git) git:sync existing [--no-build commit]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_1_0_0_SHA"

  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git $SMOKE_TEST_APP_COMMIT_SHA"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_COMMIT_SHA"
}

@test "(git) git:sync existing [--build noarg]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_1_0_0_SHA"

  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_MASTER_SHA"

  run /bin/bash -c "dokku git:sync --build-if-changes $TEST_APP https://github.com/dokku/smoke-test-app.git"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Skipping build as no changes were detected"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_MASTER_SHA"
}

@test "(git) git:sync existing [--build branch]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 2.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_2_0_0_SHA"

  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git another-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_ANOTHER_BRANCH_SHA"

  run /bin/bash -c "dokku git:sync --build-if-changes $TEST_APP https://github.com/dokku/smoke-test-app.git another-branch"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Skipping build as no changes were detected"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_ANOTHER_BRANCH_SHA"
}

@test "(git) git:sync existing [--build tag]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_1_0_0_SHA"

  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git 2.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_2_0_0_SHA"

  run /bin/bash -c "dokku git:sync --build-if-changes $TEST_APP https://github.com/dokku/smoke-test-app.git 2.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Skipping build as no changes were detected"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_2_0_0_SHA"
}

@test "(git) git:sync existing [--build commit]" {
  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app.git 1.0.0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_1_0_0_SHA"

  run /bin/bash -c "dokku git:sync --build $TEST_APP https://github.com/dokku/smoke-test-app.git $SMOKE_TEST_APP_COMMIT_SHA"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Application deployed"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_COMMIT_SHA"

  run /bin/bash -c "dokku git:sync --build-if-changes $TEST_APP https://github.com/dokku/smoke-test-app.git $SMOKE_TEST_APP_COMMIT_SHA"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Skipping build as no changes were detected"

  run /bin/bash -c "cat /home/dokku/$TEST_APP/refs/heads/master"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$SMOKE_TEST_APP_COMMIT_SHA"
}

@test "(git) git:sync private" {
  if [[ -z "$SYNC_GITHUB_USERNAME" ]] || [[ -z "$SYNC_GITHUB_PASSWORD" ]]; then
    skip "skipping due to missing github credentials SYNC_GITHUB_USERNAME:SYNC_GITHUB_PASSWORD"
  fi

  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app-private.git"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "fatal: could not read Username for"

  run /bin/bash -c "dokku git:auth github.com $SYNC_GITHUB_USERNAME $SYNC_GITHUB_PASSWORD"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cat /home/dokku/.netrc"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:sync $TEST_APP https://github.com/dokku/smoke-test-app-private.git"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:sync --build-if-changes $TEST_APP https://github.com/dokku/smoke-test-app-private.git"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Skipping build as no changes were detected"
}

@test "(git) git:public-key" {
  run /bin/bash -c "dokku git:public-key"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "There is no deploy key associated with the dokku user."

  run /bin/bash -c "cp /root/.ssh/dokku_test_rsa.pub /home/dokku/.ssh/id_rsa.pub"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:public-key"
  echo "output: $output"
  echo "status: $status"
  assert_success
}
