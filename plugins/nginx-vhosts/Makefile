BUILD = pagesize
PLUGIN_NAME = nginx-vhosts

clean-pagesize:
	rm -rf pagesize

pagesize: clean-pagesize **/**/pagesize.go
	go build -ldflags="-s -w" $(GO_ARGS) -o pagesize src/pagesize/pagesize.go

include ../../common.mk
