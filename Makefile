SSHCOMMAND_URL ?= https://raw.github.com/progrium/sshcommand/master/sshcommand
PLUGINHOOK_URL ?= https://s3.amazonaws.com/progrium-pluginhook/pluginhook_0.1.0_amd64.deb
STACK_URL ?= github.com/progrium/buildstep
PREBUILT_STACK_URL ?= https://s3.amazonaws.com/progrium-dokku/progrium_buildstep_c30652f59a.tgz

all:
	# Type "make install" to install.

install: dependencies stack copyfiles plugins

copyfiles:
	cp dokku /usr/local/bin/dokku
	mkdir -p /var/lib/dokku/plugins
	cp -r plugins/* /var/lib/dokku/plugins

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
	apt-get install -y lxc-docker 
	sleep 2 # give docker a moment i guess

aufs:
	lsmod | grep aufs || modprobe aufs || apt-get install -y linux-image-extra-`uname -r`

stack:
ifdef BUILD_STACK
	@docker images | grep progrium/buildstep || docker build -t progrium/buildstep ${STACK_URL}
else
	@docker images | grep progrium/buildstep || curl ${PREBUILT_STACK_URL} | gunzip -cd | docker import - progrium/buildstep
endif

count:
	@echo "Core lines:"
	@cat receiver dokku bootstrap.sh | wc -l
	@echo "Plugin lines:"
	@find plugins -type f | xargs cat | wc -l
	@echo "Test lines:"
	@find tests -type f | xargs cat | wc -l
