# Dokku Event Logs

Docker provides an _events_ command to show system's real time events. Likewise, Dokku can record events as syslog entries and also provides a plugin to display the last ones.

```
events [-t]                                     Show the last events (-t follows)
events:list                                     List logged events
events:on                                       Enable events logger
events:off                                      Disable events logger
```

## Usage

Enable the plugin:

```
$ dokku events:on
Enabling dokku events logger
```

Dokku will then write log entries to ``/var/log/syslog`` as well as a Dokku-specific logs sub-directory ``/var/log/dokku/events.log``, which can be easily displayed with the command `dokku events`:

```
Jul  3 16:09:48 dokku.me dokku[127630]: INVOKED: pre-release-buildpack( pythonapp )
Jul  3 16:10:02 dokku.me dokku[128095]: INVOKED: docker-args-run( rubyapp )
Jul  3 16:10:02 dokku.me dokku[128114]: INVOKED: docker-args-run( nhl )
Jul  3 16:10:03 dokku.me dokku[128136]: INVOKED: post-release-buildpack( pythonapp )
Jul  3 16:10:03 dokku.me dokku[128195]: INVOKED: pre-deploy( pythonapp )
Jul  3 16:10:23 dokku.me dokku[129253]: INVOKED: docker-args-deploy( pythonapp )
Jul  3 16:10:24 dokku.me dokku[129337]: INVOKED: bind-external-ip( pythonapp )
Jul  3 16:10:24 dokku.me dokku[129451]: INVOKED: check-deploy( pythonapp 6274ced0d4be11af4490cd18abaf77cdd593f025133f403d984e80d86a39acec web 5000 10.0.16.80 )
Jul  3 16:10:35 dokku.me dokku[129561]: INVOKED: docker-args-deploy( pythonapp )
Jul  3 16:10:35 dokku.me dokku[129649]: INVOKED: bind-external-ip( pythonapp )
Jul  3 16:10:36 dokku.me dokku[129760]: INVOKED: check-deploy( pythonapp ac88a56ee4161ff37e4b92d1498c3eadc91f0aa7c8b81b44fc077e2a51d54cc0 worker )
Jul  3 16:10:46 dokku.me dokku[129851]: INVOKED: post-deploy( pythonapp )
Jul  3 16:10:46 dokku.me dokku[129945]: INVOKED: nginx-pre-reload( pythonapp )
Jul  3 16:15:02 dokku.me dokku[130397]: INVOKED: docker-args-run( goapp )
Jul  3 16:21:02 dokku.me dokku[130796]: INVOKED: docker-args-run( rubyapp )
Jul  3 16:30:02 dokku.me dokku[131384]: INVOKED: docker-args-run( rubyapp )
```

You can list all events that are currently being recorded via `dokku events:list`:
```
=====> Events currently logged
backup-check
docker-args-build
pre-delete
post-deploy
receive-app
pre-release-buildpack
nginx-pre-reload
docker-args-deploy
post-build-buildpack
post-domains-update
post-release-dockerfile
check-deploy
backup-import
backup-export
git-pre-pull
git-post-pull
nginx-hostname
bind-external-ip
pre-deploy
post-delete
dependencies
pre-build-buildpack
docker-args-run
pre-build-dockerfile
post-release-buildpack
post-build-dockerfile
pre-release-dockerfile
update
```
