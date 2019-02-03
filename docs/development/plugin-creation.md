# Plugin Creation

Plugins can simply be implementations of [triggers](/docs/development/plugin-triggers.md) or they may implement a command structure of their own. Dokku has no restrictions on the language in which a plugin is implemented, it only cares that the plugin implements the appropriate [commands](/docs/development/plugin-creation.md#command-api) or [triggers](/docs/development/plugin-triggers.md) API. NOTE: all files that implement the triggers or commands API must be executable.

If you create your own plugin:

1. Take a look at the plugins shipped with Dokku and hack away!
2. Check out the [list of triggers](/docs/development/plugin-triggers.md) your plugin can implement.
3. Upload your plugin to github with a repository name in form of `dokku-<name>` (e.g. `dokku-mariadb`)
4. Edit [this page](/docs/community/plugins.md) and add a link to it.
5. Subscribe to the [dokku development blog](http://progrium.com) to be notified about API changes and releases


### Compilable Plugins (i.e. golang, java(?), c, etc.)
The plugin developer is required to implement the `install` trigger such that it will output your built executable(s) in the correct directory structure to implement the plugin's desired command and/or trigger API. See [smoke-test-plugin](https://github.com/dokku/smoke-test-plugin) for an example.


### Command API
There are 3 main implementation points: `commands`, `subcommands/default`, and `subcommands/<command-name>`

#### commands
Primarily used to supply the plugin's usage/help output. (i.e. [plugin help](https://github.com/dokku/dokku/tree/master/plugins/plugin/commands))

#### subcommands/default
Implements the plugin's default command behavior. (i.e. [`dokku plugin`](https://github.com/dokku/dokku/tree/master/plugins/plugin/subcommands/default))

#### subcommands/<command-name>
Implements the additional command interface and will translate to `dokku plugin:cmd` on the command line. (i.e. [`dokku plugin:install`](https://github.com/dokku/dokku/tree/master/plugins/plugin/subcommands/install))


### Sample plugin
The below plugin is a dummy `dokku hello` plugin.

hello/subcommands/default

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

hello/subcommands/world

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

hello/commands

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

A few notes:

- Don't forget to `chmod +x` your executable files.
- You should always support `DOKKU_TRACE` as specified on the 2nd line of the plugin.
- If your command requires that an application exists, ensure you check for it's existence in the manner prescribed above.
- A `help` command is required, though it is allowed to be empty. Also, the command syntax will need to separated by `, ` in order to maintain columnar output alignment.
- Commands *should* be namespaced.
- As of 0.3.3, a catch-all should be implemented which exits with a `DOKKU_NOT_IMPLEMENTED_EXIT` code. This allows Dokku to output a `command not found` message.
- Be sure you want the "set -eo pipefail" option. Look at the following example :

    ```shell
    IMAGE=$(docker images | grep "user/repo" | awk '{print $3}')
    if [[ -z $IMAGE ]]; then
        dokku_log_fail "user/repo image not found... Did you run 'dokku plugin:install'?"
    fi
    ```

  In the case where the "user/repo" is not installed, Dokku exits just before the awk command,
  you will never see the message printed with echo. You just want "set -e" in this case.

  Here is the documentation of the 'set -eo pipefail' option:
  ```
  help set
    Options:
      -e  Exit immediately if a command exits with a non-zero status.
      -o option-name
          pipefail     the return value of a pipeline is the status of
                       the last command to exit with a non-zero status,
                       or zero if no command exited with a non-zero status
  ```
- As some plugins require access to set app config settings and do not want/require the default Heroku-style behavior of a restart, we have the following "internal" commands that provide this functionality :

  ```shell
  dokku config:set --no-restart node-js-app KEY1=VALUE1 [KEY2=VALUE2 ...]
  dokku config:unset --no-restart node-js-app KEY1 [KEY2 ...]
  ```
- From time to time you may want to allow other plugins access to (some of) your plugin's functionality. You can expose this by including a `functions` file in your plugin for others to source. Consider all functions in that file to be publicly accessible by other plugins. Any functions not wished to be made "public" should reside within your plugin trigger or commands files.
- As of 0.4.0, we allow image tagging and deployment of said tagged images. Therefore, hard-coding of `$IMAGE` as `dokku/$APP` is no longer sufficient. Instead, for non `pre/post-build-*` plugins, use `get_running_image_tag()` & `get_app_image_name()` as sourced from common/functions. See the [plugin triggers](/docs/development/plugin-triggers.md) doc for examples.
- As of 0.5.0, we use container labels to help cleanup intermediate containers with `dokku cleanup`. If manually calling `docker run`, include `$DOKKU_GLOBAL_RUN_ARGS`. This will ensure you intermediate containers labeled correctly.
- As of 0.6.0, we advise you to *not* call the `dokku` binary directly from within plugins. Clients using the `--app` argument are potentially broken, amongst other issues, when doing so. Instead, please source the `functions` file for a given plugin when attempting to call Dokku internal functions
