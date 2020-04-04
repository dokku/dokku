RPM_ARCHITECTURE = x86_64
DOKKU_RPM_PACKAGE_NAME = dokku-$(DOKKU_VERSION)-1.$(RPM_ARCHITECTURE).rpm
DOKKU_UPDATE_RPM_PACKAGE_NAME = dokku-update-$(DOKKU_UPDATE_VERSION)-1.$(RPM_ARCHITECTURE).rpm
HEROKUISH_RPM_PACKAGE_NAME = herokuish-$(HEROKUISH_VERSION)-1.$(RPM_ARCHITECTURE).rpm
SIGIL_RPM_PACKAGE_NAME = gliderlabs-sigil-$(SIGIL_VERSION)-1.$(RPM_ARCHITECTURE).rpm

.PHONY: rpm-all

rpm-all: rpm-setup rpm-herokuish rpm-dokku rpm-sigil rpm-dokku-update
	mv /tmp/*.rpm .
	@echo "Done"

rpm-setup:
	@echo "-> Installing rpm build requirements"
	@sudo apt-get update -qq >/dev/null
	@sudo DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true apt-get install -qq -y gcc git build-essential wget ruby-dev ruby1.9.1 rpm >/dev/null 2>&1
	@command -v fpm >/dev/null || sudo gem install fpm --no-ri --no-rdoc
	@ssh -o StrictHostKeyChecking=no git@github.com || true

rpm-herokuish:
	rm -rf /tmp/tmp /tmp/build $(BUILD_DIRECTORY)/$(HEROKUISH_RPM_PACKAGE_NAME)
	mkdir -p /tmp/tmp /tmp/build

	@echo "-> Creating rpm files"
	@echo "#!/usr/bin/env bash" >> /tmp/tmp/post-install
	@echo 'echo "Starting docker"' >> /tmp/tmp/post-install
	@echo 'systemctl start docker' >> /tmp/tmp/post-install
	@echo "sleep 5" >> /tmp/tmp/post-install
	@echo "echo 'Importing herokuish into docker (around 5 minutes)'" >> /tmp/tmp/post-install
	@echo 'if [[ -n $${http_proxy+x} ]]; then echo "See the docker pull docs for proxy configuration"; fi' >> /tmp/tmp/post-install
	@echo 'if [[ -n $${https_proxy+x} ]]; then echo "See the docker pull docs for proxy configuration"; fi' >> /tmp/tmp/post-install
	@echo 'if [[ -n $${BUILDARGS+x} ]]; then echo "See the docker pull docs for proxy configuration"; fi' >> /tmp/tmp/post-install
	@echo "sudo docker pull gliderlabs/herokuish:v${HEROKUISH_VERSION} && sudo docker tag gliderlabs/herokuish:v${HEROKUISH_VERSION} gliderlabs/herokuish:latest" >> /tmp/tmp/post-install

	@echo "-> Creating $(HEROKUISH_RPM_PACKAGE_NAME)"
	sudo fpm -t rpm -s dir -C /tmp/build -n herokuish \
		--version $(HEROKUISH_VERSION) \
		--architecture $(RPM_ARCHITECTURE) \
		--package $(BUILD_DIRECTORY)/$(HEROKUISH_RPM_PACKAGE_NAME) \
		--depends '/usr/bin/docker' \
		--depends 'sudo' \
		--after-install /tmp/tmp/post-install \
		--url "https://github.com/$(HEROKUISH_REPO_NAME)" \
		--description $(HEROKUISH_DESCRIPTION) \
		--license 'MIT License' \
		.

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
		--depends 'git' \
		--depends 'gliderlabs-sigil' \
		--depends 'jq' \
		--depends 'man-db' \
		--depends 'nc' \
		--depends 'nginx >= 1.8.0' \
		--depends 'plugn' \
		--depends 'procfile-util' \
		--depends 'python' \
		--depends 'sshcommand >= 0.10.0' \
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

rpm-sigil:
	rm -rf /tmp/tmp /tmp/build $(BUILD_DIRECTORY)/$(SIGIL_RPM_PACKAGE_NAME)
	mkdir -p /tmp/tmp /tmp/build /tmp/build/usr/bin

	@echo "-> Downloading package"
	wget -q -O /tmp/tmp/sigil-$(SIGIL_VERSION).tgz $(SIGIL_URL)
	cd /tmp/tmp/ && tar zxf /tmp/tmp/sigil-$(SIGIL_VERSION).tgz

	@echo "-> Copying files into place"
	cp /tmp/tmp/sigil /tmp/build/usr/bin/sigil && chmod +x /tmp/build/usr/bin/sigil

	@echo "-> Creating $(SIGIL_RPM_PACKAGE_NAME)"
	sudo fpm -t rpm -s dir -C /tmp/build -n gliderlabs-sigil \
		--version $(SIGIL_VERSION) \
		--architecture $(RPM_ARCHITECTURE) \
		--package $(BUILD_DIRECTORY)/$(SIGIL_RPM_PACKAGE_NAME) \
		--url "https://github.com/$(SIGIL_REPO_NAME)" \
		--category utils \
		--description "$$SIGIL_DESCRIPTION" \
		--license 'MIT License' \
		.
