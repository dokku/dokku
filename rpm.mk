RPM_ARCHITECTURE = x86_64
DOKKU_RPM_PACKAGE_NAME = dokku-$(DOKKU_VERSION)-1.$(RPM_ARCHITECTURE).rpm
DOKKU_UPDATE_RPM_PACKAGE_NAME = dokku-update-$(DOKKU_UPDATE_VERSION)-1.$(RPM_ARCHITECTURE).rpm

.PHONY: rpm-all

rpm-all: rpm-setup rpm-dokku rpm-dokku-update
	mv /tmp/*.rpm .
	@echo "Done"

rpm-setup:
	@echo "-> Installing rpm build requirements"
	@sudo apt-get update -qq >/dev/null
	@sudo DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true apt-get -qq -y --no-install-recommends install gcc git build-essential wget ruby-dev ruby1.9.1 rpm >/dev/null 2>&1
	@command -v fpm >/dev/null || sudo gem install fpm --no-ri --no-rdoc
	@ssh -o StrictHostKeyChecking=no git@github.com || true

rpm-dokku: /tmp/build-dokku/var/lib/dokku/GIT_REV
	rm -f $(BUILD_DIRECTORY)/dokku_*_$(RPM_ARCHITECTURE).rpm

	cat /tmp/build-dokku/var/lib/dokku/VERSION | cut -d '-' -f 1 | cut -d 'v' -f 2 > /tmp/build-dokku/var/lib/dokku/STABLE_VERSION
ifneq (,$(findstring false,$(IS_RELEASE)))
	sed -i.bak -e "s/^/`date +%s`-/" /tmp/build-dokku/var/lib/dokku/STABLE_VERSION && rm /tmp/build-dokku/var/lib/dokku/STABLE_VERSION.bak
endif

	@echo "-> Creating rpm package"
	VERSION=$$(cat /tmp/build-dokku/var/lib/dokku/STABLE_VERSION); \
	sudo fpm -t rpm -s dir -C /tmp/build-dokku -n dokku \
		--version "$$VERSION" \
		--architecture $(RPM_ARCHITECTURE) \
		--package "$(BUILD_DIRECTORY)/$(DOKKU_RPM_PACKAGE_NAME)" \
		--depends '/usr/bin/docker' \
		--depends 'bind-utils' \
		--depends 'cpio' \
		--depends 'curl' \
		--depends 'dos2unix' \
		--depends 'docker-image-labeler >= 0.2.2' \
		--depends 'git' \
		--depends 'gliderlabs-sigil' \
		--depends 'jq' \
		--depends 'man-db' \
		--depends 'nc' \
		--depends 'nginx >= 1.8.0' \
		--depends 'plugn' \
		--depends 'procfile-util >= 0.11.0' \
		--depends '/usr/bin/python3' \
		--depends 'sshcommand >= 0.11.0' \
		--depends 'sudo' \
		--after-install rpm/dokku.postinst \
		--url "https://github.com/$(DOKKU_REPO_NAME)" \
		--description $(DOKKU_DESCRIPTION) \
		--license 'MIT License' \
		.

rpm-dokku-update:
	rm -rf $(BUILD_DIRECTORY)/$(DOKKU_UPDATE_RPM_PACKAGE_NAME)
	echo "${DOKKU_UPDATE_VERSION}" > contrib/dokku-update-version
	sudo fpm -t rpm -s dir -n dokku-update \
			 --version $(DOKKU_UPDATE_VERSION) \
			 --architecture $(RPM_ARCHITECTURE) \
			 --package $(BUILD_DIRECTORY)/$(DOKKU_UPDATE_RPM_PACKAGE_NAME) \
			 --depends 'dokku' \
			 --url "https://github.com/$(DOKKU_UPDATE_REPO_NAME)" \
			 --description $(DOKKU_UPDATE_DESCRIPTION) \
			 --license 'MIT License' \
			 contrib/dokku-update=/usr/local/bin/dokku-update \
			 contrib/dokku-update-version=/var/lib/dokku-update/VERSION
