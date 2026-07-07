GOARCH ?= amd64
TRIGGERS = triggers/report
BUILD = report-subcommand triggers
PLUGIN_NAME = builder-dockerfile

clean-report-subcommand:
	rm -rf report-subcommand

report-subcommand: clean-report-subcommand src/subcommands/subcommands.go
	GOARCH=$(GOARCH) go build -mod=readonly -ldflags="-s -w" $(GO_ARGS) -o report-subcommand src/subcommands/subcommands.go
	ln -sf ../report-subcommand subcommands/report

include ../../common.mk
