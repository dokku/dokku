#!/usr/bin/env bats

load test_helper

setup() {
  REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
  HOST_ARCH="$(dpkg --print-architecture)"
  OTHER_ARCH="amd64"
  if [[ "$HOST_ARCH" == "amd64" ]]; then
    OTHER_ARCH="arm64"
  fi
}

@test "(makefile) [arch] dependency urls default to the host architecture" {
  run make -n -C "$REPO_ROOT" docker-image-labeler
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$(dependency_url docker-image-labeler "$HOST_ARCH")" -1
}

@test "(makefile) [arch] TARGETARCH overrides the dependency architecture" {
  run env TARGETARCH="$OTHER_ARCH" make -n -C "$REPO_ROOT" docker-image-labeler
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$(dependency_url docker-image-labeler "$OTHER_ARCH")" -1
}

@test "(makefile) [arch] plugin builds default to the host architecture" {
  run make -B -n -C "$REPO_ROOT" go-build-plugin PLUGIN_NAME=common PLUGIN_MAKE_TARGET=build
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "GOARCH=$HOST_ARCH go build" -1
  assert_output_contains "GOARCH=$OTHER_ARCH" 0
}

@test "(makefile) [arch] TARGETARCH sets the plugin build architecture" {
  run env TARGETARCH="$OTHER_ARCH" make -B -n -C "$REPO_ROOT" go-build-plugin PLUGIN_NAME=common PLUGIN_MAKE_TARGET=build
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "GOARCH=$OTHER_ARCH go build" -1
  assert_output_contains "GOARCH=$HOST_ARCH" 0
}

@test "(makefile) [arch] GOARCH overrides the plugin build architecture" {
  run env GOARCH="$OTHER_ARCH" make -B -n -C "$REPO_ROOT" go-build-plugin PLUGIN_NAME=common PLUGIN_MAKE_TARGET=build
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "GOARCH=$OTHER_ARCH go build" -1
  assert_output_contains "GOARCH=$HOST_ARCH" 0

  run env GOARCH="$OTHER_ARCH" make -n -C "$REPO_ROOT" docker-image-labeler
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$(dependency_url docker-image-labeler "$HOST_ARCH")" -1
}

dependency_url() {
  declare desc="print the url for a dependency in contrib/dependencies.json"
  declare NAME="$1" ARCH="$2"
  jq -r --arg name "$NAME" --arg arch "$ARCH" '.dependencies[] | select(.name == $name) | .urls[$arch]' "$REPO_ROOT/contrib/dependencies.json"
}
