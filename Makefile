DOKKU_VERSION = davis

SSHCOMMAND_URL ?= https://raw.github.com/progrium/sshcommand/master/sshcommand
PLUGINHOOK_URL ?= https://s3.amazonaws.com/progrium-pluginhook/pluginhook_0.1.0_amd64.deb
STACK_URL ?= https://github.com/lonnen/doozer.git
DOKKU_ROOT ?= /home/dokku

.PHONY: all install copyfiles version plugins dependencies sshcommand pluginhook docker aufs stack count

all:
	# Type "make install" to install.

install: dependencies stack copyfiles plugins version

copyfiles: addman
	cp dokku /usr/local/bin/dokku
	mkdir -p /var/lib/dokku/plugins
	cp -r plugins/* /var/lib/dokku/plugins

addman:
	mkdir -p /usr/local/share/man/man1
	cp dokku.1 /usr/local/share/man/man1/dokku.1
	mandb

version:
	git describe --tags > ${DOKKU_ROOT}/VERSION  2> /dev/null || echo '~${DOKKU_VERSION} ($(shell date -uIminutes))' > ${DOKKU_ROOT}/VERSION

plugins: pluginhook docker
	dokku plugins-install

dependencies: sshcommand pluginhook docker stack

sshcommand:
	wget -qO /usr/local/bin/sshcommand ${SSHCOMMAND_URL}
	chmod +x /usr/local/bin/sshcommand
	sshcommand create dokku /usr/local/bin/dokku

pluginhook:
	wget -qO /tmp/pluginhook_0.1.0_amd64.deb ${PLUGINHOOK_URL}
	dpkg -i /tmp/pluginhook_0.1.0_amd64.deb

docker: aufs
	egrep -i "^docker" /etc/group || groupadd docker
	usermod -aG docker dokku
	curl https://get.docker.io/gpg | apt-key add -
	echo deb http://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list
	apt-get update
ifdef DOCKER_VERSION
	apt-get install -y lxc-docker-${DOCKER_VERSION}
else
	apt-get install -y lxc-docker
endif
	sleep 2 # give docker a moment i guess

aufs:
	lsmod | grep aufs || modprobe aufs || apt-get install -y linux-image-extra-`uname -r`

stack:
	@docker images | grep lonnen/buildstep || (git clone ${STACK_URL} /tmp/buildstep && docker build -t lonnen/buildstep /tmp/buildstep && rm -rf /tmp/buildstep)

count:
	@echo "Core lines:"
	@cat dokku bootstrap.sh | wc -l
	@echo "Plugin lines:"
	@find plugins -type f | xargs cat | wc -l
	@echo "Test lines:"
	@find tests -type f | xargs cat | wc -l
