GOARCH ?= amd64
BUILD = plugin-list
PLUGIN_NAME = plugin

clean-plugin-list:
	rm -rf plugin-list

plugin-list: clean-plugin-list **/**/plugin-list.go
	GOARCH=$(GOARCH) go build -ldflags="-s -w" $(GO_ARGS) -o plugin-list src/plugin-list/plugin-list.go

include ../../common.mk
