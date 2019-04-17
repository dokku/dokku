# Plugin creation

A plugin can be a simple implementation of [triggers](/docs/development/plugin-triggers.md) or can implement a command structure of its own. Dokku has no restrictions on the language in which a plugin is implemented; it only cares that the plugin implements the appropriate [commands](/docs/development/plugin-creation.md#command-api) or [triggers](/docs/development/plugin-triggers.md) for the API. **NOTE:** any file that implements triggers or uses the command API must be executable.

If you create your own plugin:

1. Take a look at [the plugins shipped with Dokku](/docs/community/plugins.md) and hack away!
2. Check out the [list of triggers](/docs/development/plugin-triggers.md) your plugin can implement
3. Upload your plugin to GitHub with a repository name following the `dokku-<name>` convention (e.g. `dokku-mariadb`)
4. Edit [this page](/docs/community/plugins.md) and add a link to your plugin
5. Subscribe to the [dokku development blog](http://progrium.com) to be notified about API changes and releases


## Compilable plugins (Golang, Java(?), C, etc.)
When developing a plugin, you must implement the `install` trigger such that it outputs the built executable(s) using a directory structure that implements the plugin's desired command and/or triggers the API. See the [smoke-test-plugin](https://github.com/dokku/smoke-test-plugin) for an example.


## Command API
There are 3 main integration points: `commands`, `subcommands/default`, and `subcommands/<command-name>`.

### `commands`
Primarily used to supply the plugin's usage/help output. (i.e. [plugin help](https://github.com/dokku/dokku/tree/master/plugins/plugin/commands)).

### `subcommands/default`
Implements the plugin's default command behavior. (i.e. [`dokku plugin`](https://github.com/dokku/dokku/tree/master/plugins/plugin/subcommands/default)).

### `subcommands/<command-name>`
Implements the additional command interface and will translate to `dokku plugin:cmd` on the command line. (i.e. [`dokku plugin:install`](https://github.com/dokku/dokku/tree/master/plugins/plugin/subcommands/install)).


# Sample plugin
The below plugin is a dummy `dokku hello` plugin.

`hello/subcommands/default`

```shell
#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"

hello_main_cmd() {
  declare desc="prints Hello \$APP"
  local cmd="hello"
  # Support --app/$DOKKU_APP_NAME flag
  # Use the following lines to reorder args into "$cmd $DOKKU_APP_NAME $@""
  local argv=("$@")
  [[ ${argv[0]} == "$cmd" ]] && shift 1
  [[ -n $DOKKU_APP_NAME ]] && set -- $DOKKU_APP_NAME $@
  set -- $cmd $@
  ##

  [[ -z $2 ]] && dokku_log_fail "Please specify an app to run the command on"
  verify_app_name "$2"
  local APP="$2";

  echo "Hello $APP"
}

hello_main_cmd "$@"
```

`hello/subcommands/world`

```shell
#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_CORE_AVAILABLE_PATH/common/functions"

hello_world_cmd() {
  declare desc="prints Hello World"
  local cmd="hello:world"
  # Support --app/$DOKKU_APP_NAME flag
  # Use the following lines to reorder args into "$cmd $DOKKU_APP_NAME $@""
  local argv=("$@")
  [[ ${argv[0]} == "$cmd" ]] && shift 1
  [[ -n $DOKKU_APP_NAME ]] && set -- $DOKKU_APP_NAME $@
  set -- $cmd $@
  ##

  [[ -z $2 ]] && dokku_log_fail "Please specify an app to run the command on"
  verify_app_name "$2"
  local APP="$2";

  echo "Hello world"
}

hello_world_cmd "$@"
```

`hello/commands`

```shell
#!/usr/bin/env bash
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

Each plugin requires a `plugin.toml` descriptor file with the following required fields:

```toml
[plugin]
description = "dokku hello plugin"
version = "0.1.0"
[plugin.config]
```

# A few notes:

- Remember to `chmod +x` your executable files
- Always support `DOKKU_TRACE` as per the 2nd line of the above example
- If your command depends on an application, include a check for whether that application exists (see the above example)
- You must implement a `help` command, though you may leave it empty. Also, you must use commas (`,`) in the command syntax to support output in columns
- Commands **should** be namespaced
- As of 0.3.3, a catch-all should be implemented that exits with a `DOKKU_NOT_IMPLEMENTED_EXIT` code. This allows Dokku to output a `command not found` message.
- Consider whether you want to include the `set -eo pipefail` option. Look at the following example :

    ```shell
    IMAGE=$(docker images | grep "user/repo" | awk '{print $3}')
    if [[ -z $IMAGE ]]; then
        dokku_log_fail "user/repo image not found... Did you run 'dokku plugin:install'?"
    fi
    ```

  If `user/repo` doesn't exist, Dokku exits just before the `awk` command and the `dokku_log_fail` message will never go to   `STDOUT`. printed with echo. You would want to use `set -e` in this case.

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
- In the case that your plugin needs to set application configuration settings and you want to avoid having to restart (default Heroku-style behavior) these "internal" commands provide this functionality:

  ```shell
  dokku config:set --no-restart node-js-app KEY1=VALUE1 [KEY2=VALUE2 ...]
  dokku config:unset --no-restart node-js-app KEY1 [KEY2 ...]
  ```
- If you want to allow other plugins access to (some of) your plugin's functionality, you can expose this by including a `functions` file in your plugin for others to source
  - You should consider all functions in that file to be publicly accessible by other plugins
  - Any functions you want to keep private should reside in your plugin's `trigger/` or `commands/` directories
- As of 0.4.0, Dokku allows image tagging and deployment of tagged images
  - This means hard-coding the `$IMAGE` as `dokku/$APP` is no longer sufficient
  - You should now use `get_running_image_tag()` and `get_app_image_name()` as sourced from `common/functions` (see the [plugin triggers](/docs/development/plugin-triggers.md) doc for examples). **Note:** This is only for plugins that are not `pre/post-build-*` plugins
- As of 0.5.0, we use container labels to help cleanup intermediate containers with `dokku cleanup
  - This means that if you manually call `docker run`, you should include `$DOKKU_GLOBAL_RUN_ARGS` to ensure your intermediate containers are labeled correctly
- As of 0.6.0, you should not **not** call the `dokku` binary directly from within plugins because clients using the `--app` argument are potentially broken when doing so (as well as other issues)
  - You should instead source the `functions` file for a given plugin when attempting to call Dokku internal functions
