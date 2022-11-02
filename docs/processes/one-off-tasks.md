# One-off Tasks

```
run [-e|--env KEY=VALUE] [--no-tty] <app> <cmd>         # Run a command in a new container using the current app image
run:detached [-e|-env KEY=VALUE] [--no-tty] <app> <cmd> # Run a command in a new detached container using the current app image
run:list [--format json|stdout] [<app>]                 # List all run containers for an app
run:stop <app|--container CONTAINER>                    # Stops all run containers for an app or a specified run container
```

Sometimes it is necessary to run a one-off command under an app. Dokku makes it easy to run a fresh container via the `run` command.

## Usage

### Running a one-off command

The `run` command can be used to run a one-off process for a specific command. This will start a new container and run the desired command within that container.  The container image will be the same container image as was used to start the currently deployed app.

> New as of 0.25.0, this container will be removed after the process exits.

```shell
# runs `ls -lah` in the `/app` directory of the app `node-js-app`
dokku run node-js-app ls -lah

# optionally, run can be passed custom environment variables
dokku run --env "NODE_ENV=development" --env "PATH=/custom/path" node-js-app npm run mytask
```

One off containers are removed at the end of process execution.

#### Running Procfile commands

The `run` command can also be used to run a command defined in the app `Procfile`:

```
console: bundle exec racksh
```

```shell
# runs `bundle exec racksh` in the `/app` directory of the app `my-app`
dokku run my-app console
```

#### Specifying container labels

Containers may have specific labels attached. In order to avoid issues with dokku internals, do not use any labels beginning with either `com.dokku` or `org.label-schema`.

```shell
dokku --label=com.example.test-label=value run node-js-app ls -lah
```

#### Disabling TTY

> New as of 0.25.0

One-off containers default to interactive mode where possible. To disable this behavior, specify the `--no-tty` flag:

```shell
dokku run --no-tty node-js-app ls -lah
```

### Running a detached container

> New as of 0.25.0

Finally, a container can be run in "detached" mode via the `run:detached` Dokku command. Running a process in detached mode will immediately return a `CONTAINER_ID`. Detached containers are run without a tty and are also removed at the end of process execution.

```shell
dokku run:detached node-js-app ls -lah
# returns the ID of the new container
```

### Listing one-off containers

> New as of 0.25.0

One-off containers for a given app can be listed via the `run:list` command:

```shell
dokku run:list node-js-app
```

```
=====> node-js-app run containers
NAMES                   COMMAND            CREATED
node-js-app.run.28689   "/exec sleep 15"   2 seconds ago
```

> The `COMMAND` displayed will be what Docker executes and may not exactly match the command specified by a `dokku run` command.

The output can also be shown in json format:

```shell
dokku run:list node-js-app --format json
```

```
[
  {
    "name": "node-js-app.run.28689",
    "state": "running",
    "command": "\"/exec 'sleep 15'\"",
    "created_at": "2022-08-03 05:47:44 +0000 UTC"
  }
]
```

### Stopping a one-off cotainer

> New as of 0.29.0

Run containers for an app can be stopped via the `run:stop` command. The output will be the container id.

```shell
# start a container
# the output will be something like: node-js-app.run.2313
dokku run node-js-app sleep 300

# stop the container
dokku run:stop --container node-js-app.run.2313
````

```
node-js-app.run.2313
```

All containers for a given app can be stopped by specifying the app name.

```shell
dokku run:stop node-js-app
```

```
node-js-app.run.2313
node-js-app.run.574
```
