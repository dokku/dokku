#!/bin/bash
set -eo pipefail
set -o errexit

echo '--> Enable Google DNS for Docker'
sed -e 's|#DOCKER_OPTS|DOCKER_OPTS|g' \
  -i /etc/default/docker
