GOARCH ?= amd64
SUBCOMMANDS = subcommands/default subcommands/annotations:set subcommands/annotations:report subcommands/create subcommands/destroy subcommands/ensure-directory subcommands/exec subcommands/info subcommands/labels:set subcommands/labels:report subcommands/list subcommands/list-entries subcommands/migrate subcommands/mount subcommands/report subcommands/set subcommands/unmount subcommands/wait
TRIGGERS = triggers/install triggers/storage-list triggers/storage-app-mounts triggers/docker-args-deploy triggers/docker-args-run triggers/post-delete triggers/post-app-clone-setup triggers/post-app-rename-setup
BUILD = commands subcommands triggers
PLUGIN_NAME = storage

include ../../common.mk
