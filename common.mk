GO_ARGS ?=
GO_PLUGIN_MAKE_TARGET ?= build
GO_REPO_ROOT := /go/src/github.com/dokku/dokku
BUILD_IMAGE := golang:1.15.6

.PHONY: build-in-docker build clean src-clean

build: $(BUILD)

build-in-docker: clean
	mkdir -p /tmp/dokku-go-build-cache
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-v /tmp/dokku-go-build-cache:/root/.cache \
		-e PLUGIN_NAME=$(PLUGIN_NAME) \
		-e GO111MODULE=on \
		-w $(GO_REPO_ROOT)/plugins/$(PLUGIN_NAME) \
		$(BUILD_IMAGE) \
		bash -c "GO_ARGS='$(GO_ARGS)' make -j4 $(GO_PLUGIN_MAKE_TARGET)" || exit $$?

clean:
	rm -rf $(BUILD)
	find . -xtype l -delete

commands: **/**/commands.go
	go build -ldflags="-s -w" $(GO_ARGS) -o commands src/commands/commands.go

subcommands:
	go build -ldflags="-s -w" $(GO_ARGS) -o subcommands/subcommands src/subcommands/subcommands.go
	$(MAKE) $(SUBCOMMANDS)

subcommands/%:
	ln -sf subcommands $@

src-clean:
	rm -rf .gitignore src vendor Makefile *.go glide.* go.sum go.mod

triggers:
	go build -ldflags="-s -w" $(GO_ARGS) -o triggers src/triggers/triggers.go
	$(MAKE) $(TRIGGERS)

triggers/%:
	ln -sf triggers $(shell echo $@ | cut -d '/' -f2)
