# Process Management

> New as of 0.3.14, Enhanced in 0.7.0

```
ps:inspect <app>                                                  # Displays a sanitized version of docker inspect for an app
ps:rebuild [--parallel count] [--all|<app>]                       # Rebuilds an app from source
ps:report [<app>] [<flag>]                                        # Displays a process report for one or more apps
ps:restart [--parallel count] [--all|<app>]  [<process-name>]     # Restart an app
ps:restore [<app>]                                                # Start previously running apps e.g. after reboot
ps:scale [--skip-deploy] <app> <proc>=<count> [<proc>=<count>...] # Get/Set how many instances of a given process to run
ps:set <app> <key> <value>                                        # Set or clear a ps property for an app
ps:start [--parallel count] [--all|<app>]                         # Start an app
ps:stop [--parallel count] [--all|<app>]                          # Stop an app
```

## Usage

### Inspecting app containers

> New as of 0.13.0

A common administrative task to perform is calling `docker inspect` on the containers that are running for an app. This can be an error-prone task to perform, and may also reveal sensitive environment variables if not done correctly. Dokku provides a wrapper around this command via the `ps:inspect` command:

```shell
dokku ps:inspect node-js-app
```

This command will gather all the running container IDs for your app and call `docker inspect`, sanitizing the output data so it can be copy-pasted elsewhere safely.

### Rebuilding apps

It may be useful to rebuild an app at will, such as for commands that do not rebuild an app or when skipping a rebuild after setting multiple config values. For these use cases, the `ps:rebuild` function can be used.

```shell
dokku ps:rebuild node-js-app
```

All apps may be rebuilt by using the `--all` flag.

```shell
dokku ps:rebuild --all
```

By default, rebuilding all apps happens serially. The parallelism may be controlled by the `--parallel` flag.

```shell
dokku ps:rebuild --all --parallel 2
```

Finally, the number of parallel workers may be automatically set to the number of CPUs available by setting the `--parallel` flag to `-1`

```shell
dokku ps:rebuild --all --parallel -1
```

A missing linked container will result in failure to boot apps. Services should all be started for apps being rebuilt.

### Restarting apps

An app may be restarted using the `ps:restart` command.

```shell
dokku ps:restart node-js-app
```

A single process type - such as `web` or `worker` - may also be specified. This _does not_ support specifying a given instance of a process type, and only supports restarting all instances of that process type.

```shell
dokku ps:restart node-js-app web
```

All apps may be restarted by using the `--all` flag. This flag is incompatible with specifying a process type.

```shell
dokku ps:restart --all
```

By default, restarting all apps happens serially. The parallelism may be controlled by the `--parallel` flag.

```shell
dokku ps:restart --all --parallel 2
```

Finally, the number of parallel workers may be automatically set to the number of CPUs available by setting the `--parallel` flag to `-1`

```shell
dokku ps:restart --all --parallel -1
```

A missing linked container will result in failure to boot apps. Services should all be started for apps being rebuilt.

### Displaying existing scale properties

Issuing the `ps:scale` command with no arguments will output the current scaling properties for an app.

```shell
dokku ps:scale node-js-app
```

```
-----> Scaling for python
proctype: qty
--------: ---
web:  1
```

### Scaling apps

#### Via CLI

> This functionality is disabled if the formation is managed via the `formation` key of `app.json`.

Dokku can also manage scaling itself via the `ps:scale` command. This command can be used to scale multiple process types at the same time.

```shell
dokku ps:scale node-js-app web=1
```

Multiple process types can be scaled at once:

```shell
dokku ps:scale node-js-app web=1 worker=1
```

If desired, the corresponding deploy will be skipped by using the `--skip-deploy` flag:

```shell
dokku ps:scale --skip-deploy node-js-app web=1
```

#### Manually managing process scaling

> Using a `formation` key in an `app.json` file disables the ability to use `ps:scale` for scaling.

An `app.json` file can be committed to the root of the pushed app repository, and must be within the built image artifact in the image's working directory as shown below.

- Buildpacks: `/app/app.json`
- Dockerfile: `WORKDIR/app.json` or `/app.json` (if no working directory specified)
- Docker Image: `WORKDIR/app.json` or `/app.json` (if no working directory specified)

The `formation` key should be specified as follows in the `app.json` file:

```Procfile
{
  "formation": {
    "web": {
      "quantity": 1
    },
    "worker": {
      "quantity": 4
    }
  }
}
```

Removing the file will result in Dokku respecting the `ps:scale` command for setting scale values. The values set via the `app.json` file from a previous deploy will be respected.

#### The `web` process

For initial app deploys, Dokku will default to starting a single `web` process for each app. This process may be defined within the `Procfile` or as the `CMD` (for Dockerfile or Docker image deploys). Scaling of the `web` process - and all other processes - may be managed via `ps:scale` or the `formation` key in the `app.json` file either before or after the initial deploy.

There are also a few other exceptions for the `web` process.

- Custom checks defined by a `CHECKS` file only apply to the `web` process type.
- By default, the built-in nginx proxy implementation only proxies the `web` process (others may be handled via a custom `nginx.conf.sigil`).
  - See the [nginx request proxying documentation](/docs/configuration/nginx.md#request-proxying) for more information on how nginx handles proxied requests.
- Only the `web` process may be bound to an external port.

#### Changing the `Procfile` location

When deploying a monorepo, it may be desirable to specify the specific path of the `Procfile` file to use for a given app. This can be done via the `ps:set` command. If a value is specified and that file does not exist within the repository, Dokku will continue the build process as if the repository has no `Procfile`.

```shell
dokku ps:set node-js-app procfile-path Procfile2
```

The default value may be set by passing an empty value for the option:

```shell
dokku ps:set node-js-app procfile-path
```

The `procfile-path` property can also be set globally. The global default is `Procfile`, and the global value is used when no app-specific value is set.

```shell
dokku ps:set --global procfile-path global-Procfile
```

The default value may be set by passing an empty value for the option.

```shell
dokku ps:set --global procfile-path
```

### Stopping apps

Deployed apps can be stopped using the `ps:stop` command. This turns off all running containers for an app, and will result in a **502 Bad Gateway** response for the default nginx proxy implementation.

```shell
dokku ps:stop node-js-app
```

All apps may be stopped by using the `--all` flag.

```shell
dokku ps:stop --all
```

By default, stopping all apps happens serially. The parallelism may be controlled by the `--parallel` flag.

```shell
dokku ps:stop --all --parallel 2
```

Finally, the number of parallel workers may be automatically set to the number of CPUs available by setting the `--parallel` flag to `-1`

```shell
dokku ps:stop --all --parallel -1
```

### Starting apps

All stopped containers can be started using the `ps:start` command. This is similar to running `ps:restart`, except no action will be taken if the app containers are running.

```shell
dokku ps:start node-js-app
```

All apps may be started by using the `--all` flag.

```shell
dokku ps:start --all
```

By default, starting all apps happens serially. The parallelism may be controlled by the `--parallel` flag.

```shell
dokku ps:start --all --parallel 2
```

Finally, the number of parallel workers may be automatically set to the number of CPUs available by setting the `--parallel` flag to `-1`

```shell
dokku ps:start --all --parallel -1
```

### Restart policies

> New as of 0.7.0, Command Changed in 0.22.0

By default, Dokku will automatically restart containers that exit with a non-zero status up to 10 times via the [on-failure Docker restart policy](https://docs.docker.com/engine/reference/run/#restart-policies---restart).

#### Setting the restart policy

> A change in the restart policy must be followed by a `ps:rebuild` call.

You can configure this via the `ps:set` command:

```shell
# always restart an exited container
dokku ps:set node-js-app restart-policy always

# never restart an exited container
dokku ps:set node-js-app restart-policy no

# only restart it on Docker restart if it was not manually stopped
dokku ps:set node-js-app restart-policy unless-stopped

# restart only on non-zero exit status
dokku ps:set node-js-app restart-policy on-failure

# restart only on non-zero exit status up to 20 times
dokku ps:set node-js-app restart-policy on-failure:20
```

Restart policies have no bearing on server reboot, and Dokku will always attempt to restart your apps at that point unless they were manually stopped.

### Displaying reports for an app

> New as of 0.12.0

You can get a report about the deployed apps using the `ps:report` command:

```shell
dokku ps:report
```

```
=====> node-js-app ps information
       Deployed:                      false
       Processes:                     0
       Ps can scale:                  true
       Ps computed procfile path:     Procfile2
       Ps global procfile path:       Procfile
       Ps restart policy:             on-failure:10
       Ps procfile path:              Procfile2
       Restore:                       true
       Running:                       false
=====> python-sample ps information
       Deployed:                      false
       Processes:                     0
       Ps can scale:                  true
       Ps computed procfile path:     Procfile
       Ps global procfile path:       Procfile
       Ps restart policy:             on-failure:10
       Ps procfile path:
       Restore:                       true
       Running:                       false
=====> ruby-sample ps information
       Deployed:                      false
       Processes:                     0
       Ps can scale:                  true
       Ps computed procfile path:     Procfile
       Ps global procfile path:       Procfile
       Ps restart policy:             on-failure:10
       Ps procfile path:
       Restore:                       true
       Running:                       false
```

You can run the command for a specific app also.

```shell
dokku ps:report node-js-app
```

```
=====> node-js-app ps information
       Deployed:                      false
       Processes:                     0
       Ps can scale:                  true
       Ps restart policy:             on-failure:10
       Restore:                       true
       Running:                       false
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku ps:report node-js-app --deployed
```
