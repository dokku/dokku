# Process and Container Management

> New as of 0.3.14

```
ps <app>                                       # List processes running in app container(s)
ps:rebuildall                                  # Rebuild all apps
ps:rebuild <app>                               # Rebuild an app
ps:restartall                                  # Restart all deployed app containers
ps:restart <app>                               # Restart app container(s)
ps:scale <app> <proc>=<count> [<proc>=<count>] # Set how many processes of a given process to run
ps:start <app>                                 # Start app container(s)
ps:stop <app>                                  # Stop app container(s)
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
