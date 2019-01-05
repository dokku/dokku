SYSTEM := $(shell sh -c 'uname -s 2>/dev/null')

bats:
ifeq ($(SYSTEM),Darwin)
ifneq ($(shell bats --version > /dev/null 2>&1 ; echo $$?),0)
	brew install bats-core
endif
else
	git clone https://github.com/josegonzalez/bats-core.git /tmp/bats
	cd /tmp/bats && sudo ./install.sh /usr/local
	rm -rf /tmp/bats
endif

shellcheck:
ifneq ($(shell shellcheck --version > /dev/null 2>&1 ; echo $$?),0)
ifeq ($(SYSTEM),Darwin)
	brew install shellcheck
else
	sudo add-apt-repository 'deb http://archive.ubuntu.com/ubuntu trusty-backports main restricted universe multiverse'
	sudo rm -rf /var/lib/apt/lists/* && sudo apt-get clean
	sudo apt-get update -qq && sudo apt-get install -qq -y shellcheck
endif
endif

xmlstarlet:
ifneq ($(shell xmlstarlet --version > /dev/null 2>&1 ; echo $$?),0)
ifeq ($(SYSTEM),Darwin)
	brew install xmlstarlet
else
	sudo apt-get update -qq && sudo apt-get install -qq -y xmlstarlet
endif
endif

ci-dependencies: shellcheck bats xmlstarlet

setup-deploy-tests:
	mkdir -p /home/dokku
ifdef ENABLE_DOKKU_TRACE
	echo "-----> Enabling tracing"
	echo "export DOKKU_TRACE=1" >> /home/dokku/dokkurc
endif
	@echo "Setting dokku.me in /etc/hosts"
	sudo /bin/bash -c "[[ `ping -c1 dokku.me > /dev/null 2>&1; echo $$?` -eq 0 ]] || echo \"127.0.0.1  dokku.me *.dokku.me www.test.app.dokku.me\" >> /etc/hosts"

	@echo "-----> Generating keypair..."
	mkdir -p /root/.ssh
	rm -f /root/.ssh/dokku_test_rsa*
	echo -e  "y\n" | ssh-keygen -f /root/.ssh/dokku_test_rsa -t rsa -N ''
	chmod 600 /root/.ssh/dokku_test_rsa*

	@echo "-----> Setting up ssh config..."
ifneq ($(shell ls /root/.ssh/config > /dev/null 2>&1 ; echo $$?),0)
	echo "Host dokku.me \\r\\n RequestTTY yes \\r\\n IdentityFile /root/.ssh/dokku_test_rsa" >> /root/.ssh/config
	echo "Host 127.0.0.1 \\r\\n Port 22333 \\r\\n RequestTTY yes \\r\\n IdentityFile /root/.ssh/dokku_test_rsa" >> /root/.ssh/config
else ifeq ($(shell grep dokku.me /root/.ssh/config),)
	echo "Host dokku.me \\r\\n RequestTTY yes \\r\\n IdentityFile /root/.ssh/dokku_test_rsa" >> /root/.ssh/config
	echo "Host 127.0.0.1 \\r\\n Port 22333 \\r\\n RequestTTY yes \\r\\n IdentityFile /root/.ssh/dokku_test_rsa" >> /root/.ssh/config
endif

ifneq ($(wildcard /etc/ssh/sshd_config),)
ifeq ($(shell grep 22333 /etc/ssh/sshd_config),)
	sed --in-place "s:^Port 22:Port 22 \\nPort 22333:g" /etc/ssh/sshd_config
	restart ssh
endif
endif

	@echo "-----> Installing SSH public key..."
	sudo sshcommand acl-remove dokku test
	cat /root/.ssh/dokku_test_rsa.pub | sudo sshcommand acl-add dokku test

	@echo "-----> Intitial SSH connection to populate known_hosts..."
	ssh -o StrictHostKeyChecking=no dokku@dokku.me help > /dev/null
	ssh -o StrictHostKeyChecking=no dokku@127.0.0.1 help > /dev/null

ifeq ($(shell grep dokku.me /home/dokku/VHOST 2>/dev/null),)
	@echo "-----> Setting default VHOST to dokku.me..."
	echo "dokku.me" > /home/dokku/VHOST
endif

lint:
	# these are disabled due to their expansive existence in the codebase. we should clean it up though
	@cat tests/shellcheck-exclude | sed -n -e '/^# SC/p'
ifeq ($(CIRCLECI),true)
	@echo creating junit output...
	@mkdir -p test-results/shellcheck
	@$(QUIET) find . -not -path '*/\.*' -not -path './debian/*' -type f | xargs file | grep text | awk -F ':' '{ print $$1 }' | xargs head -n1 | egrep -B1 "bash" | grep "==>" | awk '{ print $$2 }' | xargs shellcheck -e $(shell cat tests/shellcheck-exclude | sed -n -e '/^# SC/p' | cut -d' ' -f2 | paste -d, -s) -f checkstyle | xmlstarlet tr tests/checkstyle2junit.xslt > test-results/shellcheck/results.xml
endif
	@echo linting...
	@$(QUIET) find . -not -path '*/\.*' -not -path './debian/*' -type f | xargs file | grep text | awk -F ':' '{ print $$1 }' | xargs head -n1 | egrep -B1 "bash" | grep "==>" | awk '{ print $$2 }' | xargs shellcheck -e $(shell cat tests/shellcheck-exclude | sed -n -e '/^# SC/p' | cut -d' ' -f2 | paste -d, -s)

ci-go-coverage:
	docker run --rm -ti \
		-e DOKKU_ROOT=/home/dokku \
		-e CODACY_TOKEN=$$CODACY_TOKEN \
		-e CIRCLE_SHA1=$$CIRCLE_SHA1 \
		-v $$PWD:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT) \
		$(BUILD_IMAGE) \
		bash -c "go get github.com/onsi/gomega github.com/schrej/godacov github.com/haya14busa/goverage && \
			go list ./... | egrep -v '/vendor/|/tests/apps/' | xargs goverage -v -coverprofile=coverage.out && \
			godacov -t $$CODACY_TOKEN -r ./coverage.out -c $$CIRCLE_SHA1" || exit $$?

go-tests:
	@echo running go unit tests...
	docker run --rm -ti \
		-e DOKKU_ROOT=/home/dokku \
		-v $$PWD:$(GO_REPO_ROOT) \
		-w $(GO_REPO_ROOT) \
		$(BUILD_IMAGE) \
		bash -c "go get github.com/onsi/gomega && \
			go list ./... | egrep -v '/vendor/|/tests/apps/' | xargs go test -v -p 1 -race" || exit $$?

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
	mkdir -p test-results/bats
	@cd tests/unit && echo "executing tests: $(shell cd tests/unit ; circleci tests glob *.bats | circleci tests split --split-by=timings | xargs)"
	cd tests/unit && bats --formatter bats-format-junit -e -T -o ../../test-results/bats $(shell cd tests/unit ; circleci tests glob *.bats | circleci tests split --split-by=timings | xargs)
