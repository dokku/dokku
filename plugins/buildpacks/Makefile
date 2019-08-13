include ../../common.mk

GO_ARGS ?= -a

SUBCOMMANDS = subcommands/add subcommands/clear subcommands/list subcommands/remove subcommands/report subcommands/set
TRIGGERS = install post-delete post-extract report
build-in-docker: clean
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT)/plugins/buildpacks \
		$(BUILD_IMAGE) \
		bash -c "GO_ARGS='$(GO_ARGS)' make -j4 build" || exit $$?

build: commands subcommands triggers

commands: **/**/commands.go
	go build $(GO_ARGS) -o commands src/commands/commands.go

subcommands: $(SUBCOMMANDS)

subcommands/%: src/subcommands/*/%.go
	go build $(GO_ARGS) -o $@ $<

clean:
	rm -rf commands subcommands triggers install post-delete post-extract report

src-clean:
	rm -rf .gitignore src triggers vendor Makefile *.go glide.*

triggers: **/**/triggers.go $(TRIGGERS)
	go build $(GO_ARGS) -o triggers src/triggers/triggers.go

$(TRIGGERS): 
	ln -s triggers $@
