# Process and Container Management

> New as of 0.3.14, Enhanced in 0.7.0

```
ps <app>                                       # List processes running in app container(s)
ps:inspect <app>                               # Displays a sanitized version of docker inspect for an app
ps:rebuild <app>                               # Rebuild an app from source
ps:rebuildall                                  # Rebuild all apps from source
ps:report [<app>] [<flag>]                     # Displays a process report for one or more apps
ps:restart <app>                               # Restart app container(s)
ps:restart-policy <app>                        # Shows the restart-policy for an app
ps:restartall                                  # Restart all deployed app containers
ps:scale <app> <proc>=<count> [<proc>=<count>] # Get/Set how many instances of a given process to run
ps:set-restart-policy <app> <policy>           # Sets app restart-policy
ps:start <app>                                 # Start app container(s)
ps:startall                                    # Start all deployed app containers
ps:stop <app>                                  # Stop app container(s)
ps:stopall                                     # Stop all app container(s)
```

By default, Dokku will only start a single `web` process - if defined - though process scaling can be managed by the `ps` plugin or [via a custom `DOKKU_SCALE` file](/docs/deployment/process-management.md#manually-managing-process-scaling).

> The `web` proctype is the only proctype that will invoke custom checks as defined by a CHECKS file. It is also the only process type that will be launched in a container that is either proxied via nginx or bound to an external port.

## Usage

### Listing processes running in an app container

To find out if your application's containers are running the commands you expect, simply run the `ps` command against that application.

```shell
dokku ps node-js-app
```

### Inspecting app containers

> New as of 0.13.0

A common administrative task to perform is calling `docker inspect` on the containers that are running for an application. This can be an error-prone task to perform, and may also reveal sensitive environment variables if not done correctly. Dokku provides a wrapper around this command via the `ps:inspect` subcommand:

```shell
dokku ps:inspect node-js-app
```

This command will gather all the running container IDs for your application and call `docker inspect`, sanitizing the output data so it can be copy-pasted elsewhere safely.

### Rebuilding applications

There are some Dokku commands which will not automatically rebuild an application's environment, or which can be told to skip a rebuild.  For instance, you may wish to run multiple `config:set` commands without a restart so as to speed up configuration. In these cases, you can ultimately trigger an application rebuild using `ps:rebuild`

```shell
dokku ps:rebuild node-js-app
```

You may also rebuild all applications at once, which is useful when enabling/disabling a plugin that modifies all applications:

```shell
dokku ps:rebuildall
```

> The `ps:rebuild` and `ps:rebuildall` commands only work for applications for which there is a source, and thus
> will only always work deterministically for git-deployed application. Please see
> the [images documentation](/docs/deployment/methods/images.md) and [tar documentation](/docs/deployment/methods/tar.md)
> in for more information concerning rebuilding those applications.

### Restarting applications

Applications can be restarted, which is functionally identical to calling the `release_and_deploy` function on an application. Please not that any linked containers *must* be started before the application in order to have a successful boot.

```shell
dokku ps:restart node-js-app
```

You may also trigger a restart on all applications at one time:

```shell
dokku ps:restartall
```

### `ps:scale` command

Dokku can also manage scaling itself via the `ps:scale` command. This command can be used to scale multiple process types at the same time.

```shell
dokku ps:scale node-js-app web=1
```

Multiple process types can be scaled at once:

```shell
dokku ps:scale node-js-app web=1 worker=1
```

Issuing the `ps:scale` command with no process type argument will output the current scaling settings for an application:

```shell
dokku ps:scale node-js-app
```

```
-----> Scaling for node-js-app
-----> proctype           qty
-----> --------           ---
-----> web                1
-----> worker             1
```

### Stopping applications

Deployed applications can be stopped using the `ps:stop` command. This turns off all running containers for an application, and will result in a `502 Bad Gateway` response for the default nginx proxy implementation.

```shell
dokku ps:stop node-js-app
```

You may also stop all applications at once:

```shell
dokku ps:stopall
```

### Starting applications

All stopped containers can be started using the `ps:start` command. This is similar to running `ps:restart`, except no action will be taken if the application containers are running.

```shell
dokku ps:start node-js-app
```

### Starting all applications

In some cases, it may be necessary to start all applications from scratch - eg. if all Docker containers have been manually stopped. This can be executed via the `ps:startall` command, which supports parallelism in the same manner `ps:rebuildall`, `ps:restartall`, and `ps:stopall` do.

Be aware that no action will be taken if the application containers are running.

```shell
dokku ps:startall
```

## Restart Policies

> New as of 0.7.0

By default, Dokku will automatically restart containers that exit with a non-zero status up to 10 times via the [on-failure Docker restart policy](https://docs.docker.com/engine/reference/run/#restart-policies-restart).

### Showing the current restart policy

The `ps:restart-policy` command will show the currently configured restart policy for an application. The default policy is `on-failure:10`

```shell
dokku ps:restart-policy node-js-app
```

```
=====> node-js-app restart-policy:
on-failure:10
```

### Setting the restart policy

You can configure this via the `ps:set-restart-policy` command:

```shell
# always restart an exited container
dokku ps:set-restart-policy node-js-app always
```

```
-----> Setting restart policy: always
```

```shell
# never restart an exited container
dokku ps:set-restart-policy node-js-app no
```

```
-----> Setting restart policy: no
```

```shell
# only restart it on Docker restart if it was not manually stopped
dokku ps:set-restart-policy node-js-app unless-stopped
```

```
-----> Setting restart policy: unless-stopped
```

```shell
# restart only on non-zero exit status
dokku ps:set-restart-policy node-js-app on-failure
```

```
-----> Setting restart policy: on-failure
```

```shell
# restart only on non-zero exit status up to 20 times
dokku ps:set-restart-policy node-js-app on-failure:20
```

```
-----> Setting restart policy: on-failure:20
```

Restart policies have no bearing on server reboot, and Dokku will always attempt to restart your applications at that point unless they were manually stopped.

## Manually managing process scaling

You can optionally _commit_ a `DOKKU_SCALE` file to the root of your repository - *not* to the /home/dokku/APP directory. Dokku expects this file to contain one line for every process defined in your Procfile.

Example:

```Procfile
web=1
worker=2
```

If it is not committed to the repository, the `DOKKU_SCALE` file will otherwise be automatically generated based on your `ps:scale` settings.

> *NOTE*: Dokku will always use the DOKKU_SCALE file that ships with the repo to override any local settings.
