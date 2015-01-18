# Pluginhooks

[Pluginhooks](https://github.com/progrium/pluginhook) are a good way to jack into existing dokku infrastructure. You can use them to modify the output of various dokku commands or override internal configuration.

Pluginhooks are simply scripts that are executed by the system. You can use any language you want, so long as the script:

- Is executable
- Has the proper language requirements installed

For instance, if you wanted to write a pluginhook in PHP, you would need to have `php` installed and available on the CLI prior to pluginhook invocation.

The following is an example for the `nginx-hostname` pluginhook. It reverses the hostname that is provided to nginx during deploys. If you created an executable file named `nginx-hostname` with the following code in your plugin, it would be invoked by dokku during the normal app deployment process:

```shell
#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"; SUBDOMAIN="$2"; VHOST="$3"

NEW_SUBDOMAIN=`echo $SUBDOMAIN | rev`
echo "$NEW_SUBDOMAIN.$VHOST"
```

## Available Pluginhooks

There are a number of plugin-related pluginhooks. These can be optionally implemented by plugins and allow integration into the standard dokku plugin setup/backup/teardown process.

- pluginhook install: Used to setup any files/configuration for a plugin. Invoked by `pluginhook install`.
- pluginhook dependencies: Used to install system-level dependencies. Invoked by `plugins-install-dependencies`.
- pluginhook update: Can be used to run plugin updates on a regular interval. Invoked by `dokku plugins-update`.
- pluginhook commands help: Used to aggregate all plugin `help` output. Your plugin should implement a `help` command in your `commands` file to take advantage of this pluginhook. Invoked by `dokku help`
- pluginhook backup-export 1 $BACKUP_DIR
- pluginhook backup-check $VERSION "$BACKUP_ROOT" "$TARGET_DIR" "$BACKUP_TMP_DIR/.dokku_backup_apps"
- pluginhook backup-import $VERSION "$BACKUP_ROOT" $TARGET_DIR "$BACKUP_TMP_DIR/.dokku_backup_apps"

- pluginhook pre-build $APP
- pluginhook post-build $APP
- pluginhook pre-release $APP
- pluginhook post-release $APP
- pluginhook check-deploy $id $APP $port ${ipaddr:-localhost}
- pluginhook pre-deploy $APP
- pluginhook post-deploy  $APP $port $ipaddr
- pluginhook pre-delete $APP
- pluginhook post-delete $APP
- pluginhook docker-args $APP build
- pluginhook docker-args $APP deploy
- pluginhook docker-args $APP run
- pluginhook bind-external-ip $APP
- pluginhook post-domains-update $APP
- pluginhook post-domains-update $APP
- pluginhook post-domains-update $APP
- pluginhook git-pre-pull $APP
- pluginhook git-post-pull $APP
- pluginhook nginx-hostname $APP $SUBDOMAIN $VHOST
- pluginhook nginx-pre-reload $APP $DOKKU_APP_LISTEN_PORT $DOKKU_APP_LISTEN_IP
