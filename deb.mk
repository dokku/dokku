BUILD_DIRECTORY ?= /tmp

DOKKU_DESCRIPTION = 'Docker powered PaaS that helps you build and manage the lifecycle of applications'
DOKKU_REPO_NAME ?= dokku/dokku
DOKKU_ARCHITECTURE = amd64

DOKKU_UPDATE_DESCRIPTION = 'Binary that handles updating Dokku and related systems'
DOKKU_UPDATE_REPO_NAME ?= dokku/dokku
DOKKU_UPDATE_VERSION ?= 0.2.0
DOKKU_UPDATE_ARCHITECTURE = amd64
DOKKU_UPDATE_PACKAGE_NAME = dokku-update_$(DOKKU_UPDATE_VERSION)_$(DOKKU_UPDATE_ARCHITECTURE).deb

ifndef IS_RELEASE
	IS_RELEASE = true
endif


.PHONY: install-from-deb deb-all deb-dokku deb-dokku-update deb-setup

install-from-deb:
	@echo "--> Initial apt-get update"
	sudo apt-get update -qq >/dev/null
	sudo apt-get -qq -y --no-install-recommends install apt-transport-https

	@echo "--> Installing docker"
	wget -nv -O - https://get.docker.com/ | sh

	@echo "--> Installing dokku"
	wget -nv -O - https://packagecloud.io/dokku/dokku/gpgkey | apt-key add -
	@echo "deb https://packagecloud.io/dokku/dokku/ubuntu/ $(shell lsb_release -cs 2>/dev/null || echo "bionic") main" | sudo tee /etc/apt/sources.list.d/dokku.list
	sudo apt-get update -qq >/dev/null
	sudo DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true apt-get -qq -y --no-install-recommends install dokku

deb-all: deb-setup deb-dokku deb-dokku-update
	mv $(BUILD_DIRECTORY)/*.deb .
	@echo "Done"

deb-setup:
	@echo "-> Updating deb repository and installing build requirements"
	@sudo apt-get update -qq >/dev/null
	@sudo DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true apt-get -qq -y --no-install-recommends install gcc git build-essential wget ruby-dev ruby1.9.1 lintian >/dev/null 2>&1
	@command -v fpm >/dev/null || sudo gem install fpm --no-ri --no-rdoc
	@ssh -o StrictHostKeyChecking=no git@github.com || true

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
