BUILD = prop
PLUGIN_NAME = common

clean-prop:
	rm -rf prop

prop: clean-prop **/**/prop.go
	go build $(GO_ARGS) -o prop src/prop/prop.go

include ../../common.mk
