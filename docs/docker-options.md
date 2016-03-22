# Docker Container Options

> New as of 0.3.17

Pass [options](https://docs.docker.com/engine/reference/run/) to Docker during Dokku's `build`, `deploy` and `run` phases

```
docker-options <app> [phase(s)]                  Display app's Docker options for all phases (or comma separated phase list)
docker-options:add <app> <phase(s)> OPTION       Add Docker option to app for phase (comma-separated phase list)
docker-options:remove <app> <phase(s)> OPTION    Remove Docker option from app for phase (comma-separated phase list)
```

> When specifying multiple phases, they **must** be comma-separated _without_ spaces in between each phase, like so:
>
> ```
> dokku docker-options:add myapp deploy,run "-v /home/dokku/logs/myapp:/app/logs"
> ```

## About Dokku phases

Dokku deploys your application in multiple "phases" and the `docker-options` plugin allows you to pass arguments to their underlying docker container:

- `build`: the container that executes the appropriate buildpack
- `deploy`: the container that executes your running/deployed application
- `run`: the container that executes any arbitrary command via `dokku run myapp`

## Examples

### Add Docker options

Add some options for the deployed/running app and when executing [`dokku run`](deployment/one-off-processes/):

```shell
# Mount a host volume in a Docker container: "-v /host/path:/container/path"
dokku docker-options:add myapp deploy "-v /home/dokku/logs/myapp:/app/logs"
dokku docker-options:add myapp run "-v /home/dokku/logs/myapp:/app/logs"
```

> Note: When [mounting a host directory](https://docs.docker.com/engine/reference/run/#volume-shared-filesystems) in a Dokku app you should first create that directory as user `dokku` and then mount the directory under `/app` in the container using `docker-options` as above. Otherwise the app will lack write permission in the directory.

### Output Docker options

```shell
dokku docker-options myapp
# Deploy options:
#    -v /home/dokku/logs/myapp:/app/logs
# Run options:
#    -v /home/dokku/logs/myapp:/app/logs
```

### Remove a Docker option

```shell
dokku docker-options:remove myapp run "-v /home/dokku/logs/myapp:/app/logs"
```

## Advanced usage

In your applications folder `/home/dokku/app_name` create a file called `DOCKER_OPTIONS_RUN` (or `DOCKER_OPTIONS_BUILD` or `DOCKER_OPTIONS_DEPLOY`).

Inside this file list one Docker option per line. For example:

```shell
--link container_name:alias
-v /host/path:/container/path
-v /another/container/path
```

The above example will result in the following options being passed to Docker during `dokku run`:

```shell
--link container_name:alias -v /host/path:/container/path -v /another/container/path
```

You may also include comments (lines beginning with a #) and blank lines in the DOCKER_OPTIONS file.

Move information on Docker options can be found here: http://docs.docker.io/en/latest/reference/run/ .
