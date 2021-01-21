# Entering containers

> New as of 0.4.0

```
enter <app>  [<container-type> || --container-id <container-id>]  # Connect to a specific app container
```

## Usage

The `enter` command can be used to enter a running container. The following variations of the command exist:

```shell
dokku enter node-js-app web
dokku enter node-js-app web.1
dokku enter node-js-app --container-id ID
```

Additionally, the `enter` command can be executed with no `<container-type>`. If only a single `<container-type>` is defined in the app's Procfile, executing `enter` will drop the terminal into the only running container. This behavior is not supported when specifying a custom command; as described below.

By default, it runs a `/bin/bash`, but can also be used to run a custom command:

```shell
# just echo hi
dokku enter node-js-app web echo hi

# run a long-running command, as one might for a cron task
dokku enter node-js-app web python script/background-worker.py
```
