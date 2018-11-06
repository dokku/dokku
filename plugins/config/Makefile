include ../../common.mk

GO_ARGS ?= -a

SUBCOMMANDS = subcommands/export subcommands/get subcommands/set subcommands/unset subcommands/keys subcommands/bundle

build-in-docker: clean
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT)/plugins/config \
		$(BUILD_IMAGE) \
		bash -c "GO_ARGS='$(GO_ARGS)' make build" || exit $$?

build: commands subcommands

commands: **/**/commands.go
	go build $(GO_ARGS) -o commands cmd/commands/commands.go

subcommands: $(SUBCOMMANDS)

subcommands/%: cmd/subcommands/*/%.go
	go build $(GO_ARGS) -o $@ $<

clean:
	rm -rf commands subcommands

src-clean:
	rm -rf .gitignore cmd triggers vendor Makefile *.go
