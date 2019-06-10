set -eo pipefail
[[ $DOKKU_TRACE ]] && set -x
DOKKU_CONTAINER_NAME=${DOKKU_CONTAINER_NAME:=dokku}

docker exec -ti $DOKKU_CONTAINER_NAME dokku "$@"
