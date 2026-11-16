#!/usr/bin/env bash
set -euo pipefail

# Build only the selected Git revision, preserving executable bits.
repo=$(git -C "$(dirname "$0")/../../.." rev-parse --show-toplevel)
revision=$(git -C "$repo" rev-parse "${DOKKU_SOURCE_REVISION:-HEAD}")
artifacts=${1:?Usage: build-dokku.sh PRIVATE_ARTIFACT_DIRECTORY}
stamp=$(date -u +%Y%m%dT%H%M%SZ)
context="$artifacts/evidence/source-$stamp-${revision:0:9}"
mkdir -p "$context/build/package"
exec > >(tee "$artifacts/evidence/source-build-$stamp.log") 2>&1
git -C "$repo" archive --format=tar --output="$context/source.tar" "$revision"
go_version=$(git -C "$repo" show "$revision:common.mk" | awk '/^BUILD_IMAGE / {sub(/^golang:/,"",$3); print $3}')
version="$(git -C "$repo" show "$revision:debian/control" | awk '/^Version:/ {print $2}')build+${revision:0:9}"
builder="dokku-e2e/package:${revision:0:9}"
image="dokku-e2e/source:${revision:0:9}"
started=$SECONDS

docker build --progress plain --target amd64 -f contrib/build-dokku.Dockerfile \
  --build-arg GOLANG_VERSION="$go_version" --build-arg PLUGIN_MAKE_TARGET=build \
  --build-arg DOKKU_VERSION="$version" --build-arg DOKKU_GIT_REV="$revision" \
  --build-arg IS_RELEASE=false -t "$builder" - <"$context/source.tar"
container=$(docker create "$builder")
trap 'docker rm "$container" >/dev/null' EXIT
docker cp "$container:/tmp/dokku_${version}_amd64.deb" "$context/build/package/dokku-amd64.deb"
cp "$context/source.tar" "$context/runtime.tar"
tar --append --file="$context/runtime.tar" -C "$context" build/package/dokku-amd64.deb
docker build --progress plain --label "org.opencontainers.image.revision=$revision" \
  --build-arg DOKKU_TAG="$version" -t "$image" - <"$context/runtime.tar"
docker image inspect "$image" --format '{{json .}}' >"$artifacts/evidence/source-image-$stamp.json"
sha256sum "$context/build/package/dokku-amd64.deb" >"$artifacts/evidence/source-package-$stamp.sha256"
printf '{"elapsedSeconds":%d,"backend":"docker-default","revision":"%s"}\n' "$((SECONDS - started))" "$revision" >"$artifacts/evidence/build-timing-$stamp.json"
printf 'SOURCE_REVISION=%s\nSOURCE_IMAGE=%s\nBUILD_CONTEXT=%s\n' "$revision" "$image" "$context"
