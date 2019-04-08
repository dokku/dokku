/tmp/build-dokku/var/lib/dokku/GIT_REV:
	mkdir -p /tmp/build-dokku
	mkdir -p /tmp/build-dokku/usr/share/bash-completion/completions
	mkdir -p /tmp/build-dokku/usr/bin
	mkdir -p /tmp/build-dokku/usr/share/doc/dokku
	mkdir -p /tmp/build-dokku/usr/share/dokku/contrib
	mkdir -p /tmp/build-dokku/usr/share/lintian/overrides
	mkdir -p /tmp/build-dokku/usr/share/man/man1
	mkdir -p /tmp/build-dokku/var/lib/dokku/core-plugins/available

	cp dokku /tmp/build-dokku/usr/bin
	cp LICENSE /tmp/build-dokku/usr/share/doc/dokku/copyright
	cp contrib/bash-completion /tmp/build-dokku/usr/share/bash-completion/completions/dokku
	find . -name ".DS_Store" -depth -exec rm {} \;
	$(MAKE) go-build
	cp common.mk /tmp/build-dokku/var/lib/dokku/core-plugins/common.mk
	cp -r plugins/* /tmp/build-dokku/var/lib/dokku/core-plugins/available
	find plugins/ -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | while read plugin; do cd /tmp/build-dokku/var/lib/dokku/core-plugins/available/$$plugin && if [ -e Makefile ]; then $(MAKE) src-clean; fi; done
	find plugins/ -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | while read plugin; do touch /tmp/build-dokku/var/lib/dokku/core-plugins/available/$$plugin/.core; done
	rm /tmp/build-dokku/var/lib/dokku/core-plugins/common.mk
	$(MAKE) help2man
	$(MAKE) addman
	cp /usr/local/share/man/man1/dokku.1 /tmp/build-dokku/usr/share/man/man1/dokku.1
	gzip -9 /tmp/build-dokku/usr/share/man/man1/dokku.1
	cp contrib/dokku-installer.py /tmp/build-dokku/usr/share/dokku/contrib
ifeq ($(DOKKU_VERSION),master)
	git describe --tags > /tmp/build-dokku/var/lib/dokku/VERSION
else
	echo $(DOKKU_VERSION) > /tmp/build-dokku/var/lib/dokku/VERSION
endif
ifdef DOKKU_GIT_REV
	echo "$(DOKKU_GIT_REV)" > /tmp/build-dokku/var/lib/dokku/GIT_REV
else
	git rev-parse HEAD > /tmp/build-dokku/var/lib/dokku/GIT_REV
endif
