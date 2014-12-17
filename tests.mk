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
	sudo /bin/bash -c "[[ `ping -c1 dokku.me > /dev/null 2>&1; echo $$?` -eq 0 ]] || echo \"127.0.0.1  dokku.me *.dokku.me\" >> /etc/hosts"

	@echo "-----> Generating keypair..."
	mkdir -p /root/.ssh
	rm -f /root/.ssh/dokku_test_rsa*
	echo -e  "y\n" | ssh-keygen -f /root/.ssh/dokku_test_rsa -t rsa -N ''
	chmod 600 /root/.ssh/dokku_test_rsa*

	@echo "-----> Setting up ssh config..."
ifneq ($(shell ls /root/.ssh/config > /dev/null 2>&1 ; echo $$?),0)
	echo "Host dokku.me \\r\\n RequestTTY yes \\r\\n IdentityFile /root/.ssh/dokku_test_rsa" >> /root/.ssh/config
else ifeq ($(shell grep dokku.me /root/.ssh/config),)
	echo "Host dokku.me \\r\\n RequestTTY yes \\r\\n IdentityFile /root/.ssh/dokku_test_rsa" >> /root/.ssh/config
endif

	@echo "-----> Installing SSH public key..."
	sudo sshcommand acl-remove dokku test
	cat /root/.ssh/dokku_test_rsa.pub | sudo sshcommand acl-add dokku test

	@echo "-----> Intitial SSH connection to populate known_hosts..."
	ssh -o StrictHostKeyChecking=no dokku@dokku.me help > /dev/null

ifeq ($(shell grep dokku.me /home/dokku/VHOST),)
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
	@echo linting...
	@$(QUIET) find . -not -path '*/\.*' | xargs file | egrep "shell|bash" | awk '{ print $$1 }' | sed 's/://g' | xargs shellcheck -e SC2034,SC2086

unit-tests:
	@echo running unit tests...
	@$(QUIET) bats tests/unit

deploy-test-config:
	@echo deploying config app...
	cd tests && ./test_deploy ./apps/config dokku.me

deploy-test-gitsubmodules:
	@echo deploying config app...
	cd tests && ./test_deploy ./apps/gitsubmodules dokku.me

deploy-test-go:
	@echo deploying config app...
	cd tests && ./test_deploy ./apps/go dokku.me

deploy-test-java:
	@echo deploying config app...
	cd tests && ./test_deploy ./apps/java dokku.me

deploy-test-multi:
	@echo deploying config app...
	cd tests && ./test_deploy ./apps/multi dokku.me

deploy-test-nodejs-express:
	@echo deploying config app...
	cd tests && ./test_deploy ./apps/nodejs-express dokku.me

deploy-test-php:
	@echo deploying config app...
	cd tests && ./test_deploy ./apps/php dokku.me

deploy-test-python-flask:
	@echo deploying config app...
	cd tests && ./test_deploy ./apps/python-flask dokku.me

deploy-test-static:
	@echo deploying config app...
	cd tests && ./test_deploy ./apps/static dokku.me

deploy-tests:
	@echo running deploy tests...
	# @$(QUIET) bats tests/deploy
	@$(QUIET) $(MAKE) deploy-test-config
	@$(QUIET) $(MAKE) deploy-test-gitsubmodules
	@$(QUIET) $(MAKE) deploy-test-go
	@$(QUIET) $(MAKE) deploy-test-java
	@$(QUIET) $(MAKE) deploy-test-multi
	@$(QUIET) $(MAKE) deploy-test-nodejs-express
	# fails. ref: https://github.com/progrium/buildstep/issues/126
	# @$(QUIET) $(MAKE) deploy-test-php
	@$(QUIET) $(MAKE) deploy-test-python-flask
	# fails on apt-get update... ref: https://github.com/progrium/buildpack-nginx/issues/18
	# @$(QUIET) $(MAKE) deploy-test-static

test: setup-deploy-tests lint unit-tests deploy-tests
