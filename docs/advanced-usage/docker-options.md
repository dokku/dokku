# Docker Container Options

> New as of 0.3.17

Pass [options](https://docs.docker.com/engine/reference/run/) to Docker during Dokku's `build`, `deploy` and `run` phases

```
docker-options:add <app> <phase(s)> OPTION    # Add Docker option to app for phase (comma-separated phase list)
docker-options:remove <app> <phase(s)> OPTION # Remove Docker option from app for phase (comma-separated phase list)
docker-options:report [<app>] [<flag>]        # Displays a docker options report for one or more apps
```

> When specifying multiple phases, they **must** be comma-separated _without_ spaces in between each phase, like so:
>
> ```shell
> dokku docker-options:add node-js-app deploy,run "-v /var/log/node-js-app:/app/logs"
> ```

## About Dokku phases

Dokku deploys your application in multiple "phases" and the `docker-options` plugin allows you to pass arguments to their underlying docker container:

- `build`: the container that executes the appropriate buildpack
- `deploy`: the container that executes your running/deployed application
- `run`: the container that executes any arbitrary command via `dokku run`

## Examples

### Add Docker options

Add some options for the deployed/running app and when executing [`dokku run`](/docs/deployment/one-off-processes.md):

```shell
# Mount a host volume in a Docker container: "-v /host/path:/container/path"
dokku docker-options:add node-js-app deploy "-v /var/log/node-js-app:/app/logs"
dokku docker-options:add node-js-app run "-v /var/log/node-js-app:/app/logs"
```

> Note: When [mounting a host directory](https://docs.docker.com/engine/reference/run/#volume-shared-filesystems) in a Dokku app you should first create that directory as user `dokku` and then mount the directory under `/app` in the container using `docker-options` as above. Otherwise the app will lack write permission in the directory.

### Remove a Docker option

```shell
dokku docker-options:remove node-js-app run "-v /var/log/node-js-app:/app/logs"
```

### Displaying docker-options reports about an app

> New as of 0.8.1

You can get a report about the app's docker-options status using the `docker-options:report` command:

```shell
dokku docker-options:report
```

```
=====> node-js-app docker options information
       Docker options build:
       Docker options deploy: -v /var/log/node-js-app:/app/logs
       Docker options run:  -v /var/log/node-js-app:/app/logs
=====> python-sample docker options information
       Docker options build:
       Docker options deploy:
       Docker options run:
=====> ruby-sample docker options information
       Docker options build:
       Docker options deploy:
       Docker options run:
```

You can run the command for a specific app also.

```shell
dokku docker-options:report node-js-app
```

```
=====> node-js-app docker options information
       Docker options build:
       Docker options deploy: -v /var/log/node-js-app:/app/logs
       Docker options run:  -v /var/log/node-js-app:/app/logs
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku docker-options:report node-js-app --docker-options-build
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

More information on Docker options can be found here: https://docs.docker.com/engine/reference/commandline/run/.
