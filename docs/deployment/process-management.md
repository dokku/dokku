# Process and Container Management

> New as of 0.3.14, Enhanced in 0.7.0

```
ps <app>                                       # List processes running in app container(s)
ps:rebuildall                                  # Rebuild all apps
ps:rebuild <app>                               # Rebuild an app
ps:restartall                                  # Restart all deployed app containers
ps:restart <app>                               # Restart app container(s)
ps:scale <app> <proc>=<count> [<proc>=<count>] # Set how many processes of a given process to run
ps:start <app>                                 # Start app container(s)
ps:stop <app>                                  # Stop app container(s)
ps:restart-policy <app>                        # Shows the restart-policy for an app
ps:set-restart-policy <app> <policy>           # Sets app restart-policy
```

By default, Dokku will only start a single `web` process - if defined - though process scaling can be managed by the `ps` plugin or via a custom `DOKKU_SCALE` file.

> The `web` proctype is the only proctype that will invoke custom checks as defined by a CHECKS file. It is also the only process type that will be launched in a container that is either proxied via nginx or bound to an external port.

### `ps:scale` command

Dokku can also manage scaling itself via the `ps:scale` command. This command can be used to scale multiple process types at the same time.

```shell
dokku ps:scale APP web=1 worker=2
```

### DOKKU_SCALE file

You can optionally create a `DOKKU_SCALE` file in the root of your repository. Dokku expects this file to contain one line for every process defined in your Procfile.

Example:

```Procfile
web=1
worker=2
```

> *NOTE*: Dokku will always use the DOKKU_SCALE file that ships with the repo to override any local settings.

## Restart Policies

> New as of 0.7.0

By default, Dokku will automatically restart containers that exit with a non-zero status up to 10 times via the [on-failure Docker restart policy](https://docs.docker.com/engine/reference/run/#restart-policies-restart). You can configure this via the relevant `ps` commands:

```shell
# always restart an exited container
dokku ps:set-restart-policy node-js-app always

# never restart an exited container
dokku ps:set-restart-policy node-js-app no

# only restart it on Docker restart if it was not manually stopped
dokku ps:set-restart-policy node-js-app unless-stopped

# restart only on non-zero exit status
dokku ps:set-restart-policy node-js-app on-failure

# restart only on non-zero exit status up to 20 times
dokku ps:set-restart-policy node-js-app on-failure:20
```

Restart policies have no bearing on server reboot, and Dokku will always attempt to restart your applications at that point unless they were manually stopped.
