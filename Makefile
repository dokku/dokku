DOKKU_VERSION ?= master

TARGETARCH ?= amd64

DOCKER_IMAGE_LABELER_URL ?= $(shell jq -r --arg name docker-image-labeler --arg arch $(TARGETARCH) '.dependencies[] | select(.name == $$name) | .urls[$$arch]' contrib/dependencies.json)
DOCKER_CONTAINER_HEALTHCHECKER_URL ?= $(shell jq -r --arg name docker-container-healthchecker  --arg arch $(TARGETARCH) '.dependencies[] | select(.name == $$name) | .urls[$$arch]' contrib/dependencies.json)
LAMBDA_BUILDER_URL ?= $(shell jq -r --arg name lambda-builder  --arg arch $(TARGETARCH) '.dependencies[] | select(.name == $$name) | .urls[$$arch]' contrib/dependencies.json)
NETRC_URL ?= $(shell jq -r --arg name netrc  --arg arch $(TARGETARCH) '.dependencies[] | select(.name == $$name) | .urls[$$arch]' contrib/dependencies.json)
PLUGN_URL ?= $(shell jq -r --arg name plugn  --arg arch $(TARGETARCH) '.predependencies[] | select(.name == $$name) | .urls[$$arch]' contrib/dependencies.json)
PROCFILE_UTIL_URL ?= $(shell jq -r --arg name procfile-util  --arg arch $(TARGETARCH) '.dependencies[] | select(.name == $$name) | .urls[$$arch]' contrib/dependencies.json)
SIGIL_URL ?= $(shell jq -r --arg name gliderlabs-sigil  --arg arch $(TARGETARCH) '.predependencies[] | select(.name == $$name) | .urls[$$arch]' contrib/dependencies.json)
SSHCOMMAND_URL ?= $(shell jq -r --arg name sshcommand  --arg arch $(TARGETARCH) '.dependencies[] | select(.name == $$name) | .urls[$$arch]' contrib/dependencies.json)
STACK_URL ?= https://github.com/gliderlabs/herokuish.git
PREBUILT_STACK_URL ?= gliderlabs/herokuish:latest-22
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

.PHONY: all apt-update install version copyfiles copyplugin man-db plugins dependencies docker-image-labeler lambda-builder netrc sshcommand procfile-util plugn docker aufs stack count vagrant-acl-add vagrant-dokku go-build

include docs.mk
include tests.mk
include package.mk
include deb.mk
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

copydokku:
	cp dokku /usr/local/bin/dokku
	chmod 0755 /usr/local/bin/dokku

copyfiles: copydokku
	$(MAKE) go-build || exit 1
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

dependencies: apt-update docker-image-labeler docker-container-healthchecker lambda-builder netrc sshcommand plugn procfile-util docker help2man man-db sigil dos2unix jq parallel
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
	wget -qO /usr/local/bin/docker-image-labeler ${DOCKER_IMAGE_LABELER_URL}
	chmod +x /usr/local/bin/docker-image-labeler

docker-container-healthchecker:
	wget -qO /usr/local/bin/docker-container-healthchecker ${DOCKER_CONTAINER_HEALTHCHECKER_URL}
	chmod +x /usr/local/bin/docker-container-healthchecker

lambda-builder:
	wget -qO /usr/local/bin/lambda-builder ${LAMBDA_BUILDER_URL}
	chmod +x /usr/local/bin/lambda-builder

netrc:
	wget -qO /usr/local/bin/netrc ${NETRC_URL}
	chmod +x /usr/local/bin/netrc

procfile-util:
	wget -qO /usr/local/bin/procfile-util ${PROCFILE_UTIL_URL}
	chmod +x /usr/local/bin/procfile-util

plugn:
	wget -qO /usr/local/bin/plugn ${PLUGN_URL}
	chmod +x /usr/local/bin/plugn

sigil:
	wget -qO /usr/local/bin/sigil ${SIGIL_URL}
	chmod +x /usr/local/bin/sigil

sshcommand:
	wget -qO /usr/local/bin/sshcommand ${SSHCOMMAND_URL}
	chmod +x /usr/local/bin/sshcommand
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
