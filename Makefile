GITRECEIVE_URL = https://raw.github.com/progrium/gitreceive/master/gitreceive
SSHCOMMAND_URL = https://raw.github.com/progrium/sshcommand/master/sshcommand
PLUGINHOOK_URL = https://s3.amazonaws.com/progrium-pluginhook/pluginhook_0.1.0_amd64.deb

all: install

install: gitreceive sshcommand pluginhook copy

copy:
	cp dokku /usr/local/bin/dokku
	cp receiver /home/git/receiver
	mkdir -p /var/lib/dokku/plugins
	cp -r plugins/* /var/lib/dokku/plugins
	PLUGIN_PATH=/var/lib/dokku/plugins pluginhook install

gitreceive:
	wget -qO /usr/local/bin/gitreceive ${GITRECEIVE_URL}
	chmod +x /usr/local/bin/gitreceive
	gitreceive init

sshcommand:
	wget -qO /usr/local/bin/sshcommand ${SSHCOMMAND_URL}
	chmod +x /usr/local/bin/sshcommand
	sshcommand create dokku /usr/local/bin/dokku

pluginhook:
	wget -qO /tmp/pluginhook_0.1.0_amd64.deb ${PLUGINHOOK_URL}
	cd /tmp && dpkg -i pluginhook_0.1.0_amd64.deb

count:
	@echo "Core lines:"
	@cat receiver dokku bootstrap.sh | wc -l
	@echo "Plugin lines:"
	@find plugins -type f | xargs cat | wc -l
	@echo "Test lines:"
	@find tests -type f | xargs cat | wc -l