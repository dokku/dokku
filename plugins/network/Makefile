include ../../common.mk

GO_ARGS ?= -a

SUBCOMMANDS = subcommands/rebuild subcommands/rebuildall subcommands/report subcommands/set
TRIGGERS = install network-build-config network-compute-ports network-config-exists network-get-ipaddr network-get-listeners network-get-port network-get-property network-write-ipaddr network-write-port post-create post-delete
build-in-docker: clean
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT)/plugins/network \
		$(BUILD_IMAGE) \
		bash -c "GO_ARGS='$(GO_ARGS)' make build" || exit $$?

build: commands subcommands triggers

commands: **/**/commands.go
	CGO_ENABLED=0 go build $(GO_ARGS) -o commands src/commands/commands.go

subcommands: $(SUBCOMMANDS)

subcommands/%: src/subcommands/*/%.go
	CGO_ENABLED=0 go build $(GO_ARGS) -o $@ $<

clean:
	rm -rf commands subcommands network-* install post-create post-delete

src-clean:
	rm -rf .gitignore src vendor Makefile

triggers: $(TRIGGERS)

triggers/%: src/triggers/*/%.go
	CGO_ENABLED=0 go build $(GO_ARGS) -o $@ $<
