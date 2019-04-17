include ../../common.mk

GO_ARGS ?= -a

SUBCOMMANDS = subcommands/limit subcommands/limit-clear subcommands/report subcommands/reserve subcommands/reserve-clear
TRIGGERS = triggers/docker-args-process-deploy triggers/install triggers/post-delete triggers/report triggers/resource-get-property
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
	rm -rf commands subcommands triggers docker-args-process-deploy docker-args-process-run install post-delete report resource-get-property

src-clean:
	rm -rf .gitignore src triggers vendor Makefile *.go glide.*

triggers: $(TRIGGERS)

triggers/%: src/triggers/*/%.go
	go build $(GO_ARGS) -o $@ $<

triggers-copy:
	cp triggers/* .
	ln -sf docker-args-process-deploy docker-args-process-run
