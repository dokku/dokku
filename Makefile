
all: install

install: submodule
	cp receiver /home/git/receiver
	cp deploystep /home/git/deploystep
	cp buildstep/buildstep /home/git/buildstep
	cp nginx-app-conf /home/git/nginx-app-conf
	cp nginx-reloader.conf /etc/init/nginx-reloader.conf

submodule:
	git submodule init
	git submodule update

count:
	cat receiver deploystep bootstrap.sh nginx-app-conf nginx-reloader.conf | wc -l