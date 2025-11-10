GOARCH ?= amd64
SUBCOMMANDS = subcommands/bundle subcommands/clear subcommands/export subcommands/get subcommands/import subcommands/keys subcommands/show subcommands/set subcommands/unset
TRIGGERS = triggers/config-export triggers/config-get triggers/config-get-global triggers/config-unset triggers/post-app-clone-setup triggers/post-app-rename-setup
BUILD = commands config_sub subcommands triggers
PLUGIN_NAME = config

clean-config_sub:
	rm -rf config_sub

config_sub: clean-config_sub **/**/config_sub.go
	GOARCH=$(GOARCH) go build -ldflags="-s -w" $(GO_ARGS) -o config_sub src/config_sub/config_sub.go

include ../../common.mk
