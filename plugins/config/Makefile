.PHONY: build-in-docker build clean

GO_REPO_ROOT ?= /go/src/github.com/dokku/dokku
BUILD_IMAGE ?= golang:1.7.1

build-in-docker: clean
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT)/plugins/config \
		$(BUILD_IMAGE) \
		bash -c "make build" || exit $$?

build: commands subcommands
subcommands: subcommands/export subcommands/set subcommands/get subcommands/unset

commands: **/**/commands.go
	go build -a -o commands src/commands/commands.go

subcommands/export: **/**/**/export.go
	go build -a -o subcommands/export src/subcommands/export/export.go

subcommands/set: **/**/**/set.go
	go build -a -o subcommands/set src/subcommands/set/set.go

subcommands/get: **/**/**/get.go
	go build -a -o subcommands/get src/subcommands/get/get.go

subcommands/unset: **/**/**/unset.go
	go build -a -o subcommands/unset src/subcommands/unset/unset.go

clean:
	rm -rf commands subcommands
