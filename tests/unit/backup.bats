#!/usr/bin/env bats

# load test_helper

# setup() {
#   create_app
#   backup_file=$(mktemp /tmp/backup.XXXXXX)
#   rm $backup_file
# }

# teardown() {
#   destroy_app
# }

# @test "(backup) backup:export" {
#   run dokku backup:export $backup_file
#   echo "output: "$output
#   echo "status: "$status
#   assert_success
# }
