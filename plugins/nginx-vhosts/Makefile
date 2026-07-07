GOARCH ?= amd64
TRIGGERS = triggers/report
BUILD = pagesize nginx-property report-subcommand triggers
PLUGIN_NAME = nginx-vhosts

clean-report-subcommand:
	rm -rf report-subcommand

report-subcommand: clean-report-subcommand src/subcommands/subcommands.go
	GOARCH=$(GOARCH) go build -mod=readonly -ldflags="-s -w" $(GO_ARGS) -o report-subcommand src/subcommands/subcommands.go
	ln -sf ../report-subcommand subcommands/report

clean-pagesize:
	rm -rf pagesize

pagesize: clean-pagesize **/**/pagesize.go
	GOARCH=$(GOARCH) go build -ldflags="-s -w" $(GO_ARGS) -o pagesize src/pagesize/pagesize.go

clean-nginx-property:
	rm -rf nginx-property

nginx-property: clean-nginx-property **/**/nginx-property.go
	GOARCH=$(GOARCH) go build -ldflags="-s -w" $(GO_ARGS) -o nginx-property src/nginx-property/nginx-property.go

include ../../common.mk
