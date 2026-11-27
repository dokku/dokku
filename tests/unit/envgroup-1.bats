#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_bats_envgroups
  dokku config:unset --global --no-restart ENVGROUP_SHARED ENVGROUP_GLOBAL_ONLY >/dev/null 2>&1 || true
  destroy_app
  dokku --force apps:destroy "$TEST_APP-renamed" >/dev/null 2>&1 || true
  dokku --force apps:destroy "$TEST_APP-clone" >/dev/null 2>&1 || true
  global_teardown
}

@test "(envgroup) envgroup:help" {
  run /bin/bash -c "dokku envgroup --help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage environment variable groups shared by apps"
  help_output="$output"

  run /bin/bash -c "dokku envgroup:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage environment variable groups shared by apps"
  assert_output "$help_output"
}

@test "(envgroup:create) creates a group" {
  run /bin/bash -c "dokku envgroup:exists bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Env group bats-dev does not exist"

  run /bin/bash -c "dokku envgroup:create bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Created env group bats-dev"

  run /bin/bash -c "dokku envgroup:exists bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku envgroup:create bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Env group bats-dev already exists"

  run /bin/bash -c "dokku envgroup:create Bats_Invalid"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Env group name must begin with lowercase alphanumeric character"

  run /bin/bash -c "dokku envgroup:create"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Please specify an env group name"
}

@test "(envgroup:list) lists groups" {
  dokku envgroup:create bats-staging
  dokku envgroup:create bats-dev

  run /bin/bash -c "dokku --quiet envgroup:list | grep -c '^bats-'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "2"

  run /bin/bash -c "dokku --quiet envgroup:list | grep '^bats-' | head -n1"
  echo "output: $output"
  echo "status: $status"
  assert_output "bats-dev"

  run /bin/bash -c "dokku envgroup:list"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "=====> Env groups"
}

@test "(envgroup:destroy) destroys a group" {
  dokku envgroup:create bats-dev
  dokku envgroup:set bats-dev KEY=value

  run /bin/bash -c "dokku envgroup:destroy bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Destroyed env group bats-dev"

  run /bin/bash -c "dokku envgroup:exists bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "test -d /var/lib/dokku/data/envgroup/bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku envgroup:destroy bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Env group bats-dev does not exist"
}

@test "(envgroup:set) sets values in a group" {
  run /bin/bash -c "dokku envgroup:set bats-dev KEY=value"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Env group bats-dev does not exist"

  dokku envgroup:create bats-dev

  run /bin/bash -c "dokku envgroup:set bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "At least one env pair must be given"

  run /bin/bash -c "dokku envgroup:set bats-dev KEY=value SPACED=\"hello world\" \"QUOTES=it's\""
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Setting config vars"

  run /bin/bash -c "dokku envgroup:set --encoded bats-dev ENCODED=\"$(printf '%s' 'line one' | base64 -w 0)\""
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku envgroup:set bats-dev 1INVALID=value"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid key name: '1INVALID'"

  run /bin/bash -c "dokku envgroup:get bats-dev KEY"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "value"

  run /bin/bash -c "dokku envgroup:get bats-dev SPACED"
  echo "output: $output"
  echo "status: $status"
  assert_output "hello world"

  run /bin/bash -c "dokku envgroup:get bats-dev ENCODED"
  echo "output: $output"
  echo "status: $status"
  assert_output "line one"

  run /bin/bash -c "dokku envgroup:get --quoted bats-dev QUOTES"
  echo "output: $output"
  echo "status: $status"
  assert_output "'it'\\''s'"

  run /bin/bash -c "dokku envgroup:get bats-dev MISSING"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output ""
}

@test "(envgroup:unset) unsets values in a group" {
  dokku envgroup:create bats-dev
  dokku envgroup:set bats-dev KEY=value OTHER=value

  run /bin/bash -c "dokku envgroup:unset bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "At least one key must be given"

  run /bin/bash -c "dokku envgroup:unset bats-dev KEY MISSING"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Unsetting KEY"
  assert_output_contains "Skipping MISSING, it is not set in the environment"

  run /bin/bash -c "dokku envgroup:keys bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "OTHER"
}

@test "(envgroup:clear) clears a group" {
  dokku envgroup:create bats-dev
  dokku envgroup:set bats-dev KEY=value OTHER=value

  run /bin/bash -c "dokku envgroup:clear bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku envgroup:keys bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""
}

@test "(envgroup:export) exports a group in each format" {
  dokku envgroup:create bats-dev
  dokku envgroup:set bats-dev B_KEY=two A_KEY=one

  run /bin/bash -c "dokku envgroup:export bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(printf "export A_KEY='one'\nexport B_KEY='two'")"

  run /bin/bash -c "dokku envgroup:export --format shell bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "A_KEY='one' B_KEY='two' "

  run /bin/bash -c "dokku envgroup:export --format json bats-dev | jq -r .B_KEY"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "two"

  run /bin/bash -c "dokku envgroup:export --format docker-args-keys bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "--env=A_KEY --env=B_KEY"

  run /bin/bash -c "dokku envgroup:export --format yaml bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Unknown export format: yaml"

  run /bin/bash -c "dokku envgroup:show bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "=====> bats-dev env vars"
  assert_output_contains "A_KEY:"
  assert_output_contains "two"

  run /bin/bash -c "dokku envgroup:bundle bats-dev | tar -tf -"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$(printf "A_KEY\nB_KEY")"
}

@test "(envgroup:import) imports values into a group" {
  dokku envgroup:create bats-dev
  dokku envgroup:set bats-dev EXISTING=value

  run /bin/bash -c "printf 'FROM_STDIN=one\nOTHER=two\n' | dokku envgroup:import bats-dev -"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku envgroup:keys bats-dev | tr '\n' ' '"
  echo "output: $output"
  echo "status: $status"
  assert_output "EXISTING FROM_STDIN OTHER "

  run /bin/bash -c "echo '{\"FROM_JSON\": \"three\"}' | dokku envgroup:import --format json --replace bats-dev -"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku envgroup:keys bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_output "FROM_JSON"

  run /bin/bash -c "dokku envgroup:get bats-dev FROM_JSON"
  echo "output: $output"
  echo "status: $status"
  assert_output "three"

  run /bin/bash -c "printf '' | dokku envgroup:import bats-dev -"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "No config vars to set"
}

@test "(envgroup:add) adds apps to groups in order" {
  dokku envgroup:create bats-first
  dokku envgroup:create bats-second
  dokku envgroup:create bats-third

  run /bin/bash -c "dokku envgroup:add bats-missing $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Env group bats-missing does not exist"

  run /bin/bash -c "dokku envgroup:add bats-first $TEST_APP-missing"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "App $TEST_APP-missing does not exist"

  run /bin/bash -c "dokku envgroup:add bats-first $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Added app $TEST_APP to env group bats-first"

  run /bin/bash -c "dokku envgroup:add bats-first $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "App $TEST_APP is already a member of env group bats-first"

  run /bin/bash -c "dokku envgroup:add bats-third $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku envgroup:add --index 2 bats-second $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku envgroup:report $TEST_APP --envgroup-groups"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "bats-first bats-second bats-third"

  run /bin/bash -c "dokku envgroup:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Envgroup groups:"
  assert_output_contains "bats-first bats-second bats-third"

  run /bin/bash -c "dokku envgroup:report $TEST_APP --format json | jq -r .groups"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "bats-first bats-second bats-third"

  dokku envgroup:set bats-second KEY=value
  run /bin/bash -c "dokku envgroup:info bats-second"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "=====> bats-second env group information"
  assert_output_contains "Env group apps:  $TEST_APP"
  assert_output_contains "Env group keys:  1"
}

@test "(envgroup:remove) removes an app from a group" {
  dokku envgroup:create bats-dev
  dokku envgroup:add bats-dev "$TEST_APP"

  run /bin/bash -c "dokku envgroup:remove bats-dev $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Removed app $TEST_APP from env group bats-dev"

  run /bin/bash -c "dokku envgroup:remove bats-dev $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "App $TEST_APP is not a member of env group bats-dev"

  run /bin/bash -c "dokku envgroup:report $TEST_APP --envgroup-groups"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output ""
}

@test "(envgroup) merged config layers groups between global and app" {
  dokku config:set --global --no-restart ENVGROUP_SHARED=global ENVGROUP_GLOBAL_ONLY=global
  dokku envgroup:create bats-first
  dokku envgroup:create bats-second
  dokku envgroup:set bats-first ENVGROUP_SHARED=first FIRST_ONLY=first APP_WINS=first
  dokku envgroup:set bats-second ENVGROUP_SHARED=second SECOND_ONLY=second
  dokku config:set --no-restart "$TEST_APP" APP_WINS=app
  dokku envgroup:add bats-first "$TEST_APP"
  dokku envgroup:add bats-second "$TEST_APP"

  run /bin/bash -c "dokku config:export --merged --format json $TEST_APP | jq -r '[.ENVGROUP_SHARED, .ENVGROUP_GLOBAL_ONLY, .FIRST_ONLY, .SECOND_ONLY, .APP_WINS] | join(\" \")'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "second global first second app"

  run /bin/bash -c "dokku envgroup:remove bats-second $TEST_APP && dokku envgroup:add --index 1 bats-second $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:export --merged --format json $TEST_APP | jq -r .ENVGROUP_SHARED"
  echo "output: $output"
  echo "status: $status"
  assert_output "first"

  run /bin/bash -c "dokku config:keys --merged $TEST_APP | grep -c '_ONLY$'"
  echo "output: $output"
  echo "status: $status"
  assert_output "3"

  run /bin/bash -c "dokku config:get $TEST_APP FIRST_ONLY"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku config:export --format json $TEST_APP | jq -r 'has(\"FIRST_ONLY\")'"
  echo "output: $output"
  echo "status: $status"
  assert_output "false"
}

@test "(envgroup:destroy) requires --force while apps are members" {
  dokku envgroup:create bats-dev
  dokku envgroup:create bats-other
  dokku envgroup:set bats-dev KEY=value
  dokku envgroup:add bats-dev "$TEST_APP"
  dokku envgroup:add bats-other "$TEST_APP"

  run /bin/bash -c "dokku envgroup:destroy bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Env group bats-dev is in use by the following apps: $TEST_APP. Pass --force to remove it from them and destroy it"

  run /bin/bash -c "dokku envgroup:exists bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku envgroup:destroy --force bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Removed app $TEST_APP from env group bats-dev"
  assert_output_contains "Destroyed env group bats-dev"

  run /bin/bash -c "dokku envgroup:report $TEST_APP --envgroup-groups"
  echo "output: $output"
  echo "status: $status"
  assert_output "bats-other"

  run /bin/bash -c "dokku config:export --merged --format json $TEST_APP | jq -r 'has(\"KEY\")'"
  echo "output: $output"
  echo "status: $status"
  assert_output "false"
}

@test "(envgroup) memberships follow app renames and clones and are removed on destroy" {
  dokku envgroup:create bats-dev
  dokku envgroup:set bats-dev KEY=value
  dokku envgroup:add bats-dev "$TEST_APP"

  run /bin/bash -c "dokku apps:clone $TEST_APP $TEST_APP-clone"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku envgroup:report $TEST_APP-clone --envgroup-groups"
  echo "output: $output"
  echo "status: $status"
  assert_output "bats-dev"

  run /bin/bash -c "dokku apps:rename $TEST_APP $TEST_APP-renamed"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku envgroup:report $TEST_APP-renamed --envgroup-groups"
  echo "output: $output"
  echo "status: $status"
  assert_output "bats-dev"

  run /bin/bash -c "dokku config:export --merged --format json $TEST_APP-renamed | jq -r .KEY"
  echo "output: $output"
  echo "status: $status"
  assert_output "value"

  run /bin/bash -c "test -e /var/lib/dokku/config/envgroup/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP-clone"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku envgroup:info bats-dev"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Env group apps:  $TEST_APP-renamed"

  run /bin/bash -c "test -e /var/lib/dokku/config/envgroup/$TEST_APP-clone"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  dokku apps:rename "$TEST_APP-renamed" "$TEST_APP"
}

destroy_bats_envgroups() {
  declare desc="destroys every env group created by these tests"
  local group
  for group in $(dokku --quiet envgroup:list 2>/dev/null | grep '^bats-'); do
    dokku envgroup:destroy --force "$group" >/dev/null 2>&1 || true
  done
}
