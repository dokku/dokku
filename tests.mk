shellcheck:
ifeq ($(shell shellcheck > /dev/null 2>&1 ; echo $$?),127)
ifeq ($(shell uname),Darwin)
		brew install shellcheck
else
		sudo add-apt-repository 'deb http://archive.ubuntu.com/ubuntu trusty-backports main restricted universe multiverse'
		sudo apt-get update && sudo apt-get install -y shellcheck
endif
endif

ci-dependencies: shellcheck bats

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

ifeq ($(shell grep 22333 /etc/ssh/sshd_config),)
	sed --in-place "s:^Port 22:Port 22 \\nPort 22333:g" /etc/ssh/sshd_config
	restart ssh
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

bats:
	git clone https://github.com/sstephenson/bats.git /tmp/bats
	cd /tmp/bats && sudo ./install.sh /usr/local
	rm -rf /tmp/bats

lint:
	# these are disabled due to their expansive existence in the codebase. we should clean it up though
	# SC2034: VAR appears unused - https://github.com/koalaman/shellcheck/wiki/SC2034
	# SC2086: Double quote to prevent globbing and word splitting - https://github.com/koalaman/shellcheck/wiki/SC2086
	# SC2143: Instead of [ -n $(foo | grep bar) ], use foo | grep -q bar - https://github.com/koalaman/shellcheck/wiki/SC2143
	# SC2001: See if you can use ${variable//search/replace} instead. - https://github.com/koalaman/shellcheck/wiki/SC2001
	@echo linting...
	@$(QUIET) shellcheck -e SC2029 ./contrib/dokku_client.sh
	@$(QUIET) find . -not -path '*/\.*' | xargs file | egrep "shell|bash" | egrep -v "directory|toml" | awk '{ print $$1 }' | sed 's/://g' | grep -v dokku_client.sh | xargs shellcheck -e SC2034,SC2086,SC2143,SC2001

unit-tests:
	@echo running unit tests...
ifndef UNIT_TEST_BATCH
	@$(QUIET) bats tests/unit
else
	@$(QUIET) ./tests/ci/unit_test_runner.sh $$UNIT_TEST_BATCH
endif

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
	@$(QUIET) $(MAKE) deploy-test-config
	@$(QUIET) $(MAKE) deploy-test-clojure
	@$(QUIET) $(MAKE) deploy-test-dockerfile
	@$(QUIET) $(MAKE) deploy-test-dockerfile-noexpose
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
