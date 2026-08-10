#!/usr/bin/env bats

load test_helper

setup() {
  rm "${BATS_PARENT_TMPNAME}.skip" || true
  global_setup
}

teardown() {
  destroy_app
  # a leftover app-label-alias makes every later test's vector source filter on
  # a label that dokku never applies to a container, which silently disables log
  # collection for the rest of the file
  dokku logs:set --global app-label-alias >/dev/null 2>/dev/null || true
  dokku logs:set --global vector-image >/dev/null 2>/dev/null || true
  dokku logs:set --global vector-networks >/dev/null 2>/dev/null || true
  dokku logs:set --global vector-sink >/dev/null 2>/dev/null || true
  dokku logs:set --global vector-cron-sink >/dev/null 2>/dev/null || true
  docker network rm test-vector-net-a >/dev/null || true
  docker network rm test-vector-net-b >/dev/null || true
  global_teardown
}

@test "(logs) vector validates the generated cron routing config" {
  run create_app
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink console://?encoding[codec]=json"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-cron-sink console://?encoding[codec]=text"
  assert_success

  run /bin/bash -c "dokku logs:vector-start 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # unit tests cannot catch a VRL syntax error or a malformed route schema
  run /bin/bash -c "docker exec vector-vector-1 vector validate --no-environment /etc/vector/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # the relabel branches are generated VRL too, and a syntax error there would
  # take the whole config down rather than just the rename
  run /bin/bash -c "dokku logs:set --global vector-sink console://?encoding[codec]=json"
  assert_success

  run /bin/bash -c "dokku logs:set --global app-label-alias global_alt_name"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP app-label-alias app_alt_name"
  assert_success

  run /bin/bash -c "docker exec vector-vector-1 vector validate --no-environment /etc/vector/vector.json"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-stop 2>&1"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP app-label-alias"
  assert_success

  run /bin/bash -c "dokku logs:set --global app-label-alias"
  assert_success

  run /bin/bash -c "dokku logs:set --global vector-sink"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-cron-sink"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink"
  assert_success
}

# the regression test for the alias silently disabling collection: the source
# has to keep filtering on the label dokku applies, while the event that comes
# out the other end carries the alias instead
@test "(logs) a non-default app-label-alias still ships logs" {
  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink 'console://?encoding[codec]=json'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set --global app-label-alias alt_name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-start 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run start_vector_probe VECTOR_ALIAS_OK
  echo "output: $output"
  echo "status: $status"
  assert_success

  run wait_for_vector_alias_event VECTOR_ALIAS_OK alt_name
  echo "output: $output"
  echo "status: $status"
  dump_vector_diagnostics "$TEST_APP"
  assert_success

  run /bin/bash -c "docker container rm --force vector-alias-probe"
  assert_success

  run /bin/bash -c "dokku logs:vector-stop 2>&1"
  assert_success
}

@test "(logs) logs:vector-start attaches configured networks" {
  docker network create test-vector-net-a >/dev/null
  docker network create test-vector-net-b >/dev/null

  run /bin/bash -c "dokku logs:set --global vector-networks test-vector-net-a,test-vector-net-b 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-start 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container is running"

  run /bin/bash -c "sudo docker inspect --format='{{range \$k, \$v := .NetworkSettings.Networks}}{{\$k}} {{end}}' vector-vector-1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "test-vector-net-a"
  assert_output_contains "test-vector-net-b"
  assert_output_contains "bridge" 0

  run /bin/bash -c "dokku logs:vector-stop 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-start 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container is running"

  run /bin/bash -c "sudo docker inspect --format='{{range \$k, \$v := .NetworkSettings.Networks}}{{\$k}} {{end}}' vector-vector-1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "test-vector-net-a"
  assert_output_contains "test-vector-net-b"
  assert_output_contains "bridge" 0

  run /bin/bash -c "dokku logs:set --global vector-networks 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-stop 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-start 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "sudo docker inspect --format='{{range \$k, \$v := .NetworkSettings.Networks}}{{\$k}} {{end}}' vector-vector-1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "bridge"

  run /bin/bash -c "dokku logs:vector-stop 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(logs) logs:vector" {
  run /bin/bash -c "dokku logs:vector-logs 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Vector container does not exist"

  run /bin/bash -c "dokku apps:create example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-start 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container is running"

  run /bin/bash -c "sudo docker inspect --format='{{.HostConfig.RestartPolicy.Name}}' vector-vector-1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "unless-stopped"

  run /bin/bash -c "dokku logs:vector-logs 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container logs"

  run /bin/bash -c "dokku --force apps:destroy example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-logs --num 10 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container logs"
  assert_output_contains "vector:" 10
  assert_line_count 11

  run /bin/bash -c "dokku logs:vector-logs --num 5 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container logs"
  assert_output_contains "vector:" 5
  assert_line_count 6

  run /bin/bash -c "docker stop vector-vector-1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:vector-logs 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Vector container logs"
  assert_output_contains "Vector container is not running"

  run /bin/bash -c "dokku logs:vector-stop 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Stopping and removing vector container"
}

@test "(logs) vector-cron-sink routes cron task output to the cron sink" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP template_cron_file_marker
  echo "output: $output"
  echo "status: $status"
  assert_success

  cron_id="$(dokku cron:list $TEST_APP --format json | jq -r '.[0].id')"
  echo "cron_id: $cron_id"

  run /bin/bash -c "dokku logs:vector-start 2>&1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # both sinks write to vector's own stdout, with different codecs, so the sink
  # an event was routed to is identifiable without depending on a writable
  # mount: the cron branch emits json carrying the fields the remap adds, while
  # anything reaching the plain sink emits the bare message
  run /bin/bash -c "dokku logs:set $TEST_APP vector-sink 'console://?encoding[codec]=text'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku logs:set $TEST_APP vector-cron-sink 'console://?encoding[codec]=json'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # vector reloads via --watch-config, so wait for the routing to be live
  run wait_for_vector_route "$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku cron:run $TEST_APP $cron_id"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run wait_for_vector_cron_event VECTOR_CRON_OK "$cron_id"
  echo "output: $output"
  echo "status: $status"
  dump_vector_diagnostics "$TEST_APP"
  assert_success

  # a bare message line would mean the event reached the plain sink instead of
  # being routed onto the cron branch
  run count_vector_plain_lines VECTOR_CRON_OK
  echo "output: $output"
  echo "status: $status"
  assert_output "0"

  run /bin/bash -c "dokku logs:vector-stop 2>&1"
  assert_success
}

wait_for_vector_route() {
  declare desc="waits for the vector config on disk to contain the app's cron router"
  declare APP="$1"
  local i

  for i in $(seq 1 30); do
    if jq -e ".transforms[\"docker-router:$APP\"]" /var/lib/dokku/data/logs/vector.json >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  echo "timed out waiting for the cron router in vector.json"
  return 1
}

start_vector_probe() {
  declare desc="runs a container carrying only the label dokku applies, emitting a marker for long enough for vector to attach"
  declare MARKER="$1"

  docker container run --detach --name vector-alias-probe \
    --label "com.dokku.app-name=$TEST_APP" \
    gliderlabs/herokuish \
    bash -c "for i in \$(seq 1 30); do echo $MARKER; sleep 1; done"
}

wait_for_vector_alias_event() {
  declare desc="waits for a marker to arrive carrying the configured alias in place of the default label"
  declare MARKER="$1" ALIAS="$2"
  local i

  for i in $(seq 1 60); do
    if docker logs vector-vector-1 2>/dev/null | grep "$MARKER" \
      | jq -e --arg alias "$ALIAS" \
        'select(.label[$alias] != null and .label["com.dokku.app-name"] == null)' >/dev/null 2>/dev/null; then
      return 0
    fi
    sleep 1
  done

  echo "timed out waiting for a $MARKER event labelled $ALIAS"
  return 1
}

wait_for_vector_cron_event() {
  declare desc="waits for a marker to arrive on the cron branch, carrying the fields the remap adds"
  declare MARKER="$1" CRON_ID="$2"
  local i

  for i in $(seq 1 60); do
    if docker logs vector-vector-1 2>/dev/null | grep "$MARKER" \
      | jq -e --arg id "$CRON_ID" 'select(.dokku_cron_id == $id)' >/dev/null 2>/dev/null; then
      return 0
    fi
    sleep 1
  done

  echo "timed out waiting for a cron-routed $MARKER event with dokku_cron_id=$CRON_ID"
  return 1
}

count_vector_plain_lines() {
  declare desc="counts bare message lines, which only the text-encoded plain sink emits"
  declare MARKER="$1"

  docker logs vector-vector-1 2>/dev/null | grep -c -x "$MARKER" || true
}

dump_vector_diagnostics() {
  declare desc="prints the state needed to tell apart the ways cron shipping can fail"
  declare APP="$1"

  # sources matter as much as routing here: a stale app-label-alias makes the
  # source filter on a label no container carries, which collects nothing
  echo "--- generated config ---"
  jq '{sources, transforms, sinks}' /var/lib/dokku/data/logs/vector.json || true

  echo "--- containers carrying the app label ---"
  docker ps --all --filter "label=com.dokku.app-name=$APP" --format '{{.Names}} {{.Status}}' || true

  echo "--- vector stdout, where both sinks write ---"
  docker logs vector-vector-1 2>/dev/null | tail -20 || true

  # vector reports template_failed and other component errors on stderr, so
  # keep it separate from the collected log lines on stdout
  echo "--- vector stderr ---"
  docker logs vector-vector-1 2>&1 1>/dev/null | tail -20 || true
}

template_cron_file_marker() {
  declare desc="writes an app.json with a cron task emitting a known marker"
  local APP="$1" APP_REPO_DIR="$2"
  [[ -z "$APP" ]] && local APP="$TEST_APP"
  echo "injecting cron app.json -> $APP_REPO_DIR/app.json"
  # cron_task.py sleeps either side of its output. a task that exits
  # immediately is removed before vector can attach to it, which is documented
  # behavior rather than something this test should assert against
  cat <<EOF >"$APP_REPO_DIR/app.json"
{
  "cron": [
    {
      "command": "python3 cron_task.py VECTOR_CRON_OK",
      "schedule": "@daily"
    }
  ]
}
EOF
}
