# Plugin triggers

[Plugin triggers](https://github.com/dokku/plugn) (formerly [pluginhooks](https://github.com/progrium/pluginhook)) are a good way to jack into existing Dokku infrastructure. You can use them to modify the output of various Dokku commands or override internal configuration.

Plugin triggers are simply scripts that are executed by the system. You can use any language you want, so long as the script:

- Is executable
- Has the proper language requirements installed

For instance, if you wanted to write a plugin trigger in PHP, you would need to have `php` installed and available on the CLI prior to plugin trigger invocation.

The following is an example for the `nginx-hostname` plugin trigger. It reverses the hostname that is provided to nginx during deploys. If you created an executable file named `nginx-hostname` with the following code in your plugin trigger, it would be invoked by Dokku during the normal app deployment process:

```shell
#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"; SUBDOMAIN="$2"; VHOST="$3"

NEW_SUBDOMAIN=`echo $SUBDOMAIN | rev`
echo "$NEW_SUBDOMAIN.$VHOST"
```

## Available plugin triggers

There are a number of plugin-related triggers. These can be optionally implemented by plugins and allow integration into the standard Dokku setup/teardown process.

The following plugin triggers describe those available to a Dokku installation. As well, there is an example for each trigger that you can use as templates for your own plugin development.

> The example plugin trigger code is not guaranteed to be implemented as in within dokku, and are merely simplified examples. Please look at the Dokku source for larger, more in-depth examples.

### `app-restart`

- Description: Triggers an app restart
- Invoked by: `dokku config:clear`, `dokku config:set`, `dokku config:unset`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `app-urls`

- Description: Allows you to change the urls Dokku reports for an application. Will override any auto-detected urls.
- Invoked by: `dokku deploy`, `dokku url`, and `dokku urls`
- Arguments: `$APP $URL_TYPE`
- Example:

```shell
#!/usr/bin/env bash
# Sets the domain to `internal.tld`
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"; URL_TYPE="$2"
case "$URL_TYPE" in
  url)
    echo "https://internal.tld/${APP}/"
    ;;
  urls)
    echo "https://internal.tld/${APP}/"
    echo "http://internal.tld/${APP}/"
    ;;
esac
```

### `check-deploy`

- Description: Allows you to run checks on a deploy before Dokku allows the container to handle requests.
- Invoked by: `dokku deploy`
- Arguments: `$APP $CONTAINER_ID $PROC_TYPE $PORT $IP`
- Example:

```shell
#!/usr/bin/env bash
# Disables deploys of containers based on whether the
# `DOKKU_DISABLE_DEPLOY` env var is set to `true` for an app

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_AVAILABLE_PATH/config/functions"

APP="$1"; CONTAINERID="$2"; PROC_TYPE="$3"; PORT="$4" ; IP="$5"

eval "$(config_export app $APP)"
DOKKU_DISABLE_DEPLOY="${DOKKU_DISABLE_DEPLOY:-false}"

if [[ "$DOKKU_DISABLE_DEPLOY" = "true" ]]; then
  echo -e "\033[31m\033[1mDeploys disabled, sorry.\033[0m"
  exit 1
fi
```

### `commands help` and `commands <PLUGIN_NAME>:help`

- Description: Your plugin should implement a `help` command in your `commands` file to take advantage of this plugin trigger. `commands help` is used by `dokku help` to aggregate all plugins abbreviated `help` output. Implementing  `<PLUGIN_NAME>:help` in your `commands` file gives users looking for help, a more detailed output. 'commands help' must be implemented inside the `commands` plugin file. It's recommended that `PLUGIN_NAME:help` be added to the commands file to ensure consistency among community plugins and give you a new avenue to share rich help content with your user.
- Invoked by: `dokku help` and `commands <PLUGIN_NAME>:help`
- Arguments: None
- Example:

```shell
#!/usr/bin/env bash
# Outputs help for the derp plugin

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

case "$1" in
  help | hello:help)
    help_content_func () {
      declare desc="return help_content string"
      cat<<help_content
    hello <app>, Says "Hello <app>"
    hello:world, Says "Hello world"
help_content
    }

    if [[ $1 = "hello:help" ]] ; then
        echo -e 'Usage: dokku hello[:world] [<app>]'
        echo ''
        echo 'Say Hello World.'
        echo ''
        echo 'Example:'
        echo ''
        echo '$ dokku hello:world'
        echo 'Hello world'
        echo ''
        echo 'Additional commands:'
        help_content_func | sort | column -c2 -t -s,
    else
        help_content_func
    fi
    ;;

  *)
    exit $DOKKU_NOT_IMPLEMENTED_EXIT
    ;;

esac
```

### `core-post-deploy`

> To avoid issues with community plugins, this plugin trigger should be used *only* for core plugins. Please avoid using this trigger in your own plugins.

- Description: Allows running of commands after an app's processes have been scaled up, but before old containers are torn down. Dokku core currently uses this to switch traffic on nginx.
- Invoked by: `dokku deploy`
- Arguments: `$APP $INTERNAL_PORT $INTERNAL_IP_ADDRESS $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash
# Notify an external service that a successful deploy has occurred.

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

curl "http://httpstat.us/200"
```

### `dependencies`

- Description: Used to install system-level dependencies.
- Invoked by: `dokku plugin:install-dependencies`
- Arguments: None
- Example:

```shell
#!/usr/bin/env bash
# Installs nginx for the current plugin
# Supports both opensuse and ubuntu

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

export DEBIAN_FRONTEND=noninteractive

case "$DOKKU_DISTRO" in
  debian|ubuntu)
    apt-get install --force-yes -qq -y nginx
    ;;

  opensuse)
    zypper -q in -y nginx
    ;;
esac
```

### `deploy-source`

- Description: Used for reporting what the current detected deployment source is. The first detected source should always win.
- Invoked by: `dokku apps:report`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash
# Checks if the app should be deployed via git

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"
STDIN=$(cat)

# bail if another source is detected
if [[ -n "$STDIN" ]]; then
  echo "$STDIN"
  return
fi

if [[ -d "$DOKKU_ROOT/$APP/refs" ]]; then
  echo "git"
fi
```

### `deployed-app-image-repo`

- Description: Used to manage the full repo of the image being deployed. Useful for deploying from an external registry where the repository name is not `dokku/$APP`
- Invoked by: `internal function dokku_deploy_cmd() (deploy phase)`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"
# change the repo from dokku/APP to dokkupaas/APP
echo "dokkupaas/$APP"
```

### `deployed-app-image-tag`

- Description: Used to manage the tag of the image being deployed. Useful for deploying a specific version of an image, or when deploying from an external registry
- Invoked by: `internal function dokku_deploy_cmd() (deploy phase)`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# customize the tag version
echo 'not-latest'
```

### `deployed-app-repository`

- Description: Used to manage the remote repository of the image being deployed.
- Invoked by: `internal function dokku_deploy_cmd() (deploy phase)`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

echo 'derp.dkr.ecr.us-east-1.amazonaws.com'
```

### `docker-args-build`

> Warning: Deprecated, please use `docker-args-process-build` instead

- Description:
- Invoked by: `internal function dokku_build() (build phase)`
- Arguments: `$APP $IMAGE_SOURCE_TYPE`
- Example:

```shell
#!/usr/bin/env bash
# Sets a docker build-arg called CACHEBUST which can be used
# to bust cache at any arbitrary point in a Dockerfile build
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

STDIN=$(cat)
APP="$1"; IMAGE_SOURCE_TYPE="$2"
output=""

if [[ "$IMAGE_SOURCE_TYPE" == "dockerfile" ]]; then
  output=" --build-arg CACHEBUST=$(date +%s)"
fi
echo -n "$STDIN$output"
```

### `docker-args-deploy`

> Warning: Deprecated, please use `docker-args-process-deploy` instead

- Description:
- Invoked by: `dokku deploy`
- Arguments: `$APP $IMAGE_TAG [$PROC_TYPE $CONTAINER_INDEX]`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; IMAGE_TAG="$2"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
verify_app_name "$APP"

# TODO
```

### `docker-args-process-build`

- Description: `$PROC_TYPE` may be set to magic `_all_` process type to signify global docker deploy options.
- Invoked by: `dokku ps:rebuild`
- Arguments: `$APP $IMAGE_TAG $IMAGE_SOURCE_TYPE`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; IMAGE_TAG="$2"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
verify_app_name "$APP"

# TODO
```

### `docker-args-process-deploy`

- Description: `$PROC_TYPE` may be set to magic `_all_` process type to signify global docker deploy options.
- Invoked by: `dokku deploy`
- Arguments: `$APP $IMAGE_TAG $IMAGE_SOURCE_TYPE [$PROC_TYPE $CONTAINER_INDEX]`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; IMAGE_TAG="$2"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
verify_app_name "$APP"

# TODO
```

### `docker-args-process-run`

- Description: `$PROC_TYPE` may be set to magic `_all_` process type to signify global docker run options.
- Invoked by: `dokku run`
- Arguments: `$APP $IMAGE_TAG $IMAGE_SOURCE_TYPE`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; IMAGE_TAG="$2"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
verify_app_name "$APP"

# TODO
```

### `docker-args-run`

> Warning: Deprecated, please use `docker-args-process-run` instead

- Description:
- Invoked by: `dokku run`
- Arguments: `$APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; IMAGE_TAG="$2"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
verify_app_name "$APP"

# TODO
```

### `git-post-pull`

- Description:
- Invoked by: `dokku git-upload-pack`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `git-pre-pull`

- Description:
- Invoked by: `dokku git-upload-pack`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

> Warning: The `git-pre-pull` trigger should _not_ be used for authentication
since it does not get called for commands that use `git-upload-archive` such
as `git archive`. Instead, use the [`user-auth`](#user-auth) trigger.

### `git-revision`

- Description: Allows you to fetch the current git revision for a given application
- Invoked by:
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `install`

- Description: Used to setup any files/configuration for a plugin.
- Invoked by: `dokku plugin:install`.
- Arguments: None
- Example:

```shell
#!/usr/bin/env bash
# Sets the hostname of the Dokku server
# based on the output of `hostname -f`

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

if [[ ! -f  "$DOKKU_ROOT/HOSTNAME" ]]; then
  hostname -f > $DOKKU_ROOT/HOSTNAME
fi
```

### `network-build-config`

- Description: Rebuilds network configuration
- Invoked by: `internally triggered by proxy-build-config within proxy implementations`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `network-clear-config`

- Description: Clears network configuration
- Invoked by: `internally triggered by proxy-clear-config within proxy implementations`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `network-compute-ports`

- Description: Computes the ports for a given app container
- Invoked by: `internally triggered by proxy-build-config within proxy implementations`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `network-config-exists`

- Description: Returns whether the network configuration for a given app exists
- Invoked by: `internally triggered by core-post-deploy within proxy implementations`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `network-get-ipaddr`

- Description: Return the ipaddr for a given app container
- Invoked by: `internally triggered by a deploy`
- Arguments: `$APP $PROC_TYPE $CONTAINER_ID`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `network-get-listeners`

- Description: Return the listeners (host:port combinations) for a given app container
- Invoked by: `internally triggered by a deploy`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `network-get-port`

- Description: Return the port for a given app container
- Invoked by: `internally triggered by a deploy`
- Arguments: `$APP $PROC_TYPE $CONTAINER_ID $IS_HEROKUISH_CONTAINER`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `network-get-property`

- Description: Return the network value for an app's property
- Invoked by: `internally triggered by a deploy`
- Arguments: `$APP $KEY`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `network-write-ipaddr`

- Description: Write the ipaddr for a given app index
- Invoked by: `internally triggered by a deploy`
- Arguments: `$APP $PROC_TYPE $CONTAINER_INDEX $IP_ADDRESS`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `network-write-port`

- Description: Write the port for a given app index
- Invoked by: `internally triggered by a deploy`
- Arguments: `$APP $PROC_TYPE $CONTAINER_INDEX $PORT`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `nginx-hostname`

- Description: Allows you to customize the hostname for a given app
- Invoked by: `dokku domains:setup`
- Arguments: `$APP $SUBDOMAIN $VHOST`
- Example:

```shell
#!/usr/bin/env bash
# Reverses the hostname for the app

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"; SUBDOMAIN="$2"; VHOST="$3"

NEW_SUBDOMAIN=`echo $SUBDOMAIN | rev`
echo "$NEW_SUBDOMAIN.$VHOST"
```

### `nginx-pre-reload`

> Warning: The arguments INTERNAL_PORT and INTERNAL_IP_ADDRESS are no longer sufficient to retrieve all app listeners. Please run `plugn trigger network-get-listeners APP` within any implementation of `nginx-pre-reload` in order to retrieve all application listeners.

- Description: Run before nginx reloads hosts
- Invoked by: `dokku nginx:build-config`
- Arguments: `$APP $INTERNAL_PORT $INTERNAL_IP_ADDRESS`
- Example:

```shell
#!/usr/bin/env bash
# Runs a check against all nginx conf files
# to ensure they are valid

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

nginx -t
```

### `post-app-clone`

- Description: Allows you to run commands after an app was cloned.
- Invoked by: `dokku apps:clone`
- Arguments: `$OLD_APP_NAME $NEW_APP_NAME`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `post-app-clone-setup`

- Description: Allows you to run commands after an app is setup, and before it is rebuild. This is useful for cleaning up tasks, or ensuring configuration from an old app is copied to the new app
- Invoked by: `dokku apps:clone`
- Arguments: `$OLD_APP_NAME $NEW_APP_NAME`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `post-app-rename`

- Description: Allows you to run commands after an app was renamed.
- Invoked by: `dokku apps:rename`
- Arguments: `$OLD_APP_NAME $NEW_APP_NAME`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `post-build-buildpack`

- Description: Allows you to run commands after the build image is create for a given app. Only applies to apps using buildpacks.
- Invoked by: `internal function dokku_build() (build phase)`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `post-build-dockerfile`

- Description: Allows you to run commands after the build image is create for a given app. Only applies to apps using a dockerfile.
- Invoked by: `internal function dokku_build() (build phase)`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `post-certs-remove`

- Description: Allows you to run commands after a cert is removed
- Invoked by: `dokku certs:remove`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; verify_app_name "$APP"

# TODO
```

### `post-certs-update`

- Description: Allows you to run commands after a cert is added/updated
- Invoked by: `dokku certs:add`, `dokku certs:update`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; verify_app_name "$APP"

# TODO
```

### `post-config-update`

- Description: Allows you to get notified when one or more configs is added or removed. Action can be *set* or *unset*
- Invoked by: `dokku config:set`, `dokku config:unset`
- Arguments: `$APP` `set|unset` `key1=VALUE1 key2=VALUE2`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `post-create`

- Description: Can be used to run commands after an app is created.
- Invoked by: `dokku apps:create`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash
# Runs a command to ensure that an app
# has a postgres database when it is starting

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1";
POSTGRES="$1"

dokku postgres:create $POSTGRES
dokku postgres:link $POSTGRES $APP
```

### `post-delete`

- Description: Can be used to run commands after an app is deleted.
- Invoked by: `dokku apps:destroy`
- Arguments: `$APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash
# Runs a command to ensure that an app's
# postgres installation is removed

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1";

dokku postgres:destroy $APP
```

### `post-deploy`

> Please see [core-post-deploy](#core-post-deploy) if contributing a core plugin with the `post-deploy` hook.

- Description: Allows running of commands after an app's processes have been scaled up, but before old containers are torn down. Dokku calls this _after_ `core-post-deploy`. Deployment Tasks are also invoked by this plugin trigger.
- Invoked by: `dokku deploy`
- Arguments: `$APP $INTERNAL_PORT $INTERNAL_IP_ADDRESS $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash
# Notify an external service that a successful deploy has occurred.

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

curl "http://httpstat.us/200"
```

### `post-domains-update`

- Description: Allows you to run commands once the domain for an app has been updated. It also sends in the command that has been used. This can be "add", "clear" or "remove". The third argument will be the optional list of domains
- Invoked by: `dokku domains:add`, `dokku domains:clear`, `dokku domains:remove`, `dokku domains:set`
- Arguments: `$APP` `action name` `domains`
- Example:

```shell
#!/usr/bin/env bash
# Reloads haproxy for our imaginary haproxy plugin
# that replaces the nginx-vhosts plugin

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

sudo service haproxy reload
```

### `post-extract`

- Description: Allows you to modify the contents of an app *after* it has been extracted from git/tarball but *before* the image source type is detected.
- Invoked by: `dokku tar:in`, `dokku tar:from` and the `receive-app` plugin trigger
- Arguments: `$APP` `$TMP_WORK_DIR` `$REV`
- Example:

```shell
#!/usr/bin/env bash
# Adds a clock process to an app's Procfile

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; verify_app_name "$APP"
TMP_WORK_DIR="$2"
REV="$3" # optional, may not be sent for tar-based builds

pushd "$TMP_WORK_DIR" >/dev/null
touch Procfile
echo "clock: some-command" >> Procfile
```

### `post-proxy-ports-update`

- Description: Allows you to run commands once the proxy port mappings for an app have been updated. It also sends the invoking command. This can be "add", "clear" or "remove".
- Invoked by: `dokku proxy:ports-add`, `dokku proxy:ports-clear`, `dokku proxy:ports-remove`
- Arguments: `$APP` `action name`
- Example:

```shell
#!/usr/bin/env bash
# Rebuilds haproxy config for our imaginary haproxy plugin
# that replaces the nginx-vhosts plugin

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
source "$PLUGIN_AVAILABLE_PATH/haproxy/functions"
APP="$1"; verify_app_name "$APP"

haproxy-build-config "$APP"
```

### `post-release-buildpack`

- Description: Allows you to run commands after environment variables are set for the release step of the deploy. Only applies to apps using buildpacks.
- Invoked by: `internal function dokku_release() (release phase)`
- Arguments: `$APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash
# Installs a package specified by the `CONTAINER_PACKAGE` env var

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; IMAGE_TAG="$2"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
verify_app_name "$APP"

dokku_log_info1 "Installing $CONTAINER_PACKAGE..."

CMD="cat > gm && \
  dpkg -s CONTAINER_PACKAGE >/dev/null 2>&1 || \
  (apt-get update && apt-get install -y CONTAINER_PACKAGE && apt-get clean)"

CID=$(docker run $DOKKU_GLOBAL_RUN_ARGS -i -a stdin $IMAGE /bin/bash -c "$CMD")
test $(docker wait $CID) -eq 0
DOCKER_COMMIT_LABEL_ARGS=("--change" "LABEL org.label-schema.schema-version=1.0" "--change" "LABEL org.label-schema.vendor=dokku" "--change" "LABEL com.dokku.app-name=$APP")
docker commit "${DOCKER_COMMIT_LABEL_ARGS[@]}" $CID $IMAGE >/dev/null
```

### `post-release-dockerfile`

- Description: Allows you to run commands after environment variables are set for the release step of the deploy. Only applies to apps using a dockerfile.
- Invoked by: `internal function dokku_release() (release phase)`
- Arguments: `$APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; IMAGE_TAG="$2"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
verify_app_name "$APP"

# TODO
```

### `post-stop`

- Description: Can be used to run commands after an app is manually stopped
- Invoked by: `dokku ps:stop` and `dokku ps:stopall`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash
# Marks an app as manually stopped

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1";

dokku config:set --no-restart $APP MANUALLY_STOPPED=1
```

### `pre-build-buildpack`

- Description: Allows you to run commands before the build image is created for a given app. For instance, this can be useful to add env vars to your container. Only applies to apps using buildpacks.
- Invoked by: `internal function dokku_build() (build phase)`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `pre-build-dockerfile`

- Description: Allows you to run commands before the build image is created for a given app. For instance, this can be useful to add env vars to your container. Only applies to apps using a dockerfile.
- Invoked by: `internal function dokku_build() (build phase)`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `pre-delete`

- Description: Can be used to run commands before an app is deleted.
- Invoked by: `dokku apps:destroy`
- Arguments: `$APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash
# Clears out the gulp asset build cache for apps

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"

APP="$1"; GULP_CACHE_DIR="$DOKKU_ROOT/$APP/gulp"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
verify_app_name "$APP"

if [[ -d $GULP_CACHE_DIR ]]; then
  docker run "${DOCKER_COMMIT_LABEL_ARGS[@]}" --rm -v "$GULP_CACHE_DIR:/gulp" "$IMAGE" find /gulp -depth -mindepth 1 -maxdepth 1 -exec rm -Rf {} \; || true
fi
```

### `pre-deploy`

- Description: Allows the running of code before the app's processes are scaled up and after the docker images are prepared.
- Invoked by: `dokku deploy`
- Arguments: `$APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash
# Runs gulp in our container

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; IMAGE_TAG="$2"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
verify_app_name "$APP"

dokku_log_info1 "Running gulp"
CID=$(docker run "${DOCKER_COMMIT_LABEL_ARGS[@]}" -d $IMAGE /bin/bash -c "cd /app && gulp default")
test $(docker wait $CID) -eq 0
DOCKER_COMMIT_LABEL_ARGS=("--change" "LABEL org.label-schema.schema-version=1.0" "--change" "LABEL org.label-schema.vendor=dokku" "--change" "LABEL com.dokku.app-name=$APP")
docker commit "${DOCKER_COMMIT_LABEL_ARGS[@]}" $CID $IMAGE >/dev/null
```

### `pre-disable-vhost`

- Description: Allows you to run commands before the VHOST feature is disabled
- Invoked by: `dokku domains:disable`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; verify_app_name "$APP"

# TODO
```

### `pre-enable-vhost`

- Description: Allows you to run commands before the VHOST feature is enabled
- Invoked by: `dokku domains:enable`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; verify_app_name "$APP"

# TODO
```

### `pre-receive-app`

- Description: Allows you to customize the contents of an app directory before they are processed for deployment. The `IMAGE_SOURCE_TYPE` can be any of `[herokuish, dockerfile]`
- Invoked by: `dokku git-hook`, `dokku tar-build-locked`
- Arguments: `$APP $IMAGE_SOURCE_TYPE $TMP_WORK_DIR $REV`
- Example:

```shell
#!/usr/bin/env bash
# Adds a file called `dokku-is-awesome` to the repository
# the contents will be the app name

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"; IMAGE_SOURCE_TYPE="$2"; TMP_WORK_DIR="$3"; REV="$4"

echo "$APP" > "$TMP_WORK_DIR/dokku-is-awesome"
```

### `pre-release-buildpack`

- Description: Allows you to run commands before environment variables are set for the release step of the deploy. Only applies to apps using buildpacks.
- Invoked by: `internal function dokku_release() (release phase)`
- Arguments: `$APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash
# Installs the graphicsmagick package into the container

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; IMAGE_TAG="$2"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
verify_app_name "$APP"

dokku_log_info1 "Installing GraphicsMagick..."

CMD="cat > gm && \
  dpkg -s graphicsmagick >/dev/null 2>&1 || \
  (apt-get update && apt-get install -y graphicsmagick && apt-get clean)"

CID=$(docker run "${DOCKER_COMMIT_LABEL_ARGS[@]}" -i -a stdin $IMAGE /bin/bash -c "$CMD")
DOCKER_COMMIT_LABEL_ARGS=("--change" "LABEL org.label-schema.schema-version=1.0" "--change" "LABEL org.label-schema.vendor=dokku" "--change" "LABEL com.dokku.app-name=$APP")
test $(docker wait $CID) -eq 0
docker commit "${DOCKER_COMMIT_LABEL_ARGS[@]}" $CID $IMAGE >/dev/null
```

### `pre-release-dockerfile`

- Description: Allows you to run commands before environment variables are set for the release step of the deploy. Only applies to apps using a dockerfile.
- Invoked by: `internal function dokku_release() (release phase)`
- Arguments: `$APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"
APP="$1"; IMAGE_TAG="$2"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
verify_app_name "$APP"

# TODO
```

### `pre-restore`

- Description: Allows you to run commands before all containers are restored
- Invoked by: `dokku ps:restore`
- Arguments: `$DOKKU_SCHEDULER`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"

# TODO
```

### `pre-start`

- Description: Can be used to run commands before an app is started
- Invoked by: `dokku ps:start`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash
# Notifies an example url that an app is starting

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1";

curl "https://dokku.me/starting/${APP}" || true
```

### `proxy-build-config`

- Description: Builds the proxy implementation configuration for a given app
- Invoked by: `internally triggered by ps:restore`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `proxy-clear-config`

- Description: Clears the proxy implementation configuration for a given app
- Invoked by: `internally triggered by apps:rename`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `proxy-disable`

- Description: Disables the configured proxy implementation for an app
- Invoked by: `internally triggered by ps:restore`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `proxy-enable`

- Description: Enables the configured proxy implementation for an app
- Invoked by: `internally triggered by ps:restore`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

# TODO
```

### `receive-app`

- Description: Allows you to customize what occurs when an app is received. Normally just triggers an app build.
- Invoked by: `dokku git-hook`, `dokku ps:rebuild`
- Arguments: `$APP $REV` (`$REV` may not be included in cases where a repository is not pushed)
- Example:

```shell
#!/usr/bin/env bash
# For our imaginary mercurial plugin, triggers a rebuild

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"; REV="$2"

dokku hg-build $APP $REV
```

### `receive-branch`

- Description: Allows you to customize what occurs when a specific branch is received. Can be used to add support for specific branch names
- Invoked by: `dokku git-hook`, `dokku ps:rebuild`
- Arguments: `$APP $REV $REFNAME`
- Example:

```shell
#!/bin/bash
# Gives Dokku the ability to support multiple branches for a given service
# Allowing you to have multiple staging environments on a per-branch basis

reference_app=$1
refname=$3
newrev=$2
APP=${refname/*\//}.$reference_app

if [[ ! -d "$DOKKU_ROOT/$APP" ]]; then
  REFERENCE_REPO="$DOKKU_ROOT/$reference_app"
  git clone --bare --shared --reference "$REFERENCE_REPO" "$REFERENCE_REPO" "$DOKKU_ROOT/$APP" >/dev/null
fi
plugn trigger receive-app $APP $newrev
```

### `report`

- Description: Allows you to report on any custom configuration in use by your application
- Invoked by: `dokku report`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
APP="$1";

# TODO
```

### `resource-get-property`

- Description: Fetches a given resource property value
- Invoked by:
- Arguments: `$APP` `$PROC_TYPE` `$RESOURCE_TYPE` `$PROPERTY`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
APP="$1"; PROC_TYPE="$2" RESOURCE_TYPE="$3" PROPERTY="$4"

# TODO
```

### `retire-container-failed`

- Description: Allows you to run commands if/when retiring old containers has failed
- Invoked by: `dokku deploy`
- Arguments: `$APP`
- Example:

```shell
#!/usr/bin/env bash
# Send an email when a container failed to retire

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
APP="$1"; HOSTNAME=$(hostname -s)

mail -s "$APP containers on $HOSTNAME failed to retire" ops@dokku.me
```

### `scheduler-app-status`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Fetch the status of an app
- Invoked by: `dokku ps:report`
- Arguments: `$DOKKU_SCHEDULER $APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"; APP="$2";

# TODO
```

### `scheduler-deploy`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to run scheduler commands when an app is deployed
- Invoked by: `dokku deploy`
- Arguments: `$DOKKU_SCHEDULER $APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"; APP="$2"; IMAGE_TAG="$3";

# TODO
```

### `scheduler-docker-cleanup`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to run scheduler commands when dokku cleanup is invoked
- Invoked by: `dokku deploy, dokku cleanup`
- Arguments: `$DOKKU_SCHEDULER $APP $FORCE_CLEANUP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"; APP="$2"; FORCE_CLEANUP="$3";

# TODO
```

### `scheduler-inspect`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to run inspect commands for all containers for a given app
- Invoked by: `dokku ps:inspect`
- Arguments: `$DOKKU_SCHEDULER $APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"; APP="$2";

# TODO
```

### `scheduler-is-deployed`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to check if an app has been deployed
- Invoked by: `dokku ps:rebuild`
- Arguments: `$DOKKU_SCHEDULER $APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"; APP="$2";

# TODO
```
### `scheduler-logs`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to run scheduler commands when retrieving container logs
- Invoked by: `dokku logs:failed`
- Arguments: `$DOKKU_SCHEDULER $APP $PROCESS_TYPE $TAIL $PRETTY_PRINT $NUM`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"; APP="$2"; PROCESS_TYPE="$3"; TAIL="$4"; PRETTY_PRINT="$5"; NUM="$6"

# TODO
```

### `scheduler-logs-failed`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to run scheduler commands when retrieving failed container logs
- Invoked by: `dokku logs:failed`
- Arguments: `$DOKKU_SCHEDULER $APP`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"; APP="$2";

# TODO
```

### `scheduler-post-delete`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to run scheduler commands when an app is deleted
- Invoked by: `dokku apps:destroy`
- Arguments: `$DOKKU_SCHEDULER $APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"; APP="$2"; IMAGE_TAG="$3";

# TODO
```

### `scheduler-retire`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to run scheduler commands when containers should be force retired from the system
- Invoked by: `dokku run`
- Arguments: `$DOKKU_SCHEDULER`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1";

# TODO
```

### `scheduler-run`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to run scheduler commands when a command is executed for your app
- Invoked by: `dokku run`
- Arguments: `$DOKKU_SCHEDULER $APP ...ARGS`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"; APP="$2"; ARGS="${@:3}";

# TODO
```

### `scheduler-stop`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to run scheduler commands when a tag is destroyed
- Invoked by: `dokku apps:destroy, dokku ps:stop`
- Arguments: `$DOKKU_SCHEDULER $APP $REMOVE_CONTAINERS`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"; APP="$2"; REMOVE_CONTAINERS="$3";

# TODO
```

### `scheduler-tags-create`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to run scheduler commands when a tag is created
- Invoked by: `dokku tags:create`
- Arguments: `$DOKKU_SCHEDULER $APP $SOURCE_IMAGE $TARGET_IMAGE`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1" APP="$2" SOURCE_IMAGE="$3" TARGET_IMAGE="$4"

# TODO
```

### `scheduler-tags-destroy`

> Warning: The scheduler plugin trigger apis are under development and may change
> between minor releases until the 1.0 release.

- Description: Allows you to run scheduler commands when a tag is destroyed
- Invoked by: `dokku tags:destroy`
- Arguments: `$DOKKU_SCHEDULER $APP $IMAGE_REPO $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
DOKKU_SCHEDULER="$1"; APP="$2"; IMAGE_REPO="$3"; IMAGE_TAG="$4";

# TODO
```

### `tags-create`

- Description: Allows you to run commands once a tag for an app image has been added
- Invoked by: `dokku tags:create`
- Arguments: `$APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash
# Upload an app image to docker hub

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
APP="$1"; IMAGE_TAG="$2"; IMAGE=$(get_app_image_name $APP $IMAGE_TAG)

IMAGE_ID=$(docker inspect --format '{{ .Id }}' $IMAGE)
docker tag -f $IMAGE_ID $DOCKER_HUB_USER/$APP:$IMAGE_TAG
docker push $DOCKER_HUB_USER/$APP:$IMAGE_TAG
```

### `tags-destroy`

- Description: Allows you to run commands once a tag for an app image has been removed
- Invoked by: `dokku tags:destroy`
- Arguments: `$APP $IMAGE_TAG`
- Example:

```shell
#!/usr/bin/env bash
# Remove an image tag from docker hub

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
APP="$1"; IMAGE_TAG="$2"

# some code to remove a docker hub tag because it's not implemented in the CLI...
```

### `uninstall`

 - Description: Used to cleanup after itself.
 - Invoked by: `dokku plugin:uninstall`
 - Arguments: `$PLUGIN`
 - Example:

```shell
#!/usr/bin/env bash
# Cleanup up extra containers created

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

PLUGIN="$1"

[[ "$PLUGIN" = "my-plugin" ]] && docker rmi -f "${PLUGIN_IMAGE_DEPENDENCY}"
```

 > To avoid uninstalling other plugins make sure to check the plugin name like shown in the example.

### `update`

- Description: Can be used to run plugin updates on a regular interval. You can schedule the invoker in a cron-task to ensure your system gets regular updates.
- Invoked by: `dokku plugin:update`.
- Arguments: None
- Example:

```shell
#!/usr/bin/env bash
# Update the herokuish image from git source

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

cd /root/dokku
sudo BUILD_STACK=true make install
```

### `user-auth`

This is a special plugin trigger that is executed on *every* command run. As Dokku sometimes internally invokes the `dokku` command, special care should be taken to properly handle internal command redirects.

Note that the trigger should exit as follows:

- `0` to continue running as normal
- `1` to halt execution of the command

The `SSH_USER` is the original ssh user. If you are running remote commands, this user will typically be `dokku`, and as such should not be trusted when checking permissions. If you are connected via ssh as a different user who then invokes `dokku`, the value of this variable will be that user's name (`root`, `myuser`, etc.).

The `SSH_NAME` is the `NAME` variable set via the `sshcommand acl-add` command. If you have set a user via the `dokku-installer`, this value will be set to `admin`. For installs via debian package, this value *may* be `default`. For reference, the following command can be run as the root user to specify a specific `NAME` for a given ssh key:

```shell
sshcommand acl-add dokku NAME < $PATH_TO_SSH_KEY
```

Note that the `NAME` value is set at the first ssh key match. If an ssh key is set in the `/home/dokku/.ssh/authorized_keys` multiple times, the first match will decide the value.

- Description: Allows you to deny access to a Dokku command by either ssh user or associated ssh-command NAME user.
- Invoked by: `dokku`
- Arguments: `$SSH_USER $SSH_NAME $DOKKU_COMMAND`
- Example:

```shell
#!/usr/bin/env bash
# Allow root/admin users to do everything
# Deny plugin access to default users
# Allow access to all other commands

set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

SSH_USER=$1
SSH_NAME=$2
shift 2
[[ "$SSH_USER" == "root" ]] && exit 0
[[ "$SSH_NAME" == "admin" ]] && exit 0
[[ "$SSH_NAME" == "default" && $1 == plugin:* ]] && exit 1
exit 0
```
