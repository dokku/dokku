# Docker Container Options

> [!IMPORTANT]
> New as of 0.3.17

```
docker-options:add <app> <phase(s)> OPTION    # Add Docker option to app for phase (comma-separated phase list)
docker-options:clear <app> [<phase(s)>...]    # Clear a docker options from app
docker-options:remove <app> <phase(s)> OPTION # Remove Docker option from app for phase (comma-separated phase list)
docker-options:report [<app>] [<flag>]        # Displays a docker options report for one or more apps
```

The `docker-options` plugin allows users to specify custom [container options](https://docs.docker.com/engine/reference/run/) for containers created by Dokku at various phases.

## Usage

### Background

#### Phases

Dokku deploys your app in multiple "phases" and the `docker-options` plugin allows you to pass arguments to their underlying docker container:

- `build`: The `build` phase is used to provide container options that are available during the build process for use by the various builders.
    - A given builder may strip out or ignore options that are unsupported by the builder in question - as an example, the `dockerfile` builder does not support mounted volumes.
- `deploy`: The `deploy` phase is used to provide container options that are set on _deployed_ process types. This covers every process type specified in an app `Procfile` as well as any default processes your app may deploy.
    - The `deploy` phase is usually the correct phase to add options for running containers.
- `run`: The `run` phase is used to provide container options to one-off containers created by `dokku run`, `dokku run:detached`, and any cron tasks specified in your `app.json`.

> [!IMPORTANT]
> The `run` phase does _not_ correspond 1-to-1 to `docker run` or `docker container run` commands. Specifying a container option at the `run` phase will only be invoked on containers created by the `run` plugin and cron tasks. Please be sure to specify options at the correct phase for your use-case.

Adding or removing docker-options will not apply to any running containers, and only applies to containers created _after_ the options have been modified. As such, changing an app's docker options must be followed by a `dokku ps:rebuild` or a deploy in order to take effect.

#### Supported Docker Options

More information on supported Docker options can be found [here](https://docs.docker.com/engine/reference/commandline/run/).

Container options configured via the `docker-options` plugin are not used to modify the process a container runs. Container options are the `[OPTIONS]` portion of the following, where `[CONTAINER_COMMAND]` and `[ARG]` are the process and the arguments passed to it that are launched in the created container: `docker run [OPTIONS] [CONTAINER_COMMAND] [ARG...]`. Please see the documentation for [customizing the run command](/docs/deployment/builders/dockerfiles.md#customizing-the-run-command) or use a [Procfile](/docs/deployment/builders/dockerfiles.md#procfiles-and-multiple-processes) to modify the command used by a Dockerfile-based container.

#### Mounting volumes and host directories

Docker supports volume and host directory mounting via the `-v` or `--volume` flags. In order to simplify usage, Dokku provides a `storage` plugin as an abstraction to interact with persistent storage. In most cases, the Dokku project recommends using the persistent storage plugin over directly manipulating docker options at different phases. See the [persistent storage documentation](/docs/advanced-usage/persistent-storage.md) for more information on how to attach persistent storage to your app.

### Commands

#### Add Docker options

To add a docker option to an app, use the `docker-options:add` command. This takes an app name, a comma-separated list of phases, and the docker-option to add.

```shell
dokku docker-options:add node-js-app deploy "--ulimit nofile=12"
```

Multiple phases can be specified by using a comma when specifying phases:

```shell
dokku docker-options:add node-js-app deploy,run "--ulimit nofile=12"
```

The `docker-options:add` does not support setting multiple options in a single call. To specify multiple options, call `docker-options:add` multiple times.

```shell
dokku docker-options:add node-js-app deploy "--ulimit nofile=12"
dokku docker-options:add node-js-app deploy "--shm-size 256m"
```

#### Remove a Docker option

To remove docker options from an app, use the `docker-options:remove` command. This takes an app name, a comma-separated list of phases, and the docker-option to remove.

```shell
dokku docker-options:remove node-js-app run "--ulimit nofile=12"
```

Multiple phases can be specified by using a comma when specifying phases:

```shell
dokku docker-options:remove node-js-app deploy,run "--ulimit nofile=12"
```

The `docker-options:remove` does not support setting multiple options in a single call. To specify multiple options, call `docker-options:remove` multiple times.

```shell
dokku docker-options:remove node-js-app deploy "--ulimit nofile=12"
dokku docker-options:remove node-js-app deploy "--shm-size 256m"
```

#### Clear all Docker options for an app

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

#### Displaying docker-options reports for an app

> [!IMPORTANT]
> New as of 0.8.1

You can get a report about the app's docker-options status using the `docker-options:report` command:

```shell
dokku docker-options:report
```

```
=====> node-js-app docker options information
       Docker options build:
       Docker options deploy: --ulimit nofile=12 --shm-size 256m
       Docker options run:  --ulimit nofile=12 --shm-size 256m
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
