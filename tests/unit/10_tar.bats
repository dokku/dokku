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

deploy_app_tar() {
  APP_TYPE="$1"; APP_TYPE=${APP_TYPE:="nodejs-express"}
  TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")

  rmdir "$TMP" && cp -r "${BATS_TEST_DIRNAME}/../../tests/apps/$APP_TYPE" "$TMP"
  pushd "$TMP" &>/dev/null || exit 1
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' RETURN INT TERM

  shift 1
  tar c . $* | ssh dokku@dokku.me tar:in $TEST_APP || destroy_app $?
  sleep 5 # nginx needs some time to itself...
}

@test "(tar) non-tarbomb deploy using tar:in" {
  deploy_app_tar nodejs-express --transform 's,^,prefix/,'

  run /bin/bash -c "response=\"$(curl -s -S ${TEST_APP}.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(tar) tarbomb deploy using tar:in" {
  deploy_app_tar nodejs-express

  run /bin/bash -c "response=\"$(curl -s -S ${TEST_APP}.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: $output"
  echo "status: $status"
  assert_success
}
