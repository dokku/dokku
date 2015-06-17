docker-options
========================

Usage
-----

```bash
$ dokku help
...
    docker-options <app>                               Display apps docker options for all phases
    docker-options <app> <phase(s)>                    Display apps docker options for phase (comma seperated phase list)
    docker-options:add <app> <phase(s)> OPTION         Add docker option to app for phase (comma seperated phase list)
    docker-options:remove <app> <phase(s)> OPTION      Remove docker option from app for phase (comma seperated phase list)
...
````

Add some options (first, for the deployed/running app and second when executing `dokku run`)

```bash
$ dokku docker-options:add myapp deploy "-v /host/path:/container/path"
$ dokku docker-options:add myapp run "-v /another/container/path"
```

Check what we added

```bash
$ dokku docker-options myapp
Deploy options:
    -v /host/path:/container/path
Run options:
    -v /another/container/path
```

Remove an option
```bash
$ dokku docker-options:remove myapp run "--link container_name:alias"
```

Note about `dokku` phases and `docker-options`
------------
`dokku` deploys your application in multiple "phases" and the `docker-options` plugin allows you to pass arguments to the underlying docker container in the following 3 phases/containers
- `build`: the container that executes the appropriate buildpack
- `deploy`: the container that executes your running/deployed application
- `run`: the container that executes any arbitrary command via `dokku run myapp`


Advanced Usage (avoid if possible)
------------

In your applications folder (`/home/dokku/app_name`) create a file called `DOCKER_OPTIONS_RUN` (or `DOCKER_OPTIONS_BUILD` or `DOCKER_OPTIONS_DEPLOY`).

Inside this file list one docker option per line. For example:

```bash
--link container_name:alias
-v /host/path:/container/path
-v /another/container/path
```

The above example will result in the following options being passed to docker during `dokku run`:

```bash
--link container_name:alias -v /host/path:/container/path -v /another/container/path
```

You may also include comments (lines beginning with a #) and blank lines in the DOCKER_OPTIONS file.

Move information on docker options can be found here: http://docs.docker.io/en/latest/reference/run/ .
