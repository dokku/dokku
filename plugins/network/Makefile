include ../../common.mk

GO_ARGS ?= -a

SUBCOMMANDS = subcommands/rebuild subcommands/rebuildall subcommands/report subcommands/set
TRIGGERS = triggers/install triggers/network-build-config triggers/network-compute-ports triggers/network-config-exists triggers/network-get-ipaddr triggers/network-get-listeners triggers/network-get-port triggers/network-get-property triggers/network-write-ipaddr triggers/network-write-port triggers/post-app-clone-setup triggers/post-create triggers/post-delete triggers/report
build-in-docker: clean
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT)/plugins/network \
		$(BUILD_IMAGE) \
		bash -c "GO_ARGS='$(GO_ARGS)' make -j4 build" || exit $$?

build: commands subcommands triggers
	$(MAKE) triggers-copy

commands: **/**/commands.go
	go build $(GO_ARGS) -o commands src/commands/commands.go

subcommands: $(SUBCOMMANDS)

subcommands/%: src/subcommands/*/%.go
	go build $(GO_ARGS) -o $@ $<

clean:
	rm -rf commands subcommands triggers network-* install post-app-clone-setup post-create post-delete report

src-clean:
	rm -rf .gitignore src triggers vendor Makefile *.go glide.*

triggers: $(TRIGGERS)

triggers/%: src/triggers/*/%.go
	go build $(GO_ARGS) -o $@ $<

triggers-copy:
	cp triggers/* .
