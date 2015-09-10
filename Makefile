DOKKU_VERSION = master

SSHCOMMAND_URL ?= https://raw.github.com/progrium/sshcommand/master/sshcommand
PLUGN_URL ?= https://github.com/progrium/plugn/releases/download/v0.1.0/plugn_0.1.0_linux_x86_64.tgz
STACK_URL ?= https://github.com/gliderlabs/herokuish.git
PREBUILT_STACK_URL ?= gliderlabs/herokuish:latest
DOKKU_LIB_ROOT ?= /var/lib/dokku
PLUGINS_PATH ?= ${DOKKU_LIB_ROOT}/plugins
CORE_PLUGINS_PATH ?= ${DOKKU_LIB_ROOT}/core-plugins

# If the first argument is "vagrant-dokku"...
ifeq (vagrant-dokku,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "vagrant-dokku"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

ifeq ($(CIRCLECI),true)
	BUILD_STACK_TARGETS = circleci deps build
else
	BUILD_STACK_TARGETS = build-in-docker
endif

.PHONY: all apt-update install version copyfiles man-db plugins dependencies sshcommand plugn docker aufs stack count dokku-installer vagrant-acl-add vagrant-dokku

include tests.mk
include deb.mk

all:
	# Type "make install" to install.

install: dependencies version copyfiles plugin-dependencies plugins

release: deb-all package_cloud packer

package_cloud:
	package_cloud push dokku/dokku/ubuntu/trusty herokuish*.deb
	package_cloud push dokku/dokku/ubuntu/trusty sshcommand*.deb
	package_cloud push dokku/dokku/ubuntu/trusty plugn*.deb
	package_cloud push dokku/dokku/ubuntu/trusty rubygem*.deb
	package_cloud push dokku/dokku/ubuntu/trusty dokku*.deb

packer:
	packer build contrib/packer.json

copyfiles:
	cp dokku /usr/local/bin/dokku
	mkdir -p ${CORE_PLUGINS_PATH} ${PLUGINS_PATH}
	rm -rf ${CORE_PLUGINS_PATH}/*
	test -d ${CORE_PLUGINS_PATH}/enabled || PLUGIN_PATH=${CORE_PLUGINS_PATH} plugn init
	test -d ${PLUGINS_PATH}/enabled || PLUGIN_PATH=${PLUGINS_PATH} plugn init
	find plugins/ -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | while read plugin; do \
		rm -Rf ${CORE_PLUGINS_PATH}/available/$$plugin && \
		rm -Rf ${PLUGINS_PATH}/available/$$plugin && \
		rm -rf ${CORE_PLUGINS_PATH}/$$plugin && \
		rm -rf ${PLUGINS_PATH}/$$plugin && \
		cp -R plugins/$$plugin ${CORE_PLUGINS_PATH}/available && \
		ln -s ${CORE_PLUGINS_PATH}/available/$$plugin ${PLUGINS_PATH}/available; \
		PLUGIN_PATH=${CORE_PLUGINS_PATH} plugn enable $$plugin ;\
		PLUGIN_PATH=${PLUGINS_PATH} plugn enable $$plugin ;\
		done
	chown dokku:dokku -R ${PLUGINS_PATH} ${CORE_PLUGINS_PATH}
	$(MAKE) addman

addman:
	mkdir -p /usr/local/share/man/man1
	help2man -Nh help -v version -n "configure and get information from your dokku installation" -o /usr/local/share/man/man1/dokku.1 dokku
	mandb

version:
	git describe --tags > ~dokku/VERSION  2> /dev/null || echo '~${DOKKU_VERSION} ($(shell date -uIminutes))' > ~dokku/VERSION

plugin-dependencies: plugn
	dokku plugin:install-dependencies --core

plugins: plugn docker
	dokku plugin:install --core

dependencies: apt-update sshcommand plugn docker help2man man-db
	$(MAKE) -e stack

apt-update:
	apt-get update

help2man:
	apt-get install -qq -y help2man

man-db:
	apt-get install -qq -y man-db

sshcommand:
	wget -qO /usr/local/bin/sshcommand ${SSHCOMMAND_URL}
	chmod +x /usr/local/bin/sshcommand
	sshcommand create dokku /usr/local/bin/dokku

plugn:
	wget -qO /tmp/plugn_latest.tgz ${PLUGN_URL}
	tar xzf /tmp/plugn_latest.tgz -C /usr/local/bin

docker: aufs
	apt-get install -qq -y curl
	egrep -i "^docker" /etc/group || groupadd docker
	usermod -aG docker dokku
ifndef CI
	curl -sSL https://get.docker.com/gpg | apt-key add -
	echo deb https://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list
	apt-get update
ifdef DOCKER_VERSION
	apt-get install -qq -y lxc-docker-${DOCKER_VERSION}
else
	apt-get install -qq -y lxc-docker-1.6.2
endif
	sleep 2 # give docker a moment i guess
endif

aufs:
ifndef CI
	lsmod | grep aufs || modprobe aufs || apt-get install -qq -y linux-image-extra-`uname -r` > /dev/null
endif

stack:
ifeq ($(shell test -e /var/run/docker.sock && touch -c /var/run/docker.sock && echo $$?),0)
ifdef BUILD_STACK
	@echo "Start building herokuish from source"
	docker images | grep gliderlabs/herokuish || (git clone ${STACK_URL} /tmp/herokuish && cd /tmp/herokuish && IMAGE_NAME=gliderlabs/herokuish BUILD_TAG=latest VERSION=master make -e ${BUILD_STACK_TARGETS} && rm -rf /tmp/herokuish)
else
ifeq ($(shell echo ${PREBUILT_STACK_URL} | egrep -q 'http.*://' && echo $$?),0)
	@echo "Start importing herokuish from ${PREBUILT_STACK_URL}"
	docker images | grep gliderlabs/herokuish || curl --silent -L ${PREBUILT_STACK_URL} | gunzip -cd | docker import - gliderlabs/herokuish
else
	@echo "Start pulling herokuish from ${PREBUILT_STACK_URL}"
	docker images | grep gliderlabs/herokuish || docker pull ${PREBUILT_STACK_URL}
endif
endif
endif

count:
	@echo "Core lines:"
	@cat dokku bootstrap.sh | sed 's/^$$//g' | wc -l
	@echo "Plugin lines:"
	@find plugins -type f -not -name .DS_Store | xargs cat | sed 's/^$$//g' | wc -l
	@echo "Test lines:"
	@find tests -type f -not -name .DS_Store | xargs cat | sed 's/^$$//g' | wc -l

dokku-installer:
	apt-get install -qq -y ruby
	test -f /var/lib/dokku/.dokku-installer-created || gem install rack -v 1.5.2 --no-rdoc --no-ri
	test -f /var/lib/dokku/.dokku-installer-created || gem install rack-protection -v 1.5.3 --no-rdoc --no-ri
	test -f /var/lib/dokku/.dokku-installer-created || gem install sinatra -v 1.4.5 --no-rdoc --no-ri
	test -f /var/lib/dokku/.dokku-installer-created || gem install tilt -v 1.4.1 --no-rdoc --no-ri
	test -f /var/lib/dokku/.dokku-installer-created || ruby contrib/dokku-installer.rb onboot
	test -f /var/lib/dokku/.dokku-installer-created || service dokku-installer start
	test -f /var/lib/dokku/.dokku-installer-created || service nginx reload
	test -f /var/lib/dokku/.dokku-installer-created || touch /var/lib/dokku/.dokku-installer-created

vagrant-acl-add:
	vagrant ssh -- sudo sshcommand acl-add dokku $(USER)

vagrant-dokku:
	vagrant ssh -- "sudo -H -u root bash -c 'dokku $(RUN_ARGS)'"

