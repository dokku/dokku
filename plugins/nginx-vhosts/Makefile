GOARCH ?= amd64
BUILD = pagesize nginx-property
PLUGIN_NAME = nginx-vhosts

clean-pagesize:
	rm -rf pagesize

pagesize: clean-pagesize **/**/pagesize.go
	GOARCH=$(GOARCH) go build -ldflags="-s -w" $(GO_ARGS) -o pagesize src/pagesize/pagesize.go

clean-nginx-property:
	rm -rf nginx-property

nginx-property: clean-nginx-property **/**/nginx-property.go
	GOARCH=$(GOARCH) go build -ldflags="-s -w" $(GO_ARGS) -o nginx-property src/nginx-property/nginx-property.go

include ../../common.mk
