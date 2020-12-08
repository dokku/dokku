SYSTEM := $(shell sh -c 'uname -s 2>/dev/null')
DOKKU_SSH_PORT ?= 22

bats:
ifeq ($(SYSTEM),Darwin)
ifneq ($(shell bats --version >/dev/null 2>&1 ; echo $$?),0)
	brew install bats-core
endif
else
	git clone https://github.com/josegonzalez/bats-core.git /tmp/bats
	cd /tmp/bats && sudo ./install.sh /usr/local
	rm -rf /tmp/bats
endif

shellcheck:
ifneq ($(shell shellcheck --version >/dev/null 2>&1 ; echo $$?),0)
ifeq ($(SYSTEM),Darwin)
	brew install shellcheck
else
	sudo apt-get update -qq && sudo apt-get install -qq -y shellcheck
endif
endif

shfmt:
ifneq ($(shell shfmt --version >/dev/null 2>&1 ; echo $$?),0)
ifeq ($(shfmt),Darwin)
	brew install shfmt
else
	wget -qO /tmp/shfmt https://github.com/mvdan/sh/releases/download/v2.6.2/shfmt_v2.6.2_linux_amd64
	chmod +x /tmp/shfmt
	sudo mv /tmp/shfmt /usr/local/bin/shfmt
endif
endif

xmlstarlet:
ifneq ($(shell xmlstarlet --version >/dev/null 2>&1 ; echo $$?),0)
ifeq ($(SYSTEM),Darwin)
	brew install xmlstarlet
else
	sudo apt-get update -qq && sudo apt-get install -qq -y xmlstarlet
endif
endif

ci-dependencies: bats shellcheck xmlstarlet

setup-deploy-tests:
ifdef ENABLE_DOKKU_TRACE
	echo "-----> Enable dokku trace"
	dokku trace:on
endif
	@echo "Setting dokku.me in /etc/hosts"
	sudo /bin/bash -c "[[ `ping -c1 dokku.me >/dev/null 2>&1; echo $$?` -eq 0 ]] || echo \"127.0.0.1  dokku.me *.dokku.me www.test.app.dokku.me\" >> /etc/hosts"

	@echo "-----> Generating keypair..."
	mkdir -p /root/.ssh
	rm -f /root/.ssh/dokku_test_rsa*
	echo -e  "y\n" | ssh-keygen -f /root/.ssh/dokku_test_rsa -t rsa -N ''
	chmod 700 /root/.ssh
	chmod 600 /root/.ssh/dokku_test_rsa
	chmod 644 /root/.ssh/dokku_test_rsa.pub

	@echo "-----> Setting up ssh config..."
ifneq ($(shell ls /root/.ssh/config >/dev/null 2>&1 ; echo $$?),0)
	echo "Host dokku.me \\r\\n Port $(DOKKU_SSH_PORT) \\r\\n RequestTTY yes \\r\\n IdentityFile /root/.ssh/dokku_test_rsa" >> /root/.ssh/config
	echo "Host 127.0.0.1 \\r\\n Port 22333 \\r\\n RequestTTY yes \\r\\n IdentityFile /root/.ssh/dokku_test_rsa" >> /root/.ssh/config
else ifeq ($(shell grep dokku.me /root/.ssh/config),)
	echo "Host dokku.me \\r\\n Port $(DOKKU_SSH_PORT) \\r\\n RequestTTY yes \\r\\n IdentityFile /root/.ssh/dokku_test_rsa" >> /root/.ssh/config
	echo "Host 127.0.0.1 \\r\\n Port 22333 \\r\\n RequestTTY yes \\r\\n IdentityFile /root/.ssh/dokku_test_rsa" >> /root/.ssh/config
else
	sed --in-place 's/Port 22 \r/Port $(DOKKU_SSH_PORT) \r/g' /root/.ssh/config
	cat /root/.ssh/config
endif

ifneq ($(wildcard /etc/ssh/sshd_config),)
	sed --in-place "s/^#Port 22$\/Port 22/g" /etc/ssh/sshd_config
ifeq ($(shell grep 22333 /etc/ssh/sshd_config),)
	sed --in-place "s:^Port 22:Port 22 \\nPort 22333:g" /etc/ssh/sshd_config
endif
	service ssh restart
endif

	@echo "-----> Installing SSH public key..."
	echo "" > /home/dokku/.ssh/authorized_keys
	sudo sshcommand acl-remove dokku test
	cat /root/.ssh/dokku_test_rsa.pub | sudo sshcommand acl-add dokku test
	chmod 700 /home/dokku/.ssh
	chmod 600 /home/dokku/.ssh/authorized_keys

ifeq ($(shell grep dokku.me /home/dokku/VHOST 2>/dev/null),)
	@echo "-----> Setting default VHOST to dokku.me..."
	echo "dokku.me" > /home/dokku/VHOST
endif
ifeq ($(DOKKU_SSH_PORT), 22)
	$(MAKE) prime-ssh-known-hosts
endif

setup-docker-deploy-tests: setup-deploy-tests
ifdef ENABLE_DOKKU_TRACE
	echo "-----> Enable dokku trace"
	docker exec -ti dokku bash -c "dokku trace:on"
endif
	docker exec -ti dokku bash -c "sshcommand acl-remove dokku test"
	docker exec -ti dokku bash -c "echo `cat /root/.ssh/dokku_test_rsa.pub` | sshcommand acl-add dokku test"
	$(MAKE) prime-ssh-known-hosts

prime-ssh-known-hosts:
	@echo "-----> Intitial SSH connection to populate known_hosts..."
	@echo "=====> SSH dokku.me"
	ssh -o StrictHostKeyChecking=no dokku@dokku.me help >/dev/null
	@echo "=====> SSH 127.0.0.1"
	ssh -o StrictHostKeyChecking=no dokku@127.0.0.1 help >/dev/null

lint-setup:
	@mkdir -p test-results/shellcheck tmp/shellcheck
	@find . -not -path '*/\.*' -not -path './debian/*' -not -path './docs/*' -not -path './tests/*' -not -path './vendor/*' -type f | xargs file | grep text | awk -F ':' '{ print $$1 }' | xargs head -n1 | grep -B1 "bash" | grep "==>" | awk '{ print $$2 }' > tmp/shellcheck/test-files
	@cat tests/shellcheck-exclude | sed -n -e '/^# SC/p' | cut -d' ' -f2 | paste -d, -s > tmp/shellcheck/exclude

lint-ci: lint-setup
	# these are disabled due to their expansive existence in the codebase. we should clean it up though
	@cat tests/shellcheck-exclude | sed -n -e '/^# SC/p'
	@echo linting...
	@cat tmp/shellcheck/test-files | xargs shellcheck -e $(shell cat tmp/shellcheck/exclude) | tests/shellcheck-to-junit --output test-results/shellcheck/results.xml --files tmp/shellcheck/test-files --exclude $(shell cat tmp/shellcheck/exclude)

lint-shfmt: shfmt
	# verifying via shfmt
	# shfmt -l -bn -ci -i 2 -d .
	@shfmt -l -bn -ci -i 2 -d .

lint: lint-shfmt lint-ci

ci-go-coverage:
	@$(MAKE) ci-go-coverage-plugin PLUGIN_NAME=common
	@$(MAKE) ci-go-coverage-plugin PLUGIN_NAME=config
	@$(MAKE) ci-go-coverage-plugin PLUGIN_NAME=network

ci-go-coverage-plugin:
	docker run --rm -ti \
		-e DOKKU_ROOT=/home/dokku \
		-e CODACY_TOKEN=$$CODACY_TOKEN \
		-e CIRCLE_SHA1=$$CIRCLE_SHA1 \
		-e GO111MODULE=on \
		-v $$PWD:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT) \
		$(BUILD_IMAGE) \
		bash -c "cd plugins/$(PLUGIN_NAME) && \
			go get github.com/onsi/gomega github.com/schrej/godacov github.com/haya14busa/goverage && \
			goverage -v -coverprofile=coverage.out && \
			godacov -t $$CODACY_TOKEN -r ./coverage.out -c $$CIRCLE_SHA1" || exit $$?

go-tests:
	@$(MAKE) go-test-plugin PLUGIN_NAME=common
	@$(MAKE) go-test-plugin PLUGIN_NAME=config
	@$(MAKE) go-test-plugin PLUGIN_NAME=network

go-test-plugin:
	@echo running go unit tests...
	docker run --rm -ti \
		-e DOKKU_ROOT=/home/dokku \
		-e GO111MODULE=on \
		-v $$PWD:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT) \
		$(BUILD_IMAGE) \
		bash -c "cd plugins/$(PLUGIN_NAME) && go get github.com/onsi/gomega && go test -v -p 1 -race " || exit $$?

unit-tests: go-tests
	@echo running bats unit tests...
ifndef UNIT_TEST_BATCH
	@$(QUIET) bats tests/unit
else
	@$(QUIET) ./tests/ci/unit_test_runner.sh $$UNIT_TEST_BATCH
endif

deploy-test-go-fail-predeploy:
	@echo deploying go-fail-predeploy app...
	cd tests && ./test_deploy ./apps/go-fail-predeploy dokku.me '' true

deploy-test-go-fail-postdeploy:
	@echo deploying go-fail-postdeploy app...
	cd tests && ./test_deploy ./apps/go-fail-postdeploy dokku.me '' true

deploy-test-checks-root:
	@echo deploying checks-root app...
	cd tests && ./test_deploy ./apps/checks-root dokku.me '' true

deploy-test-main-branch:
	@echo deploying checks-root app to main branch...
	cd tests && ./test_deploy ./apps/checks-root dokku.me '' true main

deploy-test-clojure:
	@echo deploying config app...
	cd tests && ./test_deploy ./apps/clojure dokku.me

deploy-test-config:
	@echo deploying config app...
	cd tests && ./test_deploy ./apps/config dokku.me

deploy-test-dockerfile:
	@echo deploying dockerfile app...
	cd tests && ./test_deploy ./apps/dockerfile dokku.me

deploy-test-dockerfile-noexpose:
	@echo deploying dockerfile-noexpose app...
	cd tests && ./test_deploy ./apps/dockerfile-noexpose dokku.me

deploy-test-dockerfile-procfile:
	@echo deploying dockerfile-procfile app...
	cd tests && ./test_deploy ./apps/dockerfile-procfile dokku.me

deploy-test-gitsubmodules:
	@echo deploying gitsubmodules app...
	cd tests && ./test_deploy ./apps/gitsubmodules dokku.me

deploy-test-go:
	@echo deploying go app...
	cd tests && ./test_deploy ./apps/go dokku.me

deploy-test-java:
	@echo deploying java app...
	cd tests && ./test_deploy ./apps/java dokku.me

deploy-test-multi:
	@echo deploying multi app...
	cd tests && ./test_deploy ./apps/multi dokku.me

deploy-test-nodejs-express:
	@echo deploying nodejs-express app...
	cd tests && ./test_deploy ./apps/nodejs-express dokku.me

deploy-test-nodejs-express-noprocfile:
	@echo deploying nodejs-express app with no Procfile...
	cd tests && ./test_deploy ./apps/nodejs-express-noprocfile dokku.me

deploy-test-nodejs-worker:
	@echo deploying nodejs-worker app...
	cd tests && ./test_deploy ./apps/nodejs-worker dokku.me

deploy-test-php:
	@echo deploying php app...
	cd tests && ./test_deploy ./apps/php dokku.me

deploy-test-python-flask:
	@echo deploying python-flask app...
	cd tests && ./test_deploy ./apps/python-flask dokku.me

deploy-test-ruby:
	@echo deploying ruby app...
	cd tests && ./test_deploy ./apps/ruby dokku.me

deploy-test-scala:
	@echo deploying scala app...
	cd tests && ./test_deploy ./apps/scala dokku.me

deploy-test-static:
	@echo deploying static app...
	cd tests && ./test_deploy ./apps/static dokku.me

deploy-tests:
	@echo running deploy tests...
	@$(QUIET) $(MAKE) deploy-test-checks-root
	@$(QUIET) $(MAKE) deploy-test-main-branch
	@$(QUIET) $(MAKE) deploy-test-go-fail-predeploy
	@$(QUIET) $(MAKE) deploy-test-go-fail-postdeploy
	@$(QUIET) $(MAKE) deploy-test-config
	@$(QUIET) $(MAKE) deploy-test-clojure
	@$(QUIET) $(MAKE) deploy-test-dockerfile
	@$(QUIET) $(MAKE) deploy-test-dockerfile-noexpose
	@$(QUIET) $(MAKE) deploy-test-dockerfile-procfile
	@$(QUIET) $(MAKE) deploy-test-gitsubmodules
	@$(QUIET) $(MAKE) deploy-test-go
	@$(QUIET) $(MAKE) deploy-test-java
	@$(QUIET) $(MAKE) deploy-test-multi
	@$(QUIET) $(MAKE) deploy-test-nodejs-express
	@$(QUIET) $(MAKE) deploy-test-nodejs-express-noprocfile
	@$(QUIET) $(MAKE) deploy-test-nodejs-worker
	@$(QUIET) $(MAKE) deploy-test-php
	@$(QUIET) $(MAKE) deploy-test-python-flask
	@$(QUIET) $(MAKE) deploy-test-scala
	@$(QUIET) $(MAKE) deploy-test-static

test: setup-deploy-tests lint unit-tests deploy-tests

test-ci:
	@mkdir -p test-results/bats
	@cd tests/unit && echo "executing tests: $(shell cd tests/unit ; circleci tests glob *.bats | circleci tests split --split-by=timings --timings-type=classname | xargs)"
	cd tests/unit && bats --formatter bats-format-junit -e -T -o ../../test-results/bats $(shell cd tests/unit ; circleci tests glob *.bats | circleci tests split --split-by=timings --timings-type=classname | xargs)

test-ci-docker: setup-docker-deploy-tests deploy-test-checks-root deploy-test-config deploy-test-multi deploy-test-go-fail-predeploy deploy-test-go-fail-postdeploy
