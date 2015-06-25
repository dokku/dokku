# Plugin Creation

If you create your own plugin:

1. Take a look at the plugins shipped with dokku and hack away!
2. Check out the [list of triggers](http://progrium.viewdocs.io/dokku/development/plughooks) your plugin can implement.
3. Upload your plugin to github with a repository name in form of `dokku-<name>` (e.g. `dokku-mariadb`)
4. Edit [this page](http://progrium.viewdocs.io/dokku/plugins) and add a link to it.
5. Subscribe to the [dokku development blog](http://progrium.com) to be notified about API changes and releases

### Sample plugin

The below plugin is a dummy `dokku hello` plugin. If your plugin exposes commands, this is a good template for your `commands` file:

```shell
#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$PLUGIN_PATH/common/functions"

case "$1" in
  hello)
    [[ -z $2 ]] && dokku_log_fail "Please specify an app to run the command on"
    APP="$2"; IMAGE_TAG=$(get_running_image_tag $APP); IMAGE=$(get_app_image_name $APP $IMAGE_TAG)
    verify_app_name "$APP"

    echo "Hello $APP"
    ;;

  hello:world)
    echo "Hello world"
    ;;

  help)
    cat<<EOF
    hello <app>, Says "Hello <app>"
    hello:world, Says "Hello world"
EOF
    ;;

  *)
    exit $DOKKU_NOT_IMPLEMENTED_EXIT
    ;;

esac
```

Each plugn requires a `plugin.toml` descriptor file with the following required fields:

```shell
[plugin]
description = "dokku hello plugin"
version = "0.1.0"
[plugin.config]
```

A few notes:

- You should always support `DOKKU_TRACE` as specified on the 2nd line of the plugin.
- If your command requires that an application exists, ensure you check for it's existence in the manner prescribed above.
- A `help` command is required, though it is allowed to be empty. Also, the command syntax will need to separated by `, ` in order to maintain columnar output alignment.
- Commands *should* be namespaced.
- As of 0.3.3, a catch-all should be implemented which exits with a `DOKKU_NOT_IMPLEMENTED_EXIT` code. This allows dokku to output a `command not found` message.
- Be sure you want the "set -eo pipefail" option. Look at the following example :

    ```shell
    IMAGE=$(docker images | grep "user/repo" | awk '{print $3}')
    if [[ -z $IMAGE ]]; then
        dokku_log_fail "user/repo image not found... Did you run 'dokku plugins-install'?"
    fi
    ```

  In the case where the "user/repo" is not installed, dokku exits just before the awk command,
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
  dokku config:set --no-restart APP KEY1=VALUE1 [KEY2=VALUE2 ...]
  dokku config:unset --no-restart APP KEY1 [KEY2 ...]
  ```
- From time to time you may want to allow other plugins access to (some of) your plugin's functionality. You can expose this by including a `functions` file in your plugin for others to source. Consider all functions in that file to be publicly accessible by other plugins. Any functions not wished to be made "public" should reside within your pluginhook or commands files.
- As of 0.4.0, we allow image tagging and deployment of said tagged images. Therefore, hard-coding of `$IMAGE` as `dokku/$APP` is no longer sufficient. Instead, for non `pre/post-build-*` plugins, use `get_running_image_tag()` & `get_app_image_name()` as sourced from common/functions. See [pluginhooks](http://progrium.viewdocs.io/dokku/development/pluginhooks) doc for examples.
