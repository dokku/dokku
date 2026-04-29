# Docker Container Options

> [!IMPORTANT]
> New as of 0.3.17

```
docker-options:add [--process PROC...] <app> <phase(s)> OPTION    # Add Docker option to app for phase
docker-options:clear [--process PROC...] <app> [<phase(s)>...]    # Clear docker options from app
docker-options:list <app> [--process PROC] --phase PHASE          # List docker options for one process+phase pair
docker-options:remove [--process PROC...] <app> <phase(s)> OPTION # Remove Docker option from app for phase
docker-options:report [<app>] [<flag>] [--format json|stdout]     # Displays a docker options report for one or more apps
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

Multiple docker options can also be specified in a single call. Each `--flag [value]` group is detected on flag boundaries, shell-tokenized for quoting safety, and stored as its own entry so it round-trips through `docker-options:report` and `docker-options:list`:

```shell
dokku docker-options:add node-js-app deploy "--ulimit nofile=12" "--shm-size 256m"
```

A misplaced `--process PROC` (i.e. one specified after the app name instead of before it) is honored as a subcommand flag rather than stored as a docker option, so the example above and the equivalent process-scoped form below behave identically:

```shell
dokku docker-options:add --process web node-js-app deploy "--ulimit nofile=12" "--shm-size 256m"
dokku docker-options:add node-js-app deploy "--ulimit nofile=12" "--shm-size 256m" --process web
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

Multiple docker options can also be removed in a single call, mirroring the splitting that `docker-options:add` performs:

```shell
dokku docker-options:remove node-js-app deploy "--ulimit nofile=12" "--shm-size 256m"
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

When process-specific options are configured (see below), the report exposes one additional dynamic flag per configured `process.deploy` pair, named `--docker-options-deploy.<process>`:

```shell
dokku docker-options:report node-js-app --docker-options-deploy.web
```

A machine-readable JSON view is available via `--format json`:

```shell
dokku docker-options:report node-js-app --format json
```

### Process-Specific Options

> [!IMPORTANT]
> New as of 0.38.0

Docker options can be scoped to specific process types declared in the app's `Procfile` by passing one or more `--process` flags. This is useful when a deploy-phase option (for example a port mapping) makes sense for one process type but would conflict with another - the canonical case being a `web` process that needs `-p 6789:5000` published while the `worker` process must not bind that port.

Process scoping is supported only for the `deploy` phase. The `build` phase runs once per app and the `run` phase covers ad-hoc commands and cron tasks where no Procfile process type is in play; both reject `--process`.

There is no `--global` flag. Omitting `--process` keeps the historical behavior of applying the option to every container in the app. Avoiding a `--global` flag here is intentional: elsewhere in Dokku `--global` means "across all apps" (e.g. `dokku config:set --global`), which would be misleading in this plugin where the scope is always one app.

#### Setting process-specific options

```shell
# Add a port mapping only to the web process
dokku docker-options:add --process web node-js-app deploy "-p 6789:5000"

# Add a GPU mount only to the worker process
dokku docker-options:add --process worker node-js-app deploy "--gpus all"
```

Multiple `--process` flags can be combined to apply the same option to several process types in one call:

```shell
dokku docker-options:add --process web --process api node-js-app deploy "-v /shared:/shared"
```

If `--process` names a process type that is not currently declared in the app's `Procfile`, the command succeeds but emits a warning. This allows configuring options ahead of a deploy that adds the new process type.

The `_default_` value is reserved internally and cannot be passed to `--process`.

#### Removing and clearing process-specific options

```shell
# Remove a single option from one process
dokku docker-options:remove --process web node-js-app deploy "-p 6789:5000"

# Clear every option for a process+phase
dokku docker-options:clear --process worker node-js-app deploy
```

Without `--process`, `:remove` and `:clear` operate on the default scope only - per-process lists are left untouched.

#### Listing options for a process and phase

The `docker-options:list` command prints the options stored for a single process+phase pair, one option per line. Omitting `--process` lists the default scope.

```shell
dokku docker-options:list node-js-app --process web --phase deploy
dokku docker-options:list node-js-app --phase deploy
```
