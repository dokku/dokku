include ../../common.mk

GO_ARGS ?= -a

SUBCOMMANDS = subcommands/limit subcommands/reserve
TRIGGERS = triggers/install triggers/post-delete
build-in-docker: clean
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT)/plugins/resource \
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
	rm -rf commands subcommands triggers install post-delete

src-clean:
	rm -rf .gitignore src triggers vendor Makefile *.go

triggers: $(TRIGGERS)

triggers/%: src/triggers/*/%.go
	go build $(GO_ARGS) -o $@ $<

triggers-copy:
	cp triggers/* .
