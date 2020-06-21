#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  global_teardown
}

@test "(buildpacks) buildpacks:help" {
  run /bin/bash -c "dokku buildpacks"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage buildpacks settings for an app"
  help_output="$output"

  run /bin/bash -c "dokku buildpacks:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage buildpacks settings for an app"
  assert_output "$help_output"
}

@test "(buildpacks) buildpacks:add - failure" {
  run /bin/bash -c "dokku buildpacks:add $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku buildpacks:add $TEST_APP nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku buildpacks:add $TEST_APP /nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku buildpacks:add $TEST_APP /nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku buildpacks:add $TEST_APP http://nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(buildpacks) buildpacks:add - success" {
  run /bin/bash -c "dokku buildpacks:add $TEST_APP heroku/nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "https://github.com/heroku/heroku-buildpack-nodejs.git"

  run /bin/bash -c "dokku buildpacks:add $TEST_APP heroku/ruby"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "https://github.com/heroku/heroku-buildpack-nodejs.git https://github.com/heroku/heroku-buildpack-ruby.git"

  run /bin/bash -c "dokku buildpacks:add --index 1 $TEST_APP heroku/golang"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "https://github.com/heroku/heroku-buildpack-golang.git https://github.com/heroku/heroku-buildpack-nodejs.git https://github.com/heroku/heroku-buildpack-ruby.git"

  run /bin/bash -c "dokku buildpacks:add --index 2 $TEST_APP heroku/python"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "https://github.com/heroku/heroku-buildpack-golang.git https://github.com/heroku/heroku-buildpack-python.git https://github.com/heroku/heroku-buildpack-nodejs.git https://github.com/heroku/heroku-buildpack-ruby.git"

  run /bin/bash -c "dokku buildpacks:add --index 100 $TEST_APP heroku/php"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "https://github.com/heroku/heroku-buildpack-golang.git https://github.com/heroku/heroku-buildpack-python.git https://github.com/heroku/heroku-buildpack-nodejs.git https://github.com/heroku/heroku-buildpack-ruby.git https://github.com/heroku/heroku-buildpack-php.git"
}

@test "(buildpacks) buildpacks:set - failure" {
  run /bin/bash -c "dokku buildpacks:set $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku buildpacks:set $TEST_APP nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku buildpacks:set $TEST_APP /nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku buildpacks:set $TEST_APP /nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku buildpacks:set $TEST_APP http://nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(buildpacks) buildpacks:set - success" {
  run /bin/bash -c "dokku buildpacks:set $TEST_APP heroku/nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "https://github.com/heroku/heroku-buildpack-nodejs.git"

  run /bin/bash -c "dokku buildpacks:set $TEST_APP heroku/ruby"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "https://github.com/heroku/heroku-buildpack-ruby.git"

  run /bin/bash -c "dokku buildpacks:set --index 1 $TEST_APP heroku/golang"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "https://github.com/heroku/heroku-buildpack-golang.git"

  run /bin/bash -c "dokku buildpacks:set --index 2 $TEST_APP heroku/python"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "https://github.com/heroku/heroku-buildpack-golang.git https://github.com/heroku/heroku-buildpack-python.git"

  run /bin/bash -c "dokku buildpacks:set --index 100 $TEST_APP heroku/php"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output "https://github.com/heroku/heroku-buildpack-golang.git https://github.com/heroku/heroku-buildpack-python.git https://github.com/heroku/heroku-buildpack-php.git"
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
  assert_output_contains "https://github.com/heroku/heroku-buildpack-nodejs.git https://github.com/heroku/heroku-buildpack-ruby.git"

  run /bin/bash -c "dokku buildpacks:remove $TEST_APP heroku/nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet buildpacks:list $TEST_APP | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "https://github.com/heroku/heroku-buildpack-ruby.git"

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

@test "(buildpacks) buildpacks deploy" {
  create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku buildpacks:set $TEST_APP https://github.com/dokku/buildpack-null"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku buildpacks:set $TEST_APP heroku/nodejs"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
}
