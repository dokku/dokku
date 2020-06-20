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

@test "(ps) ps:help" {
  run /bin/bash -c "dokku ps"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage container processes"
  help_output="$output"

  run /bin/bash -c "dokku ps:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage container processes"
  assert_output "$help_output"
}

@test "(ps) ps:inspect" {
  dokku config:set "$TEST_APP" key=value key=value=value
  deploy_app dockerfile

  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "dokku ps:inspect $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$CID" 6
}

@test "(ps:scale) procfile commands extraction" {
  source "$PLUGIN_CORE_AVAILABLE_PATH/ps/functions"
  cat <<EOF > "$DOKKU_ROOT/$TEST_APP/DOKKU_PROCFILE"
web: node web.js --port \$PORT
worker: node worker.js
EOF
  run get_cmd_from_procfile "$TEST_APP" web 5001
  echo "output: $output"
  echo "status: $status"
  assert_output "node web.js --port 5001"

  run get_cmd_from_procfile "$TEST_APP" worker
  echo "output: $output"
  echo "status: $status"
  assert_output "node worker.js"
}

@test "(ps:scale) update DOKKU_SCALE from Procfile" {
  local TMP=$(mktemp -d "/tmp/dokku.me.XXXXX")
  trap 'popd &>/dev/null || true; rm -rf "$TMP"' INT TERM

  CUSTOM_TMP="$TMP" deploy_app nodejs-express

  run /bin/bash -c "dokku --quiet ps:scale $TEST_APP"
  output=$(echo "$output" | tr -s " ")
  echo "output: ($output)"
  assert_output $'cron: 0 \ncustom: 0 \nrelease: 0 \nweb: 1 \nworker: 0 '

  pushd $TMP
  echo scaletest: sleep infinity >> Procfile
  git commit Procfile -m 'Add scaletest process'
  git push target master:master

  run /bin/bash -c "dokku --quiet ps:scale $TEST_APP"
  output=$(echo "$output" | tr -s " ")
  echo "output: ($output)"
  assert_output $'cron: 0 \ncustom: 0 \nrelease: 0 \nweb: 1 \nworker: 0 \nscaletest: 0 '

  popd
  rm -rf "$TMP"
}

@test "(ps:restart-policy) default policy" {
  run /bin/bash -c "dokku --quiet ps:restart-policy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output "on-failure:10"
}


@test "(ps:restart-policy) ps:set-restart-policy, ps:restart-policy" {
  for policy in no unless-stopped always on-failure on-failure:20; do
    run /bin/bash -c "dokku ps:set-restart-policy $TEST_APP $policy"
    echo "output: $output"
    echo "status: $status"
    assert_success

    run /bin/bash -c "dokku --quiet ps:restart-policy $TEST_APP"
    echo "output: $output"
    echo "status: $status"
    assert_output "$policy"
  done
}

@test "(ps:restart-policy) deployed policy" {
  test_restart_policy="on-failure:20"
  run /bin/bash -c "dokku ps:set-restart-policy $TEST_APP $test_restart_policy"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet ps:restart-policy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_output "$test_restart_policy"

  deploy_app dockerfile

  CID=$(< $DOKKU_ROOT/$TEST_APP/CONTAINER.web.1)
  run /bin/bash -c "docker inspect -f '{{ .HostConfig.RestartPolicy.Name }}:{{ .HostConfig.RestartPolicy.MaximumRetryCount }}' $CID"
  echo "output: $output"
  echo "status: $status"
  assert_output "$test_restart_policy"
}
