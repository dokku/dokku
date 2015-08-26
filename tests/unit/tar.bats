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
  GIT_REMOTE="$2"; GIT_REMOTE=${GIT_REMOTE:="dokku@dokku.me:$TEST_APP"}
  TMP=$(mktemp -d -t "dokku.me.XXXXX")
  rmdir $TMP && cp -r ./tests/apps/$APP_TYPE $TMP
  cd $TMP
  tar c . | ssh dokku@dokku.me tar:in $TEST_APP || destroy_app $?
}

@test "(tar) basic deploy using tar:in" {
  deploy_app_tar

  run bash -c "response=\"$(curl -s -S my-cool-guy-test-app.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success  
}

