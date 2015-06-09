# Plugin Creation

If you create your own plugin:

1. Take a look at the plugins shipped with dokku and hack away!
2. Check out the [list of hooks](http://progrium.viewdocs.io/dokku/development/pluginhooks) your plugin can implement.
3. Upload your plugin to github with a repository name in form of `dokku-<name>` (e.g. `dokku-mariadb`)
4. Edit [this page](http://progrium.viewdocs.io/dokku/plugins) and add a link to it.
5. Subscribe to the [dokku development blog](http://progrium.com) to be notified about API changes and releases

### Sample plugin

The below plugin is a dummy `dokku hello` plugin. If your plugin exposes commands, this is a good template for your `commands` file:

```shell
#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x
source "$(dirname $0)/../common/functions"

case "$1" in
  hello)
    [[ -z $2 ]] && echo "Please specify an app to run the command on" && exit 1
    verify_app_name "$2"
    APP="$2";

    echo "Hello $APP"
    ;;

  hello:world)
    echo "Hello world"
    ;;

  help)
    cat && cat<<EOF
    hello <app>                                     Says "Hello <app>"
    hello:world                                     Says "Hello world"
EOF
    ;;

  *)
    exit $DOKKU_NOT_IMPLEMENTED_EXIT
    ;;

esac
```

A few notes:

- You should always support `DOKKU_TRACE` as specified on the 2nd line of the plugin.
- If your command requires that an application exists, ensure you check for it's existence in the manner prescribed above.
- A `help` command is required, though it is allowed to be empty.
- Commands *should* be namespaced.
- As of 0.3.3, a catch-all should be implemented which exits with a `DOKKU_NOT_IMPLEMENTED_EXIT` code. This allows dokku to output a `command not found` message.
- Be sure you want the "set -eo pipefail" option. Look at the following example :

    ```shell
    IMAGE=$(docker images | grep "user/repo" | awk '{print $3}')
    if [[ -z $IMAGE ]]; then
        echo "user/repo image not found... Did you run 'dokku plugins-install'?"
        exit 1
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
  dokku config:set-norestart APP KEY1=VALUE1 [KEY2=VALUE2 ...]
  dokku config:unset-norestart APP KEY1 [KEY2 ...]
  ```
