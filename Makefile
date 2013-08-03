GITRECEIVE_URL ?= https://raw.github.com/progrium/gitreceive/master/gitreceive
SSHCOMMAND_URL ?= https://raw.github.com/progrium/sshcommand/master/sshcommand
PLUGINHOOK_URL ?= https://s3.amazonaws.com/progrium-pluginhook/pluginhook_0.1.0_amd64.deb
DOCKER_URL ?= https://launchpad.net/~dotcloud/+archive/lxc-docker/+files/lxc-docker_0.4.8-1_amd64.deb
DOCKER_BIN ?= https://s3.amazonaws.com/get.docker.io/builds/Linux/x86_64/docker-1004d57b85fc3714b089da4c457228690f254504
STACK_URL ?= https://s3.amazonaws.com/progrium-dokku/progrium_buildstep.tgz

all: dependencies stack install plugins

install:
	cp dokku /usr/local/bin/dokku
	cp receiver /home/git/receiver
	mkdir -p /var/lib/dokku/plugins
	cp -r plugins/* /var/lib/dokku/plugins

plugins: pluginhook docker
	dokku plugins-install

dependencies: gitreceive sshcommand pluginhook docker stack

gitreceive:
	apt-get -y install git
	wget -qO /usr/local/bin/gitreceive ${GITRECEIVE_URL}
	chmod +x /usr/local/bin/gitreceive
	test -f /home/git/receiver || gitreceive init

sshcommand:
	wget -qO /usr/local/bin/sshcommand ${SSHCOMMAND_URL}
	chmod +x /usr/local/bin/sshcommand
	sshcommand create dokku /usr/local/bin/dokku

pluginhook:
	wget -qO /tmp/pluginhook_0.1.0_amd64.deb ${PLUGINHOOK_URL}
	dpkg -i /tmp/pluginhook_0.1.0_amd64.deb

docker: aufs
	wget -qO /tmp/lxc-docker_0.4.8-1_amd64.deb ${DOCKER_URL}
	dpkg --force-depends -i /tmp/lxc-docker_0.4.8-1_amd64.deb && apt-get install -f -y
	# newer docker 0.4.8 binary
	stop docker
	wget -qO /usr/bin/docker ${DOCKER_BIN}
	chmod +x /usr/bin/docker
	start docker
	sleep 2 # give docker a moment i guess

aufs:
	lsmod | grep aufs || modprobe aufs || apt-get install -y linux-image-extra-`uname -r`

stack:
	@docker images | grep progrium/buildstep || curl ${STACK_URL} | gunzip -cd | docker import - progrium/buildstep

count:
	@echo "Core lines:"
	@cat receiver dokku bootstrap.sh | wc -l
	@echo "Plugin lines:"
	@find plugins -type f | xargs cat | wc -l
	@echo "Test lines:"
	@find tests -type f | xargs cat | wc -l
