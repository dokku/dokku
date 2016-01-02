HEROKUISH_DESCRIPTION = 'Herokuish uses Docker and Buildpacks to build applications like Heroku'
HEROKUISH_REPO_NAME ?= gliderlabs/herokuish
HEROKUISH_VERSION ?= 0.3.6
HEROKUISH_ARCHITECTURE = amd64
HEROKUISH_PACKAGE_NAME = herokuish_$(HEROKUISH_VERSION)_$(HEROKUISH_ARCHITECTURE).deb

DOKKU_DESCRIPTION = 'Docker powered mini-Heroku in around 100 lines of Bash'
DOKKU_REPO_NAME ?= dokku/dokku
DOKKU_ARCHITECTURE = amd64

PLUGN_DESCRIPTION = 'Hook system that lets users extend your application with plugins'
PLUGN_REPO_NAME ?= dokku/plugn
PLUGN_VERSION ?= 0.2.1
PLUGN_ARCHITECTURE = amd64
PLUGN_PACKAGE_NAME = plugn_$(PLUGN_VERSION)_$(PLUGN_ARCHITECTURE).deb

SSHCOMMAND_DESCRIPTION = 'Turn SSH into a thin client specifically for your app'
SSHCOMMAND_REPO_NAME ?= dokku/sshcommand
SSHCOMMAND_VERSION ?= 0.1.0
SSHCOMMAND_ARCHITECTURE = amd64
SSHCOMMAND_PACKAGE_NAME = sshcommand_$(SSHCOMMAND_VERSION)_$(SSHCOMMAND_ARCHITECTURE).deb

GOROOT = /usr/lib/go
GOBIN = /usr/bin/go
GOPATH = /home/vagrant/gocode

.PHONY: install-from-deb deb-all deb-herokuish deb-dokku deb-plugn deb-setup deb-sshcommand

install-from-deb:
	echo "--> Initial apt-get update"
	sudo apt-get update -qq > /dev/null
	sudo apt-get install -qq -y apt-transport-https

	echo "--> Installing docker"
	wget -nv -O - https://get.docker.com/ | sh

	echo "--> Installing dokku"
	wget -nv -O - https://packagecloud.io/gpg.key | apt-key add -
	echo "deb https://packagecloud.io/dokku/dokku/ubuntu/ trusty main" | sudo tee /etc/apt/sources.list.d/dokku.list
	sudo apt-get update -qq > /dev/null
	sudo apt-get install dokku

deb-all: deb-herokuish deb-dokku deb-plugn deb-sshcommand
	mv /tmp/*.deb .
	echo "Done"

deb-setup:
	echo "-> Updating deb repository and installing build requirements"
	sudo apt-get update -qq > /dev/null
	sudo apt-get install -qq -y gcc git ruby-dev ruby1.9.1 > /dev/null 2>&1
	command -v fpm > /dev/null || sudo gem install fpm --no-ri --no-rdoc
	ssh -o StrictHostKeyChecking=no git@github.com || true

deb-herokuish: deb-setup
	rm -rf /tmp/tmp /tmp/build $(HEROKUISH_PACKAGE_NAME)
	mkdir -p /tmp/tmp /tmp/build

	echo "-> Creating deb files"
	echo "#!/usr/bin/env bash" >> /tmp/tmp/post-install
	echo "sleep 5" >> /tmp/tmp/post-install
	echo "count=\`sudo docker images | grep gliderlabs/herokuish | wc -l\`" >> /tmp/tmp/post-install
	echo 'if [ "$$count" -ne 0 ]; then' >> /tmp/tmp/post-install
	echo "  echo 'Removing old herokuish image'" >> /tmp/tmp/post-install
	echo "  sudo docker rmi gliderlabs/herokuish" >> /tmp/tmp/post-install
	echo "fi" >> /tmp/tmp/post-install
	echo "echo 'Importing herokuish into docker (around 5 minutes)'" >> /tmp/tmp/post-install
	echo "sudo docker build -t gliderlabs/herokuish /var/lib/herokuish 1> /dev/null" >> /tmp/tmp/post-install

	echo "-> Cloning repository"
	git clone -q "https://github.com/$(HEROKUISH_REPO_NAME).git" /tmp/tmp/herokuish > /dev/null
	rm -rf /tmp/tmp/herokuish/.git /tmp/tmp/herokuish/.gitignore

	echo "-> Copying files into place"
	mkdir -p "/tmp/build/var/lib"
	cp -rf /tmp/tmp/herokuish /tmp/build/var/lib/herokuish

	echo "-> Creating $(HEROKUISH_PACKAGE_NAME)"
	sudo fpm -t deb -s dir -C /tmp/build -n herokuish -v $(HEROKUISH_VERSION) -a $(HEROKUISH_ARCHITECTURE) -p $(HEROKUISH_PACKAGE_NAME) --deb-pre-depends 'docker-engine | docker-engine-cs' --deb-pre-depends sudo --after-install /tmp/tmp/post-install --url "https://github.com/$(HEROKUISH_REPO_NAME)" --description $(HEROKUISH_DESCRIPTION) --license 'MIT License' .
	mv *.deb /tmp

deb-dokku: deb-setup
	rm -rf /tmp/tmp /tmp/build dokku_*_$(DOKKU_ARCHITECTURE).deb
	mkdir -p /tmp/tmp /tmp/build

	cp -r debian /tmp/build/DEBIAN
	mkdir -p /tmp/build/usr/bin
	mkdir -p /tmp/build/var/lib/dokku/core-plugins/available
	mkdir -p /tmp/build/usr/share/man/man1
	mkdir -p /tmp/build/usr/share/dokku/contrib

	cp dokku /tmp/build/usr/bin
	cp -r plugins/* /tmp/build/var/lib/dokku/core-plugins/available
	find plugins/ -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | while read plugin; do touch /tmp/build/var/lib/dokku/core-plugins/available/$$plugin/.core; done
	$(MAKE) help2man
	$(MAKE) addman
	cp /usr/local/share/man/man1/dokku.1 /tmp/build/usr/share/man/man1/dokku.1
	gzip -9 /tmp/build/usr/share/man/man1/dokku.1
	cp contrib/dokku-installer.py /tmp/build/usr/share/dokku/contrib
	git describe --tags > /tmp/build/var/lib/dokku/VERSION
	cat /tmp/build/var/lib/dokku/VERSION | cut -d '-' -f 1 | cut -d 'v' -f 2 > /tmp/build/var/lib/dokku/STABLE_VERSION
	git rev-parse HEAD > /tmp/build/var/lib/dokku/GIT_REV
	sed -i "s/^Version: .*/Version: `cat /tmp/build/var/lib/dokku/STABLE_VERSION`/g" /tmp/build/DEBIAN/control
	dpkg-deb --build /tmp/build "/vagrant/dokku_`cat /tmp/build/var/lib/dokku/STABLE_VERSION`_$(DOKKU_ARCHITECTURE).deb"
	mv *.deb /tmp

deb-plugn: deb-setup
	rm -rf /tmp/tmp /tmp/build $(PLUGN_PACKAGE_NAME)
	mkdir -p /tmp/tmp /tmp/build

	echo "-> Cloning repository"
	git clone -q "https://github.com/$(PLUGN_REPO_NAME).git" /tmp/tmp/plugn > /dev/null
	rm -rf /tmp/tmp/plugn/.git /tmp/tmp/plugn/.gitignore

	echo "-> Copying files into place"
	mkdir -p /tmp/build/usr/local/bin $(GOPATH)
	sudo apt-get clean
	sudo apt-get update -qq > /dev/null
	sudo apt-get install -qq -y git golang mercurial > /dev/null 2>&1
	export PATH=$(PATH):$(GOROOT)/bin:$(GOPATH)/bin && export GOROOT=$(GOROOT) && export GOPATH=$(GOPATH):/tmp/tmp/plugn && go get github.com/dokku/plugn
	export PATH=$(PATH):$(GOROOT)/bin:$(GOPATH)/bin && export GOROOT=$(GOROOT) && export GOPATH=$(GOPATH):/tmp/tmp/plugn && cd /home/vagrant/gocode/src/github.com/dokku/plugn && make deps
	export PATH=$(PATH):$(GOROOT)/bin:$(GOPATH)/bin && export GOROOT=$(GOROOT) && export GOPATH=$(GOPATH):/tmp/tmp/plugn && cd /home/vagrant/gocode/src/github.com/dokku/plugn && rm plugn && go build -o plugn
	mv /home/vagrant/gocode/src/github.com/dokku/plugn/plugn /tmp/build/usr/local/bin/plugn

	echo "-> Creating $(PLUGN_PACKAGE_NAME)"
	sudo fpm -t deb -s dir -C /tmp/build -n plugn -v $(PLUGN_VERSION) -a $(PLUGN_ARCHITECTURE) -p $(PLUGN_PACKAGE_NAME) --url "https://github.com/$(PLUGN_REPO_NAME)" --description $(PLUGN_DESCRIPTION) --license 'MIT License' .
	mv *.deb /tmp

deb-sshcommand: deb-setup
	rm -rf /tmp/tmp /tmp/build $(SSHCOMMAND_PACKAGE_NAME)
	mkdir -p /tmp/tmp /tmp/build

	echo "-> Cloning repository"
	git clone -q "https://github.com/$(SSHCOMMAND_REPO_NAME).git" /tmp/tmp/sshcommand > /dev/null
	rm -rf /tmp/tmp/sshcommand/.git /tmp/tmp/sshcommand/.gitignore

	echo "-> Copying files into place"
	mkdir -p "/tmp/build/usr/local/bin"
	cp /tmp/tmp/sshcommand/sshcommand /tmp/build/usr/local/bin/sshcommand
	chmod +x /tmp/build/usr/local/bin/sshcommand

	echo "-> Creating $(SSHCOMMAND_PACKAGE_NAME)"
	sudo fpm -t deb -s dir -C /tmp/build -n sshcommand -v $(SSHCOMMAND_VERSION) -a $(SSHCOMMAND_ARCHITECTURE) -p $(SSHCOMMAND_PACKAGE_NAME) --url "https://github.com/$(SSHCOMMAND_REPO_NAME)" --description $(SSHCOMMAND_DESCRIPTION) --license 'MIT License' .
	mv *.deb /tmp
