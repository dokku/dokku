include ../../common.mk

GO_ARGS ?= -a -tags netgo

SUBCOMMANDS = subcommands/gc subcommands/purge-cache

build-in-docker: clean
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT)/plugins/repo \
		$(BUILD_IMAGE) \
		bash -c "GO_ARGS='$(GO_ARGS)' make build" || exit $$?

build: commands subcommands

commands: **/**/commands.go
	CGO_ENABLED=0 go build $(GO_ARGS) -o commands src/commands/commands.go

subcommands: $(SUBCOMMANDS)

subcommands/%: src/subcommands/*/%.go
	CGO_ENABLED=0 go build $(GO_ARGS) -o $@ $<

clean:
	rm -rf commands subcommands

src-clean:
	rm -rf .gitignore src vendor Makefile
