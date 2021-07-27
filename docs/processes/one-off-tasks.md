# One-off Tasks

```
run [ --env KEY=VALUE | -e KEY=VALUE ] <app> <cmd>  # Run a command in a new container using the current application image
```

Sometimes it is necessary to run a one-off command under an application. Dokku makes it easy to run a fresh container via the `run` command.

## Usage

### Running a one-off command

The `run` command can be used to run a one-off process for a specific command. This will start a new container and run the desired command within that container. This contianer will be removed after the process exits. The container image will be the same container image as was used to start the currently deployed application.

```shell
# runs `ls -lah` in the `/app` directory of the application `node-js-app`
dokku run node-js-app ls -lah

# optionally, run can be passed custom environment variables
dokku run --env "NODE_ENV=development" --env "PATH=/custom/path" node-js-app npm run mytask
```

#### Running Procfile commands

The `run` command can also be used to run a command defined in the app `Procfile`:

```
console: bundle exec racksh
```

```shell
# runs `bundle exec racksh` in the `/app` directory of the application `my-app`
dokku run my-app console
```

#### Specifying container labels

Containers may have specific labels attached. In order to avoid issues with dokku internals, do not use any labels beginning with either `com.dokku` or `org.label-schema`.

```shell
dokku --label=com.example.test-label=value run node-js-app ls -lah
```

### Running a detached container

Finally, a container can be run in "detached" mode via the `run:detached` Dokku command. Running a process in detached mode will immediately return a `CONTAINER_ID`. It is up to the user to then further manage this container in whatever manner they see fit, as Dokku will *not* automatically terminate the container.

```shell
dokku run:detached node-js-app ls -lah
# returns the ID of the new container
```
