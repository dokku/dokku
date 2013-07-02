GITRECEIVE_URL = https://raw.github.com/progrium/gitreceive/master/gitreceive
SSHCOMMAND_URL = https://raw.github.com/progrium/sshcommand/master/sshcommand
PLUGINHOOK_URL = https://s3.amazonaws.com/progrium-pluginhook/pluginhook_0.1.0_amd64.deb
DOCKER_URL = https://launchpad.net/~dotcloud/+archive/lxc-docker/+files/lxc-docker_0.4.2-1_amd64.deb
STACK_URL = https://s3.amazonaws.com/progrium-dokku/progrium_buildstep.tgz

all: dependencies stack install plugins

install:
	cp dokku /usr/local/bin/dokku
	cp receiver /home/git/receiver

plugins: pluginhook docker
	mkdir -p /var/lib/dokku/plugins
	cp -r plugins/* /var/lib/dokku/plugins
	dokku plugins-install
	dokku plugins

dependencies: gitreceive sshcommand pluginhook docker stack

gitreceive:
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
	wget -qO /tmp/lxc-docker_0.4.2-1_amd64.deb ${DOCKER_URL}
	dpkg --force-depends -i /tmp/lxc-docker_0.4.2-1_amd64.deb && apt-get install -f -y

aufs:
	modprobe aufs || apt-get install -y linux-image-extra-`uname -r`

stack:
	@docker images | grep progrium/buildstep || curl ${STACK_URL} | gunzip -cd | docker import - progrium/buildstep

count:
	@echo "Core lines:"
	@cat receiver dokku bootstrap.sh | wc -l
	@echo "Plugin lines:"
	@find plugins -type f | xargs cat | wc -l
	@echo "Test lines:"
	@find tests -type f | xargs cat | wc -l
