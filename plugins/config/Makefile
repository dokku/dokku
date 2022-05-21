SUBCOMMANDS = subcommands/bundle subcommands/clear subcommands/export subcommands/get subcommands/keys subcommands/show subcommands/set subcommands/unset
TRIGGERS = triggers/config-export triggers/config-get triggers/config-get-global
BUILD = commands config_sub subcommands triggers
PLUGIN_NAME = config

clean-config_sub:
	rm -rf config_sub

config_sub: clean-config_sub **/**/config_sub.go
	go build -ldflags="-s -w" $(GO_ARGS) -o config_sub src/config_sub/config_sub.go

include ../../common.mk
