GITRECEIVE_URL = https://raw.github.com/progrium/gitreceive/master/gitreceive
SSHCOMMAND_URL = https://raw.github.com/progrium/sshcommand/master/sshcommand
PLUGINHOOK_URL = https://s3.amazonaws.com/progrium-pluginhook/pluginhook_0.1.0_amd64.deb

all: install

install: gitreceive sshcommand pluginhook
	cp dokku /usr/local/bin/dokku
	cp receiver /home/git/receiver
	cp nginx-app-conf /home/git/nginx-app-conf
	cp -r plugins /home/git/.plugins

gitreceive:
	wget -qO /usr/local/bin/gitreceive ${GITRECEIVE_URL}
	chmod +x /usr/local/bin/gitreceive
	gitreceive init

sshcommand:
	wget -qO /usr/local/bin/sshcommand ${SSHCOMMAND_URL}
	chmod +x /usr/local/bin/sshcommand
	sshcommand create dokku /usr/local/bin/dokku

pluginhook:
	wget -qO ${PLUGINHOOK_URL} /tmp/pluginhook_0.1.0_amd64.deb
	cd /tmp && dpkg -i pluginhook_0.1.0_amd64.deb

count:
	cat receiver dokku bootstrap.sh nginx-app-conf nginx-reloader.conf | wc -l