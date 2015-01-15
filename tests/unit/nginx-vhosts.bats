#!/usr/bin/env bats

load test_helper

setup() {
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -f "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  [[ -f "$DOKKU_ROOT/HOSTNAME" ]] && cp -f "$DOKKU_ROOT/HOSTNAME" "$DOKKU_ROOT/HOSTNAME.bak"
  create_app
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST"
  [[ -f "$DOKKU_ROOT/HOSTNAME.bak" ]] && mv "$DOKKU_ROOT/HOSTNAME.bak" "$DOKKU_ROOT/HOSTNAME"
}

@test "nginx (no server tokens)" {
  deploy_app
  run /bin/bash -c "curl -s -D - $(dokku url $TEST_APP) -o /dev/null | egrep '^Server' | egrep '[0-9]+'"
  echo "output: "$output
  echo "status: "$status
  assert_failure
}

@test "nginx:build-config (with SSL CN mismatch)" {
  setup_test_tls
  deploy_app
  run /bin/bash -c "dokku domains $TEST_APP | egrep ^node-js-app\.dokku\.me$"
  echo "output: "$output
  echo "status: "$status
  assert_output "node-js-app.dokku.me"
  run bash -c "response=\"$(curl -LkSs node-js-app.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "nginx:build-config (with SSL and Multiple SANs)" {
  setup_test_tls_with_sans
  deploy_app
  run /bin/bash -c "dokku domains $TEST_APP | egrep ^test\.dokku\.me$"
  echo "output: "$output
  echo "status: "$status
  assert_output "test.dokku.me"
  run /bin/bash -c "dokku domains $TEST_APP | grep ^www\.test\.dokku\.me$"
  echo "output: "$output
  echo "status: "$status
  assert_output "www.test.dokku.me"
  run bash -c "response=\"$(curl -LkSs test.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "response=\"$(curl -LkSs www.test.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success
  run bash -c "response=\"$(curl -LkSs www.test.app.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success
}

@test "nginx:build-config (no global VHOST and domains:add)" {
  destroy_app
  rm "$DOKKU_ROOT/VHOST"
  create_app
  run dokku domains:add $TEST_APP www.test.app.dokku.me
  echo "output: "$output
  echo "status: "$status
  assert_success
  deploy_app
  sleep 5 # wait for nginx to reload
  run bash -c "response=\"$(curl -s -S www.test.app.dokku.me)\"; echo \$response; test \"\$response\" == \"nodejs/express\""
  echo "output: "$output
  echo "status: "$status
  assert_success
}
