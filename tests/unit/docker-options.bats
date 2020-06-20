#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(docker-options) docker-options:help" {
  run /bin/bash -c "dokku docker-options"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage docker options for an app"
  help_output="$output"

  run /bin/bash -c "dokku docker-options:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage docker options for an app"
  assert_output "$help_output"
}

@test "(docker-options) docker-options:add (all phases)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-build"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp"
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp"
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-run"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp"
}

@test "(docker-options) docker-options:clear" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-build"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 0
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 0
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-run"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 0

  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:clear $TEST_APP build"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-build"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 0
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp"
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-run"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp"

  run /bin/bash -c "dokku docker-options:clear $TEST_APP deploy"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-build"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 0
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 0
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-run"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp"

  run /bin/bash -c "dokku docker-options:clear $TEST_APP run"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-build"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 0
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-deploy"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 0
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-run"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 0
}

@test "(docker-options) docker-options:add (build phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP --docker-options-build"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp"
}

@test "(docker-options) docker-options:add (deploy phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP  --docker-options-deploy"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp"
}

@test "(docker-options) docker-options:add (run phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP run \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP  --docker-options-run"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp"
}

@test "(docker-options) docker-options:remove (all phases)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 3
  run /bin/bash -c "dokku docker-options:remove $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 0
  assert_output_contains "Docker options deploy:         --restart=on-failure:10"
}

@test "(docker-options) docker-options:remove (build phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 3
  run /bin/bash -c "dokku docker-options:remove $TEST_APP build \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 2
}

@test "(docker-options) docker-options:remove (deploy phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 3
  run /bin/bash -c "dokku docker-options:remove $TEST_APP deploy \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 2
}

@test "(docker-options) docker-options:remove (run phase)" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP build,deploy,run \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 3
  run /bin/bash -c "dokku docker-options:remove $TEST_APP run \"-v /tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 2
}

@test "(docker-options) deploy with options" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy \"-v /var/tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "echo '-v /tmp' >> $DOKKU_ROOT/$TEST_APP/DOCKER_OPTIONS_DEPLOY"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "echo '# comment' >> $DOKKU_ROOT/$TEST_APP/DOCKER_OPTIONS_DEPLOY"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 1
  deploy_app

  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect -f '{{ .Config.Volumes }}' $CID | sed -e 's:map::g' | tr -d '[]' | tr ' ' $'\n' | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "/tmp:{} /var/tmp:{}"
}

@test "(docker-options) docker-options:add (all phases over SSH)" {
  run ssh dokku@dokku.me docker-options:add $TEST_APP build,deploy,run "-v /tmp"
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "-v /tmp" 3
}

@test "(docker-options) dockerfile deploy with link" {
  run /bin/bash -c "dokku docker-options:add $TEST_APP deploy \"-v /var/tmp\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  run /bin/bash -c "dokku docker-options:add $TEST_APP build \"--link postgres\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  deploy_app dockerfile
}
