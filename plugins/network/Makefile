include ../../common.mk

GO_ARGS ?= -a

SUBCOMMANDS = subcommands/rebuild subcommands/rebuildall subcommands/report subcommands/set
TRIGGERS = triggers/install triggers/network-build-config triggers/network-compute-ports triggers/network-config-exists triggers/network-get-ipaddr triggers/network-get-listeners triggers/network-get-port triggers/network-get-property triggers/network-write-ipaddr triggers/network-write-port triggers/post-create triggers/post-delete triggers/report

build-in-docker: clean
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT)/plugins/network \
		$(BUILD_IMAGE) \
		bash -c "GO_ARGS='$(GO_ARGS)' make build" || exit $$?

build: commands subcommands triggers triggers-copy

commands: **/**/commands.go
	go build $(GO_ARGS) -o commands cmd/commands/commands.go

subcommands: $(SUBCOMMANDS)

subcommands/%: cmd/subcommands/*/%.go
	go build $(GO_ARGS) -o $@ $<

clean:
	rm -rf commands subcommands triggers network-* install post-create post-delete report

src-clean:
	rm -rf .gitignore cmd triggers vendor Makefile *.go

triggers: $(TRIGGERS)

triggers/%: cmd/triggers/*/%.go
	go build $(GO_ARGS) -o $@ $<

triggers-copy:
	cp triggers/* .
