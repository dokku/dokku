SUBCOMMANDS = subcommands/export subcommands/import
TRIGGERS = triggers/install
BUILD = commands subcommands triggers backup-property
PLUGIN_NAME = backup

include ../../common.mk

backup-property:
	GOARCH=$(GOARCH) go build -mod=readonly -ldflags="-s -w" $(GO_ARGS) -o backup-property src/backup-property/backup-property.go
