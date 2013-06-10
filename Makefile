
all: install

install:
	cp receiver /home/git/receiver
	cp deploystep /home/git/deploystep
	cp nginx-app-conf /home/git/nginx-app-conf
	cp nginx-reloader.conf /etc/init/nginx-reloader.conf