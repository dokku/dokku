# Plugin creation

A plugin can be a simple implementation of [triggers](/docs/development/plugin-triggers.md) or can implement a command structure of its own. Dokku has no restrictions on the language in which a plugin is implemented; it only cares that the plugin implements the appropriate [commands](/docs/development/plugin-creation.md#command-api) or [triggers](/docs/development/plugin-triggers.md) for the API. **NOTE:** any file that implements triggers or uses the command API must be executable.

When creating custom plugins:

1. Take a look at [the plugins shipped with Dokku](/docs/community/plugins.md) and hack away!
2. Check out the [list of triggers](/docs/development/plugin-triggers.md) the plugin can implement
3. Upload the plugin to GitHub with a repository name following the `dokku-<name>` convention (e.g. `dokku-mariadb`)
4. Edit [this page](/docs/community/plugins.md) and add a link to the plugin
5. Subscribe to the [dokku development blog](http://progrium.com) to be notified about API changes and releases

## Compilable plugins (Golang, Java(?), C, etc.)

When developing a plugin, the `install` trigger must be implemented such that it outputs the built executable(s) using a directory structure that implements the plugin's desired command and/or triggers the API. See the [smoke-test-plugin](https://github.com/dokku/smoke-test-plugin) for an example.

## Command API

There are 3 main integration points: `commands`, `subcommands/default`, and `subcommands/<command-name>`.

### `commands`

Primarily used to supply the plugin's usage/help output. (i.e. [plugin help](https://github.com/dokku/dokku/tree/master/plugins/plugin/commands)).

### `subcommands/default`

Implements the plugin's default command behavior. (i.e. [`dokku plugin`](https://github.com/dokku/dokku/tree/master/plugins/plugin/subcommands/default)).

### `subcommands/<command-name>`

Implements the additional command interface and will translate to `dokku plugin:cmd` on the command line. (i.e. [`dokku plugin:install`](https://github.com/dokku/dokku/tree/master/plugins/plugin/subcommands/install)).

# Plugin Building Tips

## Always create a `plugin.toml`

The `plugin.toml` file is used to describe the plugin in help output, and helps users understand the purpose of the plugin. This _must_ have a description and a version. The version _should_ be bumped at every plugin release.

```toml
[plugin]
description = "dokku example plugin"
version = "0.1.0"
[plugin.config]
```

## Files should be executable

Commands, subcommands, triggers and source shell scripts should all be executable. On a Unix-like machine, the following command can be used to make them executable:

```shell
chmod +x path/to/file
```

Non-executable commands, subcommands, and triggers will be ignored.

## Use the `pipefail` bash option

Consider whether to include the `set -eo pipefail` option. Look at the following example:

```shell
IMAGE=$(docker images | grep "user/repo" | awk '{print $3}')
if [[ -z $IMAGE ]]; then
  dokku_log_fail "user/repo image not found... Did you run 'dokku plugin:install'?"
fi
```

If `user/repo` doesn't exist, Dokku exits just before the `awk` command and the `dokku_log_fail` message will never go to `STDOUT`. printed with echo. The `set -e` option should be used in this case.

Here is the `help` entry for `set`:

```
help set
Options:
  -e  Exit immediately if a command exits with a non-zero status.
  -o option-name
      pipefail     the return value of a pipeline is the status of
                   the last command to exit with a non-zero status,
                   or zero if no command exited with a non-zero status
```

## Support trace mode

Trace mode is useful for getting debugging output from plugins when the `--trace` flag is specified or `dokku trace:on` is triggered. This should be done at the top of each shell script:

```shell
#!/usr/bin/env bash
set -eo pipefail
[[ $DOKKU_TRACE ]] && set -x
```

In the above example, the third line enables bash's debug mode, which prints command traces before executing command.

## Verify the existence of dependencies

If a plugin depends on a specific command-line tool, check whether that tool exists before utilizing it. Either `command -v` or `which` may be used to do so:

```shell
# `command -v` example
if ! command -v "nginx" &>/dev/null; then
  log-fail "Missing nginx, install it"
fi

# `which` example
if ! which nginx >/dev/null 2>&1; then
  log-fail "Missing nginx, install it"
fi
```

In cases where a dependency should be installed before the plugin can be used at all, use the `dependencies` plugin trigger to install the dependency.

## Implement a help command

For plugins which expose commands, implement a `help` command. This may be empty, but should contain a listing of all available commands.

Commas - `,` - are used in the help output for columnizing output. Verify that the plugin conforms to the spec by running `dokku help --all` and manually verifying the output.

See the sample plugin below for an example.

## Namespace commands

All commands *should* be namespaced. In cases where a core plugin is overriden, the plugin _may_ utilize the a namespace in use by the core, but generally this should be avoided to reduce confusion as to where the command is implemented.

## Implement a proper catch-all command

As of 0.3.3, a catch-all should be implemented that exits with a `DOKKU_NOT_IMPLEMENTED_EXIT` code. This allows Dokku to output a `command not found` message.

See the sample plugin below for an example.

## Set app config without restarting

In the case that a plugin needs to set app configuration settings and a restart should be avoided (default Heroku-style behavior) these "internal" commands provide this functionality:

```shell
config_set --no-restart node-js-app KEY1=VALUE1 [KEY2=VALUE2 ...]
config_unset --no-restart node-js-app KEY1 [KEY2 ...]
```

## Expose functionality in a `functions` file

To allow other plugins access to (some of) a plugin's functionality, functions can expose by including a `functions` file in the plugin for others to source. All functions in that file should be considered publicly accessible by other plugins.

Any functions that must be kept private should reside in the plugin's `trigger/` or `commands/` directories. Other files may also be used to hide private functions; the official convention for hiding private functions is to place them an `internal-functions` file.

## Use helper functions to fetch app images

> New as of 0.4.0

Dokku allows image tagging and deployment of tagged images. This means hard-coding the `$IMAGE` as `dokku/$APP` is no longer sufficient.

Plugins should use `get_running_image_tag()` and `get_app_image_name()` as sourced from `common/functions`. See the [plugin triggers](/docs/development/plugin-triggers.md) doc for examples.

> **Note:** This is only for plugins that are not `pre/post-build-*` plugins

## Use `$DOCKER_BIN` instead of `docker` directly

> New as of 0.17.5

Certain systems may require a wrapper function around the `docker` binary for proper execution. Utilizing the `$DOCKER_BIN` environment variable when calling docker for those functions is preferred.

```shell
# good
"$DOCKER_BIN" container run -d $IMAGE /bin/bash -e -c "$COMMAND"

# bad
docker run -d $IMAGE /bin/bash -e -c "$COMMAND"
```

## Include labels for all temporary containers and images

> New as of 0.5.0

As of 0.5.0, labels are used to help cleanup intermediate containers with `dokku cleanup`. Plugins that create containers and images should add the correct labels to the `build`, `commit`, and `run` docker commands.

Note that where possible, a label `com.dokku.app-name=$APP` - where `$APP` is the name of the app - should also be included. This enables `dokku cleanup APP` to cleanup the specific containers for a given app.

```shell
# `docker build` example
"$DOCKER_BIN" image build "--label=com.dokku.app-name=${APP}" $DOKKU_GLOBAL_BUILD_ARGS ...

# `docker commit` example
# Note that the arguments must be set as a local array
# as arrays cannot be exported in shell
local DOKKU_COMMIT_ARGS=("--change" "LABEL org.label-schema.schema-version=1.0" "--change" "LABEL org.label-schema.vendor=dokku" "--change" "LABEL $DOKKU_CONTAINER_LABEL=")
"$DOCKER_BIN" container commit --change "LABEL com.dokku.app-name=$APP" "${DOKKU_COMMIT_ARGS[@]}" ...

# `docker run` example
"$DOCKER_BIN" container run "--label=com.dokku.app-name=${APP}" $DOKKU_GLOBAL_RUN_ARGS ...
```

## Copy files from the built image using `copy_from_image`

Avoid copying files from running containers as these files may change over time. Instead copy files from the image built during the deploy process. This can be done via the `copy_from_image` helper function. This will correctly handle various corner cases in copying files from an image.

```shell
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"

local TMP_FILE=$(mktemp "/tmp/dokku-${DOKKU_PID}-${FUNCNAME[0]}.XXXXXX")
trap "rm -rf '$TMP_FILE' >/dev/null" RETURN INT TERM

local IMAGE_TAG="$(get_running_image_tag "$APP")"
local IMAGE=$(get_deploying_app_image_name "$APP" "$IMAGE_TAG")
copy_from_image "$IMAGE" "file-being-copied" "$TMP_FILE" 2>/dev/null
```

Files are copied from the `/app` directory - for images built via buildpacks - or `WORKDIR` - for images built via Dockerfile.

## Avoid calling the `dokku` binary directly

> New as of 0.6.0

Plugins should **not** call the `dokku` binary directly from within plugins because clients using the `--app` argument are potentially broken when doing so.

Plugins should instead source the `functions` file for a given plugin when attempting to call Dokku internal functions.

# Sample plugin

The below plugin is a dummy `dokku hello` plugin.

Each plugin requires a `plugin.toml` descriptor file with the following required fields:

```toml
[plugin]
description = "dokku hello plugin"
version = "0.1.0"
[plugin.config]
```

`hello/subcommands/default`

```shell
#!/usr/bin/env bash
set -eo pipefail
[[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"

cmd-hello-default() {
  declare desc="prints Hello \$APP"
  declare cmd="hello"
  [[ "$1" == "$cmd" ]] && shift 1
  # Support --app/$DOKKU_APP_NAME flag
  # Use the following lines to reorder args into "$cmd $DOKKU_APP_NAME $@""
  [[ -n $DOKKU_APP_NAME ]] && set -- $DOKKU_APP_NAME $@
  set -- $cmd $@
  #
  declare APP="$1"

  [[ -z "$APP" ]] && dokku_log_fail "Please specify an app to run the command on"
  verify_app_name "$APP"

  echo "Hello $APP"
}

cmd-hello-default "$@"
```

`hello/subcommands/world`

```shell
#!/usr/bin/env bash
set -eo pipefail
[[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"

cmd-hello-world() {
  declare desc="prints Hello world"
  declare cmd="hello:world"
  [[ "$1" == "$cmd" ]] && shift 1

  echo "Hello world"
}

cmd-hello-world "$@"
```

`hello/commands`

```shell
#!/usr/bin/env bash
set -eo pipefail
[[ $DOKKU_TRACE ]] && set -x

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
