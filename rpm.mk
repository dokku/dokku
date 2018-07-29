RPM_ARCHITECTURE = x86_64
DOKKU_UPDATE_RPM_PACKAGE_NAME = dokku-update-$(DOKKU_UPDATE_VERSION)-1.$(RPM_ARCHITECTURE).rpm
HEROKUISH_RPM_PACKAGE_NAME = herokuish-$(HEROKUISH_VERSION)-1.$(RPM_ARCHITECTURE).rpm
PLUGN_RPM_PACKAGE_NAME = plugn-$(PLUGN_VERSION)-1.$(RPM_ARCHITECTURE).rpm
SSHCOMMAND_RPM_PACKAGE_NAME = sshcommand-$(SSHCOMMAND_VERSION)-1.$(RPM_ARCHITECTURE).rpm
SIGIL_RPM_PACKAGE_NAME = gliderlabs-sigil-$(SIGIL_VERSION)-1.$(RPM_ARCHITECTURE).rpm

.PHONY: rpm-all

rpm-all: rpm-setup rpm-herokuish rpm-dokku rpm-plugn rpm-sshcommand rpm-sigil rpm-dokku-update
	mv /tmp/*.rpm .
	@echo "Done"

rpm-setup:
	@echo "-> Installing rpm build requirements"
	@sudo apt-get update -qq > /dev/null
	@sudo DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true apt-get install -qq -y gcc git build-essential wget ruby-dev ruby1.9.1 rpm > /dev/null 2>&1
	@command -v fpm > /dev/null || sudo gem install fpm --no-ri --no-rdoc
	@ssh -o StrictHostKeyChecking=no git@github.com || true

rpm-herokuish:
	rm -rf /tmp/tmp /tmp/build $(HEROKUISH_RPM_PACKAGE_NAME)
	mkdir -p /tmp/tmp /tmp/build

	@echo "-> Creating rpm files"
	@echo "#!/usr/bin/env bash" >> /tmp/tmp/post-install
	@echo 'echo "Starting docker"' >> /tmp/tmp/post-install
	@echo 'systemctl start docker' >> /tmp/tmp/post-install
	@echo "sleep 5" >> /tmp/tmp/post-install
	@echo "echo 'Importing herokuish into docker (around 5 minutes)'" >> /tmp/tmp/post-install
	@echo 'if [[ ! -z $${http_proxy+x} ]]; then echo "See the docker pull docs for proxy configuration"; fi' >> /tmp/tmp/post-install
	@echo 'if [[ ! -z $${https_proxy+x} ]]; then echo "See the docker pull docs for proxy configuration"; fi' >> /tmp/tmp/post-install
	@echo 'if [[ ! -z $${BUILDARGS+x} ]]; then echo "See the docker pull docs for proxy configuration"; fi' >> /tmp/tmp/post-install
	@echo "sudo docker pull gliderlabs/herokuish:v${HEROKUISH_VERSION} && sudo docker tag gliderlabs/herokuish:v${HEROKUISH_VERSION} gliderlabs/herokuish:latest" >> /tmp/tmp/post-install

	@echo "-> Creating $(HEROKUISH_RPM_PACKAGE_NAME)"
	sudo fpm -t rpm -s dir -C /tmp/build -n herokuish \
		-v $(HEROKUISH_VERSION) \
		-a $(RPM_ARCHITECTURE) \
		-p $(HEROKUISH_RPM_PACKAGE_NAME) \
		--depends '/usr/bin/docker' \
		--depends 'sudo' \
		--after-install /tmp/tmp/post-install \
		--url "https://github.com/$(HEROKUISH_REPO_NAME)" \
		--description $(HEROKUISH_DESCRIPTION) \
		--license 'MIT License' \
		.
	mv *.rpm /tmp

rpm-dokku:
	rm -rf /tmp/tmp /tmp/build dokku_*_$(RPM_ARCHITECTURE).rpm
	mkdir -p /tmp/tmp /tmp/build

	mkdir -p /tmp/build/usr/share/bash-completion/completions
	mkdir -p /tmp/build/usr/bin
	mkdir -p /tmp/build/usr/share/doc/dokku
	mkdir -p /tmp/build/usr/share/dokku/contrib
	mkdir -p /tmp/build/usr/share/lintian/overrides
	mkdir -p /tmp/build/usr/share/man/man1
	mkdir -p /tmp/build/var/lib/dokku/core-plugins/available

	cp dokku /tmp/build/usr/bin
	cp LICENSE /tmp/build/usr/share/doc/dokku/copyright
	cp contrib/bash-completion /tmp/build/usr/share/bash-completion/completions/dokku
	find . -name ".DS_Store" -depth -exec rm {} \;
	$(MAKE) go-build
	cp common.mk /tmp/build/var/lib/dokku/core-plugins/common.mk
	cp -r plugins/* /tmp/build/var/lib/dokku/core-plugins/available
	find plugins/ -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | while read plugin; do cd /tmp/build/var/lib/dokku/core-plugins/available/$$plugin && if [ -e Makefile ]; then $(MAKE) src-clean; fi; done
	find plugins/ -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | while read plugin; do touch /tmp/build/var/lib/dokku/core-plugins/available/$$plugin/.core; done
	rm /tmp/build/var/lib/dokku/core-plugins/common.mk
	$(MAKE) help2man
	$(MAKE) addman
	cp /usr/local/share/man/man1/dokku.1 /tmp/build/usr/share/man/man1/dokku.1
	gzip -9 /tmp/build/usr/share/man/man1/dokku.1
	cp contrib/dokku-installer.py /tmp/build/usr/share/dokku/contrib
ifeq ($(DOKKU_VERSION),master)
	git describe --tags > /tmp/build/var/lib/dokku/VERSION
else
	echo $(DOKKU_VERSION) > /tmp/build/var/lib/dokku/VERSION
endif
	cat /tmp/build/var/lib/dokku/VERSION | cut -d '-' -f 1 | cut -d 'v' -f 2 > /tmp/build/var/lib/dokku/STABLE_VERSION
ifneq (,$(findstring false,$(IS_RELEASE)))
	sed -i.bak -e "s/^/`date +%s`-/" /tmp/build/var/lib/dokku/STABLE_VERSION && rm /tmp/build/var/lib/dokku/STABLE_VERSION.bak
endif
ifdef DOKKU_GIT_REV
	echo "$(DOKKU_GIT_REV)" > /tmp/build/var/lib/dokku/GIT_REV
else
	git rev-parse HEAD > /tmp/build/var/lib/dokku/GIT_REV
endif

	@echo "-> Creating rpm package"
	VERSION=$$(cat /tmp/build/var/lib/dokku/STABLE_VERSION); \
	sudo fpm -t rpm -s dir -C /tmp/build -n dokku \
		-v "$$VERSION" \
		-a $(RPM_ARCHITECTURE) \
		-p "dokku-$$VERSION-1.x86_64.rpm" \
		--depends '/usr/bin/docker' \
		--depends 'bind-utils' \
		--depends 'curl' \
		--depends 'gcc' \
		--depends 'git' \
		--depends 'gliderlabs-sigil' \
		--depends 'make' \
		--depends 'man-db' \
		--depends 'nc' \
		--depends 'nginx >= 1.8.0' \
		--depends 'plugn' \
		--depends 'procfile-util' \
		--depends 'python' \
		--depends 'sshcommand' \
		--depends 'sudo' \
		--after-install rpm/dokku.postinst \
		--url "https://github.com/$(DOKKU_REPO_NAME)" \
		--description $(DOKKU_DESCRIPTION) \
		--license 'MIT License' \
		.
	mv *.rpm "/tmp/dokku-`cat /tmp/build/var/lib/dokku/VERSION`-1.$(RPM_ARCHITECTURE).rpm"

rpm-dokku-update:
	rm -rf /tmp/dokku-update*.rpm dokku-update*.rpm
	echo "${DOKKU_UPDATE_VERSION}" > contrib/dokku-update-version
	sudo fpm -t rpm -s dir -n dokku-update \
			 --version $(DOKKU_UPDATE_VERSION) \
			 --architecture $(RPM_ARCHITECTURE) \
			 --package $(DOKKU_UPDATE_RPM_PACKAGE_NAME) \
			 --depends 'dokku' \
			 --url "https://github.com/$(DOKKU_UPDATE_REPO_NAME)" \
			 --description $(DOKKU_UPDATE_DESCRIPTION) \
			 --license 'MIT License' \
			 contrib/dokku-update=/usr/local/bin/dokku-update \
			 contrib/dokku-update-version=/var/lib/dokku-update/VERSION
	mv *.rpm /tmp

rpm-plugn:
	rm -rf /tmp/tmp /tmp/build $(PLUGN_RPM_PACKAGE_NAME)
	mkdir -p /tmp/tmp /tmp/build /tmp/build/usr/bin

	@echo "-> Downloading package"
	wget -q -O /tmp/tmp/plugn-$(PLUGN_VERSION).tgz $(PLUGN_URL)
	cd /tmp/tmp/ && tar zxf /tmp/tmp/plugn-$(PLUGN_VERSION).tgz

	@echo "-> Copying files into place"
	cp /tmp/tmp/plugn /tmp/build/usr/bin/plugn && chmod +x /tmp/build/usr/bin/plugn

	@echo "-> Creating $(PLUGN_RPM_PACKAGE_NAME)"
	sudo fpm -t rpm -s dir -C /tmp/build -n plugn \
			 --version $(PLUGN_VERSION) \
			 --architecture $(RPM_ARCHITECTURE) \
			 --package $(PLUGN_RPM_PACKAGE_NAME) \
			 --url "https://github.com/$(PLUGN_REPO_NAME)" \
			 --category utils \
			 --description "$$PLUGN_DESCRIPTION" \
			 --license 'MIT License' \
			 .
	mv *.rpm /tmp

rpm-sshcommand:
	rm -rf /tmp/tmp /tmp/build $(SSHCOMMAND_RPM_PACKAGE_NAME)
	mkdir -p /tmp/tmp /tmp/build /tmp/build/usr/bin

	@echo "-> Downloading package"
	wget -q -O /tmp/tmp/sshcommand-$(SSHCOMMAND_VERSION) $(SSHCOMMAND_URL)

	@echo "-> Copying files into place"
	mkdir -p "/tmp/build/usr/bin"
	cp /tmp/tmp/sshcommand-$(SSHCOMMAND_VERSION) /tmp/build/usr/bin/sshcommand
	chmod +x /tmp/build/usr/bin/sshcommand

	@echo "-> Creating $(SSHCOMMAND_RPM_PACKAGE_NAME)"
	sudo fpm -t rpm -s dir -C /tmp/build -n sshcommand \
			 --version $(SSHCOMMAND_VERSION) \
			 -a $(RPM_ARCHITECTURE) \
			 --package $(SSHCOMMAND_RPM_PACKAGE_NAME) \
			 --url "https://github.com/$(SSHCOMMAND_REPO_NAME)" \
			 --category admin \
			 --description "$$SSHCOMMAND_DESCRIPTION" \
			 --license 'MIT License' \
			 .
	mv *.rpm /tmp

rpm-sigil:
	rm -rf /tmp/tmp /tmp/build $(SIGIL_PACKAGE_NAME)
	mkdir -p /tmp/tmp /tmp/build /tmp/build/usr/bin

	@echo "-> Downloading package"
	wget -q -O /tmp/tmp/sigil-$(SIGIL_VERSION).tgz $(SIGIL_URL)
	cd /tmp/tmp/ && tar zxf /tmp/tmp/sigil-$(SIGIL_VERSION).tgz

	@echo "-> Copying files into place"
	cp /tmp/tmp/sigil /tmp/build/usr/bin/sigil && chmod +x /tmp/build/usr/bin/sigil

	@echo "-> Creating $(SIGIL_RPM_PACKAGE_NAME)"
	sudo fpm -t rpm -s dir -C /tmp/build -n gliderlabs-sigil \
		--version $(SIGIL_VERSION) \
		-a $(RPM_ARCHITECTURE) \
		--package $(SIGIL_RPM_PACKAGE_NAME) \
		--url "https://github.com/$(SIGIL_REPO_NAME)" \
		--category utils \
		--description "$$SIGIL_DESCRIPTION" \
		--license 'MIT License' \
		.
	mv *.rpm /tmp
