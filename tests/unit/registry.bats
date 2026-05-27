#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
  dokku checks:set $TEST_APP wait-to-retire 30
}

teardown() {
  destroy_app
  global_teardown
}

@test "(registry) registry:help" {
  run /bin/bash -c "dokku registry"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage registry settings for an app"
  help_output="$output"

  run /bin/bash -c "dokku registry:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage registry settings for an app"
  assert_output "$help_output"
}

@test "(registry:login) global login with deprecated warning" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  run /bin/bash -c "dokku registry:login docker.io $DOCKERHUB_USERNAME $DOCKERHUB_TOKEN"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Login Succeeded"
  assert_output_contains "Deprecated: please use --global flag"

  run /bin/bash -c "echo $DOCKERHUB_TOKEN | dokku registry:login docker.io --password-stdin $DOCKERHUB_USERNAME"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Login Succeeded"
}

@test "(registry:login) global login with --global flag" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  run /bin/bash -c "dokku registry:login --global docker.io $DOCKERHUB_USERNAME $DOCKERHUB_TOKEN"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Login Succeeded"
  assert_output_not_contains "Deprecated"
}

@test "(registry:login) per-app login" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  run /bin/bash -c "dokku registry:login $TEST_APP docker.io $DOCKERHUB_USERNAME $DOCKERHUB_TOKEN"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Login Succeeded"

  run /bin/bash -c "test -f /var/lib/dokku/config/registry/$TEST_APP/config.json"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(registry:logout) per-app logout" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  run /bin/bash -c "dokku registry:login $TEST_APP docker.io $DOCKERHUB_USERNAME $DOCKERHUB_TOKEN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:logout $TEST_APP docker.io"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(registry:logout) global logout" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  run /bin/bash -c "dokku registry:login --global docker.io $DOCKERHUB_USERNAME $DOCKERHUB_TOKEN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:logout --global docker.io"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(registry) per-app credentials deleted on app destroy" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  run /bin/bash -c "dokku registry:login $TEST_APP docker.io $DOCKERHUB_USERNAME $DOCKERHUB_TOKEN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -d /var/lib/dokku/config/registry/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  destroy_app

  run /bin/bash -c "test -d /var/lib/dokku/config/registry/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  create_app
}

@test "(registry:report) image-repo raw vs computed vs global" {
  run /bin/bash -c "dokku registry:set --global image-repo-template"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-image-repo\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-global-image-repo-template\"'"
  assert_success
  assert_output ""

  # When neither per-app image-repo nor a global image-repo-template are set,
  # reportComputedImageRepo falls back to common.GetAppImageRepo which returns
  # the default dokku/<app> image name.
  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-image-repo\"'"
  assert_success
  assert_output "dokku/$TEST_APP"

  run /bin/bash -c "dokku registry:set --global image-repo-template 'dokku/{{ APP }}'"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-global-image-repo-template\"'"
  assert_success
  assert_output 'dokku/{{ APP }}'

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-image-repo\"'"
  assert_success
  assert_output "dokku/$TEST_APP"

  run /bin/bash -c "dokku registry:set $TEST_APP image-repo my/$TEST_APP"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-image-repo\"'"
  assert_success
  assert_output "my/$TEST_APP"

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-image-repo\"'"
  assert_success
  assert_output "my/$TEST_APP"

  run /bin/bash -c "dokku registry:set $TEST_APP image-repo"
  assert_success

  run /bin/bash -c "dokku registry:set --global image-repo-template"
  assert_success
}

@test "(registry:report) image-repo-template raw vs computed vs global" {
  run /bin/bash -c "dokku registry:set --global image-repo-template"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-image-repo-template\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-global-image-repo-template\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-image-repo-template\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:set --global image-repo-template 'global-prefix/{{ .AppName }}'"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-image-repo-template\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-global-image-repo-template\"'"
  assert_success
  assert_output 'global-prefix/{{ .AppName }}'

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-image-repo-template\"'"
  assert_success
  assert_output 'global-prefix/{{ .AppName }}'

  # The global template renders into the computed image-repo.
  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-image-repo\"'"
  assert_success
  assert_output "global-prefix/$TEST_APP"

  # Per-app template overrides the global template.
  run /bin/bash -c "dokku registry:set $TEST_APP image-repo-template 'app-prefix/{{ .AppName }}-prod'"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-image-repo-template\"'"
  assert_success
  assert_output 'app-prefix/{{ .AppName }}-prod'

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-global-image-repo-template\"'"
  assert_success
  assert_output 'global-prefix/{{ .AppName }}'

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-image-repo-template\"'"
  assert_success
  assert_output 'app-prefix/{{ .AppName }}-prod'

  # The per-app template flows through to the computed image-repo.
  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-image-repo\"'"
  assert_success
  assert_output "app-prefix/$TEST_APP-prod"

  # Unsetting the per-app template restores the global template behavior.
  run /bin/bash -c "dokku registry:set $TEST_APP image-repo-template"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-image-repo-template\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-image-repo-template\"'"
  assert_success
  assert_output 'global-prefix/{{ .AppName }}'

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-image-repo\"'"
  assert_success
  assert_output "global-prefix/$TEST_APP"

  run /bin/bash -c "dokku registry:set --global image-repo-template"
  assert_success
}

@test "(registry:report) push-on-release raw vs computed vs global" {
  run /bin/bash -c "dokku registry:set --global push-on-release"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-push-on-release\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-global-push-on-release\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-push-on-release\"'"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku registry:set --global push-on-release true"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-global-push-on-release\"'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-push-on-release\"'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku registry:set $TEST_APP push-on-release false"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-push-on-release\"'"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-push-on-release\"'"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku registry:set $TEST_APP push-on-release"
  assert_success

  run /bin/bash -c "dokku registry:set --global push-on-release"
  assert_success
}

@test "(registry:report --global) computed and global keys" {
  run /bin/bash -c "dokku registry:set --global push-on-release"
  assert_success
  run /bin/bash -c "dokku registry:set --global server"
  assert_success
  run /bin/bash -c "dokku registry:set --global image-repo-template"
  assert_success
  run /bin/bash -c "dokku registry:set --global push-extra-tags"
  assert_success

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-global-push-on-release\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-computed-push-on-release\"'"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-global-server\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-computed-server\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-global-image-repo-template\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-computed-image-repo-template\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-global-push-extra-tags\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-computed-push-extra-tags\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:set --global push-on-release true"
  assert_success
  run /bin/bash -c "dokku registry:set --global server ghcr.io"
  assert_success
  run /bin/bash -c "dokku registry:set --global image-repo-template 'dokku/{{ APP }}'"
  assert_success
  run /bin/bash -c "dokku registry:set --global push-extra-tags latest,stable"
  assert_success

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-global-push-on-release\"'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-computed-push-on-release\"'"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-global-server\"'"
  assert_success
  assert_output "ghcr.io"

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-computed-server\"'"
  assert_success
  assert_output "ghcr.io/"

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-global-image-repo-template\"'"
  assert_success
  assert_output 'dokku/{{ APP }}'

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-computed-image-repo-template\"'"
  assert_success
  assert_output 'dokku/{{ APP }}'

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-global-push-extra-tags\"'"
  assert_success
  assert_output "latest,stable"

  run /bin/bash -c "dokku registry:report --global --format json | jq -r '.\"registry-computed-push-extra-tags\"'"
  assert_success
  assert_output "latest,stable"

  run /bin/bash -c "dokku registry:set --global push-on-release"
  assert_success
  run /bin/bash -c "dokku registry:set --global server"
  assert_success
  run /bin/bash -c "dokku registry:set --global image-repo-template"
  assert_success
  run /bin/bash -c "dokku registry:set --global push-extra-tags"
  assert_success
}

@test "(registry:report) push-extra-tags raw vs computed vs global" {
  run /bin/bash -c "dokku registry:set --global push-extra-tags"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-push-extra-tags\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-global-push-extra-tags\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-push-extra-tags\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku registry:set --global push-extra-tags global-tag"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-global-push-extra-tags\"'"
  assert_success
  assert_output "global-tag"

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-push-extra-tags\"'"
  assert_success
  assert_output "global-tag"

  run /bin/bash -c "dokku registry:set $TEST_APP push-extra-tags release,stable"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-push-extra-tags\"'"
  assert_success
  assert_output "release,stable"

  run /bin/bash -c "dokku registry:report $TEST_APP --format json | jq -r '.\"registry-computed-push-extra-tags\"'"
  assert_success
  assert_output "release,stable"

  run /bin/bash -c "dokku registry:set $TEST_APP push-extra-tags"
  assert_success

  run /bin/bash -c "dokku registry:set --global push-extra-tags"
  assert_success
}

@test "(registry) registry:set server" {
  run /bin/bash -c "dokku registry:set --global server ghcr.io"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-global-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ghcr.io"

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-computed-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ghcr.io/"

  run /bin/bash -c "dokku registry:set --global server docker.io"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-global-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker.io"

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-computed-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku registry:set $TEST_APP server docker.io"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-computed-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_exists

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-global-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker.io"

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker.io"

  run /bin/bash -c "dokku registry:set $TEST_APP server ghcr.io"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-computed-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ghcr.io/"

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-global-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker.io"

  run /bin/bash -c "dokku registry:report $TEST_APP --registry-server"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "ghcr.io"
}

@test "(registry) registry:set image-repo" {
  run /bin/bash -c "docker images"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set $TEST_APP image-repo heroku/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker inspect heroku/$TEST_APP:latest"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker images"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(registry) registry:set push-on-release" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  run /bin/bash -c "dokku registry:set $TEST_APP push-on-release true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set $TEST_APP image-repo dokku/test-app"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 60

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container ls -a"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image ls"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 60

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container ls -a"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image ls"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:set $TEST_APP key=VALUE"
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 60

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container ls -a"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image ls"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  sleep 60

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:retire"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker container ls -a"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image ls"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker image inspect dokku/test-app:1"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "docker image inspect dokku/test-app:2"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "docker image inspect dokku/test-app:3"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "docker image inspect dokku/test-app:4"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(registry) registry:set push-extra-tags" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  run /bin/bash -c "dokku registry:set $TEST_APP push-on-release true"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set $TEST_APP image-repo dokku/test-app"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku registry:set $TEST_APP push-extra-tags foo"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku/test-app:foo"
}
