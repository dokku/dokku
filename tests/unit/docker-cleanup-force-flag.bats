#!/usr/bin/env bats

setup() {
  DOKKU_TEST_TMPDIR="$(mktemp -d)"
  export DOKKU_TEST_ARGS_FILE="$DOKKU_TEST_TMPDIR/args"
  export PLUGIN_CORE_AVAILABLE_PATH="$DOKKU_TEST_TMPDIR/plugin"

  mkdir -p "$PLUGIN_CORE_AVAILABLE_PATH/common"
  cat >"$PLUGIN_CORE_AVAILABLE_PATH/common/common" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$@" > "$DOKKU_TEST_ARGS_FILE"
EOF
  chmod +x "$PLUGIN_CORE_AVAILABLE_PATH/common/common"

  source "$BATS_TEST_DIRNAME/../../plugins/common/functions"
}

teardown() {
  rm -rf "$DOKKU_TEST_TMPDIR"
}

@test "docker_cleanup forwards the manual force flag to the core command" {
  run docker_cleanup --global true

  [ "$status" -eq 0 ]
  [ "$(cat "$DOKKU_TEST_ARGS_FILE")" = $'--quiet\ndocker-cleanup\n--global\ntrue' ]
}
