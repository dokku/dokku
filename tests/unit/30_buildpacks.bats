#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  deploy_app
}

teardown() {
  destroy_app
  global_teardown
}

@test "(buildpacks) buildpacks:help" {
  run /bin/bash -c "dokku buildpacks"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manages buildpacks settings for an app"
  help_output="$output"

  run /bin/bash -c "dokku buildpacks:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manages buildpacks settings for an app"
  assert_output "$help_output"
}

@test "(buildpacks) buildpacks:add" {
  run /bin/bash -c "dokku buildpacks:add $TEST_APP heroku/nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "heroku/nodejs"

  run /bin/bash -c "dokku buildpacks:add $TEST_APP heroku/ruby"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "heroku/nodejs heroku/ruby"

  run /bin/bash -c "dokku buildpacks:add --index 1 $TEST_APP heroku/golang"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "heroku/golang heroku/nodejs heroku/ruby"

  run /bin/bash -c "dokku buildpacks:add --index 2 $TEST_APP heroku/python"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "heroku/golang heroku/python heroku/nodejs heroku/ruby"

  run /bin/bash -c "dokku buildpacks:add --index 100 $TEST_APP heroku/php"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "heroku/golang heroku/python heroku/nodejs heroku/ruby heroku/php"
}

@test "(buildpacks) buildpacks:set" {
  run /bin/bash -c "dokku buildpacks:set $TEST_APP heroku/nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "heroku/nodejs"

  run /bin/bash -c "dokku buildpacks:set $TEST_APP heroku/ruby"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "heroku/ruby"

  run /bin/bash -c "dokku buildpacks:set --index 1 $TEST_APP heroku/golang"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "heroku/golang"

  run /bin/bash -c "dokku buildpacks:set --index 2 $TEST_APP heroku/python"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "heroku/golang heroku/python"

  run /bin/bash -c "dokku buildpacks:set --index 100 $TEST_APP heroku/php"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "heroku/golang heroku/python heroku/php"
}

@test "(buildpacks) buildpacks:remove" {
  run /bin/bash -c "dokku buildpacks:set $TEST_APP heroku/nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku buildpacks:set --index 2 $TEST_APP heroku/ruby"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "heroku/nodejs heroku/ruby"

  run /bin/bash -c "dokku buildpacks:remove $TEST_APP heroku/nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "heroku/ruby"

  run /bin/bash -c "dokku buildpacks:remove $TEST_APP heroku/php"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku buildpacks:remove $TEST_APP heroku/ruby"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists
}


@test "(buildpacks) buildpacks:clear" {
  run /bin/bash -c "dokku buildpacks:set $TEST_APP heroku/nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku buildpacks:set --index 2 $TEST_APP heroku/ruby"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku buildpacks:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists

  run /bin/bash -c "dokku buildpacks:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku buildpacks:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output_not_exists
}
