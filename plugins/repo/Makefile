include ../../common.mk

build-in-docker: clean
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT)/plugins/repo \
		$(BUILD_IMAGE) \
		bash -c "make build" || exit $$?

build: commands subcommands
subcommands: subcommands/gc subcommands/purge-cache

commands: **/**/commands.go
	go build -a -o commands src/commands/commands.go

subcommands/gc: **/**/**/gc.go
	go build -a -o subcommands/gc src/subcommands/gc/gc.go

subcommands/purge-cache: **/**/**/purge-cache.go
	go build -a -o subcommands/purge-cache src/subcommands/purge-cache/purge-cache.go

clean:
	rm -rf commands subcommands
