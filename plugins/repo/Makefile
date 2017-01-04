.PHONY: build-in-docker build commands subcommands

build-in-docker: clean
	docker run --rm \
		-v $$PWD/../..:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT)/plugins/repo \
		$(BUILD_IMAGE) \
		bash -c "make build" || exit $$?

build:
	cd src && go build -a -o ../commands

clean:
	rm -f commands
