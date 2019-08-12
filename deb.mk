BUILD_DIRECTORY ?= /tmp

HEROKUISH_DESCRIPTION = 'Herokuish uses Docker and Buildpacks to build applications like Heroku'
HEROKUISH_REPO_NAME ?= gliderlabs/herokuish
HEROKUISH_VERSION ?= 0.5.2
HEROKUISH_ARCHITECTURE = amd64
HEROKUISH_PACKAGE_NAME = herokuish_$(HEROKUISH_VERSION)_$(HEROKUISH_ARCHITECTURE).deb

DOKKU_DESCRIPTION = 'Docker powered PaaS that helps you build and manage the lifecycle of applications'
DOKKU_REPO_NAME ?= dokku/dokku
DOKKU_ARCHITECTURE = amd64

DOKKU_UPDATE_DESCRIPTION = 'Binary that handles updating Dokku and related systems'
DOKKU_UPDATE_REPO_NAME ?= dokku/dokku
DOKKU_UPDATE_VERSION ?= 0.1.0
DOKKU_UPDATE_ARCHITECTURE = amd64
DOKKU_UPDATE_PACKAGE_NAME = dokku-update_$(DOKKU_UPDATE_VERSION)_$(DOKKU_UPDATE_ARCHITECTURE).deb

define SIGIL_DESCRIPTION
Standalone string interpolator and template processor
Sigil is a command line tool for template processing
and POSIX-compliant variable expansion. It was created
for configuration templating, but can be used for any
text processing.
endef
SIGIL_REPO_NAME ?= gliderlabs/sigil
SIGIL_VERSION ?= 0.4.0
SIGIL_ARCHITECTURE = amd64
SIGIL_PACKAGE_NAME = gliderlabs_sigil_$(SIGIL_VERSION)_$(SIGIL_ARCHITECTURE).deb
SIGIL_URL = https://github.com/gliderlabs/sigil/releases/download/v$(SIGIL_VERSION)/sigil_$(SIGIL_VERSION)_Linux_x86_64.tgz

ifndef IS_RELEASE
	IS_RELEASE = true
endif

export SIGIL_DESCRIPTION

.PHONY: install-from-deb deb-all deb-herokuish deb-dokku deb-dokku-update deb-setup deb-sigil

install-from-deb:
	@echo "--> Initial apt-get update"
	sudo apt-get update -qq >/dev/null
	sudo apt-get install -qq -y apt-transport-https

	@echo "--> Installing docker"
	wget -nv -O - https://get.docker.com/ | sh

	@echo "--> Installing dokku"
	wget -nv -O - https://packagecloud.io/dokku/dokku/gpgkey | apt-key add -
	@echo "deb https://packagecloud.io/dokku/dokku/ubuntu/ $(shell lsb_release -cs 2>/dev/null || echo "trusty") main" | sudo tee /etc/apt/sources.list.d/dokku.list
	sudo apt-get update -qq >/dev/null
	sudo DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true apt-get install -yy dokku

deb-all: deb-setup deb-herokuish deb-dokku deb-sigil deb-dokku-update
	mv $(BUILD_DIRECTORY)/*.deb .
	@echo "Done"

deb-setup:
	@echo "-> Updating deb repository and installing build requirements"
	@sudo apt-get update -qq >/dev/null
	@sudo DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true apt-get install -qq -y gcc git build-essential wget ruby-dev ruby1.9.1 lintian >/dev/null 2>&1
	@command -v fpm >/dev/null || sudo gem install fpm --no-ri --no-rdoc
	@ssh -o StrictHostKeyChecking=no git@github.com || true

deb-herokuish:
	rm -rf /tmp/tmp /tmp/build $(HEROKUISH_PACKAGE_NAME)
	mkdir -p /tmp/tmp /tmp/build

	@echo "-> Creating deb files"
	@echo "#!/usr/bin/env bash" >> /tmp/tmp/post-install
	@echo "sleep 5" >> /tmp/tmp/post-install
	@echo "echo 'Importing herokuish into docker (around 5 minutes)'" >> /tmp/tmp/post-install
	@echo 'if [[ -n $${http_proxy+x} ]]; then echo "See the docker pull docs for proxy configuration"; fi' >> /tmp/tmp/post-install
	@echo 'if [[ -n $${https_proxy+x} ]]; then echo "See the docker pull docs for proxy configuration"; fi' >> /tmp/tmp/post-install
	@echo 'if [[ -n $${BUILDARGS+x} ]]; then echo "See the docker pull docs for proxy configuration"; fi' >> /tmp/tmp/post-install
	@echo "sudo docker pull gliderlabs/herokuish:v${HEROKUISH_VERSION} && sudo docker tag gliderlabs/herokuish:v${HEROKUISH_VERSION} gliderlabs/herokuish:latest" >> /tmp/tmp/post-install

	@echo "-> Creating $(HEROKUISH_PACKAGE_NAME)"
	sudo fpm -t deb -s dir -C /tmp/build -n herokuish \
		--version $(HEROKUISH_VERSION) \
		--architecture $(HEROKUISH_ARCHITECTURE) \
		--package $(BUILD_DIRECTORY)/$(HEROKUISH_PACKAGE_NAME) \
		--deb-pre-depends 'docker-engine-cs (>= 1.9.1) | docker-engine (>= 1.9.1) | docker-ce | docker-ee' \
		--deb-pre-depends sudo \
		--after-install /tmp/tmp/post-install \
		--url "https://github.com/$(HEROKUISH_REPO_NAME)" \
		--description $(HEROKUISH_DESCRIPTION) \
		--license 'MIT License' \
		.

deb-dokku: /tmp/build-dokku/var/lib/dokku/GIT_REV
	rm -f $(BUILD_DIRECTORY)/dokku_*_$(DOKKU_ARCHITECTURE).deb

	cat /tmp/build-dokku/var/lib/dokku/VERSION | cut -d '-' -f 1 | cut -d 'v' -f 2 > /tmp/build-dokku/var/lib/dokku/STABLE_VERSION
ifneq (,$(findstring false,$(IS_RELEASE)))
	sed -i.bak -e "s/^/`date +%s`:/" /tmp/build-dokku/var/lib/dokku/STABLE_VERSION && rm /tmp/build-dokku/var/lib/dokku/STABLE_VERSION.bak
endif

	cp -r debian /tmp/build-dokku/DEBIAN
	rm -f /tmp/build-dokku/DEBIAN/lintian-overrides
	cp debian/lintian-overrides /tmp/build-dokku/usr/share/lintian/overrides/dokku
	sed -i.bak "s/^Version: .*/Version: `cat /tmp/build-dokku/var/lib/dokku/STABLE_VERSION`/g" /tmp/build-dokku/DEBIAN/control && rm /tmp/build-dokku/DEBIAN/control.bak
	dpkg-deb --build /tmp/build-dokku "$(BUILD_DIRECTORY)/dokku_`cat /tmp/build-dokku/var/lib/dokku/VERSION`_$(DOKKU_ARCHITECTURE).deb"
	lintian "$(BUILD_DIRECTORY)/dokku_`cat /tmp/build-dokku/var/lib/dokku/VERSION`_$(DOKKU_ARCHITECTURE).deb"

deb-dokku-update:
	rm -rf /tmp/dokku-update*.deb dokku-update*.deb
	echo "${DOKKU_UPDATE_VERSION}" > contrib/dokku-update-version
	sudo fpm -t deb -s dir -n dokku-update \
			 --version $(DOKKU_UPDATE_VERSION) \
			 --architecture $(DOKKU_UPDATE_ARCHITECTURE) \
			 --package $(BUILD_DIRECTORY)/$(DOKKU_UPDATE_PACKAGE_NAME) \
			 --depends 'dokku' \
			 --url "https://github.com/$(DOKKU_UPDATE_REPO_NAME)" \
			 --description $(DOKKU_UPDATE_DESCRIPTION) \
			 --license 'MIT License' \
			 contrib/dokku-update=/usr/local/bin/dokku-update \
			 contrib/dokku-update-version=/var/lib/dokku-update/VERSION

deb-sigil:
	rm -rf /tmp/tmp /tmp/build $(SIGIL_PACKAGE_NAME)
	mkdir -p /tmp/tmp /tmp/build /tmp/build/usr/bin

	@echo "-> Downloading package"
	wget -q -O /tmp/tmp/sigil-$(SIGIL_VERSION).tgz $(SIGIL_URL)
	cd /tmp/tmp/ && tar zxf /tmp/tmp/sigil-$(SIGIL_VERSION).tgz

	@echo "-> Copying files into place"
	cp /tmp/tmp/sigil /tmp/build/usr/bin/sigil && chmod +x /tmp/build/usr/bin/sigil

	@echo "-> Creating $(SIGIL_PACKAGE_NAME)"
	sudo fpm -t deb -s dir -C /tmp/build -n gliderlabs-sigil \
			 --version $(SIGIL_VERSION) \
			 --architecture $(SIGIL_ARCHITECTURE) \
			 --package $(BUILD_DIRECTORY)/$(SIGIL_PACKAGE_NAME) \
			 --url "https://github.com/$(SIGIL_REPO_NAME)" \
			 --maintainer "Jose Diaz-Gonzalez <dokku@josediazgonzalez.com>" \
			 --category utils \
			 --description "$$SIGIL_DESCRIPTION" \
			 --license 'MIT License' \
			 .
