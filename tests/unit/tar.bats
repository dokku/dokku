#!/usr/bin/env bats

load test_helper

setup() {
  create_app
}

teardown() {
  destroy_app
}

deploy_app_tar() {
  APP_TYPE="$1"; APP_TYPE=${APP_TYPE:="nodejs-express"}
  TMP=$(mktemp -d -t "dokku.me.XXXXX")
  rmdir $TMP && cp -r ./tests/apps/$APP_TYPE $TMP
  cd $TMP
  shift 1
  tar c . $* | ssh dokku@dokku.me tar:in $TEST_APP || destroy_app $?
  sleep 5 # nginx needs some time to itself...
}

@test "(tar) non-tarbomb deploy using tar:in" {
  deploy_app_tar nodejs-express --transform 's,^,prefix/,'

  run bash -c "response=\"$(curl -s -S my-cool-guy-test-app.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "(tar) tarbomb deploy using tar:in" {
  deploy_app_tar nodejs-express

  run bash -c "response=\"$(curl -s -S my-cool-guy-test-app.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success
}
