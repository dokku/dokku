# Docker Container Options

> New as of 0.3.17

Pass [container options](https://docs.docker.com/engine/reference/run/) to the `docker run` command  during Dokku's `build`, `deploy` and `run` phases

```
docker-options:add <app> <phase(s)> OPTION    # Add Docker option to app for phase (comma-separated phase list)
docker-options:clear <app> [<phase(s)>...]    # Clear a docker options from application
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

Manipulation of docker options will not restart running containers. This enables multiple options to be set/unset before final application. As such, changing an app's docker options must be followed by a `dokku ps:rebuild` in order to take effect.

More information on supported Docker options can be found here: https://docs.docker.com/engine/reference/commandline/run/.

Container options configured via the `docker-options` plugin are not used to modify the process a container runs. Container options are the `[OPTIONS]` portion of the following, where `[CONTAINER_COMMAND]` and `[ARG]` are the process and the arguments passed to it that are launched in the created container: `docker run [OPTIONS] [CONTAINER_COMMAND] [ARG...]`. Please see the documentation for [customizing the run command](/docs/deployment/builders/dockerfiles.md#customizing-the-run-command) or use a [Procfile](/docs/deployment/builders/dockerfiles.md#procfiles-and-multiple-processes) to modify the command used by a Dockerfile-based container.

## Examples

### Add Docker options

Add some options for the deployed/running app and when executing [`dokku run`](/docs/processes/one-off-tasks.md):

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

### Clear all Docker options for an app

Docker options can be removed for a specific app using the `docker-options:clear` command.

```shell
dokku docker-options:clear node-js-app
```

```
-----> Clearing docker-options for node-js-app on all phases
```

One or more valid phases can also be specified. Phases are comma delimited, and specifying an invalid phase will result in an error.

```shell
dokku docker-options:clear node-js-app run
```

```
-----> Clearing docker-options for node-js-app on phase run
```

```shell
dokku docker-options:clear node-js-app build,run
```

```
-----> Clearing docker-options for node-js-app on phase build
-----> Clearing docker-options for node-js-app on phase run
```

### Displaying docker-options reports for an app

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
