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

The following pluginhooks describe those available to a dokku installation. As well, there is an example for each pluginhook that you can use as templates for your own plugin development.

> The example pluginhook code is not guaranteed to be implemented as in within dokkku, and are merely simplified examples. Please look at the dokku source for larger, more in-depth examples.

### `install`

- Description: Used to setup any files/configuration for a plugin.
- Invoked by: `dokki plugins-install`.
- Arguments: None
- Example:

```shell
#!/usr/bin/env bash
# Sets the hostname of the dokku server
# based on the output of `hostname -f`

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

if [[ ! -f  "$DOKKU_ROOT/HOSTNAME" ]]; then
  hostname -f > $DOKKU_ROOT/HOSTNAME
fi
```

### `dependencies`

- Description: Used to install system-level dependencies. Invoked by `plugins-install-dependencies`.
- Invoked by: `dokku plugins-install-dependencies`
- Arguments: None
- Example:

```shell
#!/usr/bin/env bash
# Installs nginx for the current plugin
# Supports both opensuse and ubuntu

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

case "$DOKKU_DISTRO" in
  ubuntu)
    export DEBIAN_FRONTEND=noninteractive
    apt-get install --force-yes -qq -y nginx
    ;;

  opensuse)
    zypper -q in -y nginx
    ;;
esac
```

### `update`

- Description: Can be used to run plugin updates on a regular interval. You can schedule the invoker in a cron-task to ensure your system gets regular updates.
- Invoked by: `dokku plugins-update`.
- Arguments: None
- Example:

```shell
#!/usr/bin/env bash
# Update the buildstep image from git source

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

cd /root/dokku
sudo BUILD_STACK=true make install
```

### `commands help`

- Description: Used to aggregate all plugin `help` output. Your plugin should implement a `help` command in your `commands` file to take advantage of this pluginhook. This must be implemented inside the `commands` pluginhook file.
- Invoked by: `dokku help`
- Arguments: None
- Example:

```shell
#!/usr/bin/env bash
# Outputs help for the derp plugin

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

case "$1" in
  help | derp:help)
    cat && cat<<EOF
    derp:herp                                       Herps the derp
    derp:serp [file]                                Shows the file's serp
EOF
    ;;

  *)
    exit $DOKKU_NOT_IMPLEMENTED_EXIT
    ;;

esac
```

### `backup-export`

- Description: Used to backup files for a given plugin. If your plugin writes files to disk, this pluginhook should be used to echo out their full paths. Any files listed will be copied by the backup plugin to the backup tar.gz.
- Invoked by: `dokku backup:export`
- Arguments: `$VERSION $DOKKU_ROOT`
- Example:

```shell
#!/usr/bin/env bash
# Echos out the location of every `REDIRECT` file
# that are used by the apps

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

shopt -s nullglob
VERSION="$1"
DOKKU_ROOT="$2"

cat; for i in $DOKKU_ROOT/*/REDIRECT; do echo $i; done
```

### `backup-check`

- Description:
- Invoked by: `dokku backup:import`
- Arguments: `$VERSION "$BACKUP_ROOT" "$TARGET_DIR" "$BACKUP_TMP_DIR/.dokku_backup_apps"`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `backup-import`

- Description:
- Invoked by: `dokku backup:import`
- Arguments: `$VERSION "$BACKUP_ROOT" $TARGET_DIR "$BACKUP_TMP_DIR/.dokku_backup_apps"`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `pre-build-buildstep`

- Description:
- Invoked by: `dokku build`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `post-build-buildstep`

- Description:
- Invoked by: `dokku build`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `pre-release-buildstep`

- Description:
- Invoked by: `dokku release`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `post-release-buildstep`

- Description:
- Invoked by: `dokku release`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `pre-build-dockerfile`

- Description:
- Invoked by: `dokku build`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `post-build-dockerfile`

- Description:
- Invoked by: `dokku build`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `pre-release-dockerfile`

- Description:
- Invoked by: `dokku release`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `post-release-dockerfile`

- Description:
- Invoked by: `dokku release`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `check-deploy`

- Description:
- Invoked by: `dokku deploy`
- Arguments: `$CONTAINER_ID $APP $INTERNAL_PORT $INTERNAL_IP_ADDRESS`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `pre-deploy`

- Description:
- Invoked by: `dokku deploy`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `post-deploy`

- Description:
- Invoked by: `dokku deploy`
- Arguments: `$APP $INTERNAL_PORT $INTERNAL_IP_ADDRESS`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `pre-delete`

- Description:
- Invoked by: `dokku apps:destroy`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `post-delete`

- Description:
- Invoked by: `dokku apps:destroy`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `docker-args-build`

- Description:
- Invoked by: `dokku build`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `docker-args-deploy`

- Description:
- Invoked by: `dokku deploy`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `docker-args-run`

- Description:
- Invoked by: `dokku run`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `bind-external-ip`

- Description:
- Invoked by: `dokku deploy`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `post-domains-update`

- Description:
- Invoked by: `dokku domains:add`, `dokku domains:clear`, `dokku domains:remove`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `git-pre-pull`

- Description:
- Invoked by: `dokku git-upload-pack`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `git-post-pull`

- Description:
- Invoked by: `dokku git-upload-pack`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `nginx-hostname`

- Description:
- Invoked by: `dokku domains:setup`
- Arguments: `$APP $SUBDOMAIN $VHOST`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `nginx-pre-reload`

- Description:
- Invoked by: `dokku nginx:build-config`
- Arguments: `$APP $INTERNAL_PORT $INTERNAL_IP_ADDRESS`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```

### `receive-app`

- Description:
- Invoked by: `dokku git-hook`, `dokku ps:rebuild`
- Arguments: `$APP $REV`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

```
