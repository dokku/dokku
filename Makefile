DOKKU_VERSION ?= master

DOCKER_IMAGE_LABELER_VERSION ?= 0.4.1
HEROKUISH_VERSION ?= 0.5.34
NETRC_VERSION ?= 0.5.1
PLUGN_VERSION ?= 0.8.2
PROCFILE_VERSION ?= 0.14.1
SIGIL_VERSION ?= 0.8.1
SSHCOMMAND_VERSION ?= 0.15.0
DOCKER_IMAGE_LABELER_URL ?= https://github.com/dokku/docker-image-labeler/releases/download/v${DOCKER_IMAGE_LABELER_VERSION}/docker-image-labeler_${DOCKER_IMAGE_LABELER_VERSION}_linux_x86_64.tgz
NETRC_URL ?= https://github.com/dokku/netrc/releases/download/v${NETRC_VERSION}/netrc_${NETRC_VERSION}_linux_x86_64.tgz
PLUGN_URL ?= https://github.com/dokku/plugn/releases/download/v${PLUGN_VERSION}/plugn_${PLUGN_VERSION}_linux_x86_64.tgz
PROCFILE_UTIL_URL ?= https://github.com/josegonzalez/go-procfile-util/releases/download/v${PROCFILE_VERSION}/procfile-util_${PROCFILE_VERSION}_linux_x86_64.tgz
SIGIL_URL ?= https://github.com/gliderlabs/sigil/releases/download/v${SIGIL_VERSION}/sigil_${SIGIL_VERSION}_Linux_x86_64.tgz
SSHCOMMAND_URL ?= https://github.com/dokku/sshcommand/releases/download/v${SSHCOMMAND_VERSION}/sshcommand_${SSHCOMMAND_VERSION}_linux_x86_64.tgz
STACK_URL ?= https://github.com/gliderlabs/herokuish.git
PREBUILT_STACK_URL ?= gliderlabs/herokuish:latest-20
DOKKU_LIB_ROOT ?= /var/lib/dokku
PLUGINS_PATH ?= ${DOKKU_LIB_ROOT}/plugins
CORE_PLUGINS_PATH ?= ${DOKKU_LIB_ROOT}/core-plugins
PLUGIN_MAKE_TARGET ?= build-in-docker

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

include common.mk

.PHONY: all apt-update install version copyfiles copyplugin man-db plugins dependencies docker-image-labeler netrc sshcommand procfile-util plugn docker aufs stack count vagrant-acl-add vagrant-dokku go-build

include tests.mk
include package.mk
include deb.mk
include rpm.mk
include arch.mk

all:
	# Type "make install" to install.

install: dependencies version copyfiles plugin-dependencies plugins


packer:
	packer build contrib/packer.json

go-build:
	basedir=$(PWD); \
	for dir in plugins/*; do \
		if [ -e $$dir/Makefile ]; then \
			$(MAKE) -e -C $$dir $(PLUGIN_MAKE_TARGET) || exit $$? ;\
		fi ;\
	done


go-build-plugin:
ifndef PLUGIN_NAME
	$(error PLUGIN_NAME not specified)
endif
	if [ -e plugins/$(PLUGIN_NAME)/Makefile ]; then \
		$(MAKE) -e -C plugins/$(PLUGIN_NAME) $(PLUGIN_MAKE_TARGET) || exit $$? ;\
	fi

go-clean:
	basedir=$(PWD); \
	for dir in plugins/*; do \
		if [ -e $$dir/Makefile ]; then \
			$(MAKE) -e -C $$dir clean ;\
		fi ;\
	done

copyfiles:
	$(MAKE) go-build || exit 1
	cp dokku /usr/local/bin/dokku
	mkdir -p ${CORE_PLUGINS_PATH} ${PLUGINS_PATH}
	rm -rf ${CORE_PLUGINS_PATH}/*
	test -d ${CORE_PLUGINS_PATH}/enabled || PLUGIN_PATH=${CORE_PLUGINS_PATH} plugn init
	test -d ${PLUGINS_PATH}/enabled || PLUGIN_PATH=${PLUGINS_PATH} plugn init
	find plugins/ -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | while read plugin; do $(MAKE) copyplugin PLUGIN_NAME=$$plugin; done
ifndef SKIP_GO_CLEAN
	$(MAKE) go-clean
endif
	chown dokku:dokku -R ${PLUGINS_PATH} ${CORE_PLUGINS_PATH} || true
	$(MAKE) addman

copyplugin:
ifndef PLUGIN_NAME
	$(error PLUGIN_NAME not specified)
endif
	rm -Rf ${CORE_PLUGINS_PATH}/available/$(PLUGIN_NAME) && \
		rm -Rf ${PLUGINS_PATH}/available/$(PLUGIN_NAME) && \
		rm -rf ${CORE_PLUGINS_PATH}/$(PLUGIN_NAME) && \
		rm -rf ${PLUGINS_PATH}/$(PLUGIN_NAME) && \
		cp -R plugins/$(PLUGIN_NAME) ${CORE_PLUGINS_PATH}/available && \
		rm -rf ${CORE_PLUGINS_PATH}/available/$(PLUGIN_NAME)/src && \
		ln -s ${CORE_PLUGINS_PATH}/available/$(PLUGIN_NAME) ${PLUGINS_PATH}/available; \
		find /var/lib/dokku/ -xtype l -delete;\
		PLUGIN_PATH=${CORE_PLUGINS_PATH} plugn enable $(PLUGIN_NAME) ;\
		PLUGIN_PATH=${PLUGINS_PATH} plugn enable $(PLUGIN_NAME)
	chown dokku:dokku -R ${PLUGINS_PATH} ${CORE_PLUGINS_PATH} || true

addman: help2man man-db
	mkdir -p /usr/local/share/man/man1
ifneq ("$(wildcard /usr/local/share/man/man1/dokku.1-generated)","")
	cp /usr/local/share/man/man1/dokku.1-generated /usr/local/share/man/man1/dokku.1
else
	help2man -Nh help -v version -n "configure and get information from your dokku installation" -o /usr/local/share/man/man1/dokku.1 dokku
endif
	mandb

version:
	mkdir -p ${DOKKU_LIB_ROOT}
ifeq ($(DOKKU_VERSION),master)
	git describe --tags > ${DOKKU_LIB_ROOT}/VERSION  2>/dev/null || echo '~${DOKKU_VERSION} ($(shell date -uIminutes))' > ${DOKKU_LIB_ROOT}/VERSION
else
	echo $(DOKKU_VERSION) > ${DOKKU_LIB_ROOT}/STABLE_VERSION
endif

plugin-dependencies: plugn procfile-util
	sudo -E dokku plugin:install-dependencies --core

plugins: plugn procfile-util docker
	sudo -E dokku plugin:install --core

dependencies: apt-update docker-image-labeler netrc sshcommand plugn procfile-util docker help2man man-db sigil dos2unix jq parallel
	$(MAKE) -e stack

apt-update:
	apt-get update -qq

parallel:
	apt-get -qq -y --no-install-recommends install parallel

jq:
	apt-get -qq -y --no-install-recommends install jq

dos2unix:
	apt-get -qq -y --no-install-recommends install dos2unix

help2man:
	apt-get -qq -y --no-install-recommends install help2man

man-db:
	apt-get -qq -y --no-install-recommends install man-db

docker-image-labeler:
	wget -qO /tmp/docker-image-labeler_latest.tgz ${DOCKER_IMAGE_LABELER_URL}
	tar xzf /tmp/docker-image-labeler_latest.tgz -C /usr/local/bin

netrc:
	wget -qO /tmp/netrc_latest.tgz ${NETRC_URL}
	tar xzf /tmp/netrc_latest.tgz -C /usr/local/bin

procfile-util:
	wget -qO /tmp/procfile-util_latest.tgz ${PROCFILE_UTIL_URL}
	tar xzf /tmp/procfile-util_latest.tgz -C /usr/local/bin

plugn:
	wget -qO /tmp/plugn_latest.tgz ${PLUGN_URL}
	tar xzf /tmp/plugn_latest.tgz -C /usr/local/bin

sigil:
	wget -qO /tmp/sigil_latest.tgz ${SIGIL_URL}
	tar xzf /tmp/sigil_latest.tgz -C /usr/local/bin

sshcommand:
	wget -qO /tmp/sshcommand_latest.tgz ${SSHCOMMAND_URL}
	tar xzf /tmp/sshcommand_latest.tgz -C /usr/local/bin
	sshcommand create dokku /usr/local/bin/dokku

docker:
	apt-get -qq -y --no-install-recommends install curl
	grep -i -E "^docker" /etc/group || groupadd docker
	usermod -aG docker dokku
ifndef CI
	wget -nv -O - https://get.docker.com/ | sh
ifdef DOCKER_VERSION
	apt-get -qq -y --no-install-recommends install docker-engine=${DOCKER_VERSION} || (apt-cache madison docker-engine ; exit 1)
endif
	sleep 2 # give docker a moment i guess
endif

stack:
ifeq ($(shell test -e /var/run/docker.sock && touch -c /var/run/docker.sock && echo $$?),0)
ifdef BUILD_STACK
	@echo "Start building herokuish from source"
	docker images | grep gliderlabs/herokuish || (git clone ${STACK_URL} /tmp/herokuish && cd /tmp/herokuish && IMAGE_NAME=gliderlabs/herokuish BUILD_TAG=latest VERSION=master make -e ${BUILD_STACK_TARGETS} && rm -rf /tmp/herokuish)
else
ifeq ($(shell echo ${PREBUILT_STACK_URL} | grep -q -E 'http.*://|file://' && echo $$?),0)
	@echo "Start importing herokuish from ${PREBUILT_STACK_URL}"
	docker images | grep gliderlabs/herokuish || wget -nv -O - ${PREBUILT_STACK_URL} | gunzip -cd | docker import - gliderlabs/herokuish
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

vagrant-acl-add:
	vagrant ssh -- sudo sshcommand acl-add dokku $(USER)

vagrant-dokku:
	vagrant ssh -- "sudo -H -u root bash -c 'dokku $(RUN_ARGS)'"
