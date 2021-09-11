# Entering containers

> New as of 0.4.0

```
enter <app>  [<container-type> || --container-id <container-id>]  # Connect to a specific app container
```

## Usage

The `enter` command can be used to enter a running container. The following variations of the command exist:

```shell
# enter the web process
dokku enter node-js-app web

# enter the first web process
dokku enter node-js-app web.1

# enter a process for an app by container ID
dokku enter node-js-app --container-id ID
```

The `container-type` argument can be one either:

- If your app has a `Procfile`, the name of a process type in your `Procfile`.
- If your app has no `Procfile`, the word `web`.

If the specified process type is scaled up to more than one container, then the first container will be automatically selected. this can be overriden by specifying an integer index denoting the desired container, where the first container's index is `1`.

Additionally, the `enter` command can be executed with no `<container-type>`. If only a single `<container-type>` is defined in the app's Procfile, executing `enter` will drop the terminal into the only running container. This behavior is not supported when specifying a custom command; as described below.

By default, `dokku enter` will run a `/bin/bash`, but can also be used to run custom commands:

```shell
# just echo hi
dokku enter node-js-app web echo hi

# run a long-running command, as one might for a cron task
dokku enter node-js-app web python script/background-worker.py
```
