GITRECEIVE_URL = https://raw.github.com/progrium/gitreceive/master/gitreceive
SSHCOMMAND_URL = https://raw.github.com/progrium/sshcommand/master/sshcommand

all: install

install: submodule gitreceive sshcommand
	cp receiver /home/git/receiver
	cp deploystep /home/git/deploystep
	cp buildstep/buildstep /home/git/buildstep
	cp nginx-app-conf /home/git/nginx-app-conf
	cp nginx-reloader.conf /etc/init/nginx-reloader.conf

submodule:
	git submodule init
	git submodule update

gitreceive:
	wget -O /usr/local/bin/gitreceive ${GITRECEIVE_URL}
	chmod +x /usr/local/bin/gitreceive
	test ! -d /home/git && gitreceive init

sshcommand:
	wget -O /usr/local/bin/sshcommand ${SSHCOMMAND_URL}
	chmod +x /usr/local/bin/sshcommand

count:
	cat receiver deploystep bootstrap.sh nginx-app-conf nginx-reloader.conf | wc -l