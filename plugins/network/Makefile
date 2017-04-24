include ../../common.mk

build-in-docker: clean
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT)/plugins/network \
		$(BUILD_IMAGE) \
		bash -c "make build" || exit $$?

build: commands subcommands triggers
subcommands: subcommands/rebuild subcommands/rebuildall
triggers: network-build-config network-compute-ports network-get-ipaddr network-get-listeners network-get-port network-write-ipaddr network-write-port
commands: **/**/commands.go
	go build -a -o commands src/commands/commands.go

subcommands/rebuild: **/**/**/rebuild.go
	go build -a -o subcommands/rebuild src/subcommands/rebuild/rebuild.go

subcommands/rebuildall: **/**/**/rebuildall.go
	go build -a -o subcommands/rebuildall src/subcommands/rebuildall/rebuildall.go

network-build-config: **/**/**/network-build-config.go
	go build -a -o network-build-config  src/triggers/network-build-config/network-build-config.go

network-compute-ports: **/**/**/network-compute-ports.go
	go build -a -o network-compute-ports src/triggers/network-compute-ports/network-compute-ports.go

network-get-ipaddr: **/**/**/network-get-ipaddr.go
	go build -a -o network-get-ipaddr    src/triggers/network-get-ipaddr/network-get-ipaddr.go

network-get-listeners: **/**/**/network-get-listeners.go
	go build -a -o network-get-listeners src/triggers/network-get-listeners/network-get-listeners.go

network-get-port: **/**/**/network-get-port.go
	go build -a -o network-get-port      src/triggers/network-get-port/network-get-port.go

network-write-ipaddr: **/**/**/network-write-ipaddr.go
	go build -a -o network-write-ipaddr  src/triggers/network-write-ipaddr/network-write-ipaddr.go

network-write-port: **/**/**/network-write-port.go
	go build -a -o network-write-port    src/triggers/network-write-port/network-write-port.go

clean:
	rm -rf commands subcommands network-*

src-clean:
	rm -rf .gitignore src vendor Makefile
