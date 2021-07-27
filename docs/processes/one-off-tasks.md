# One-off Tasks

```
run [ --env KEY=VALUE | -e KEY=VALUE ] <app> <cmd>  # Run a command in a new container using the current application image
```

Sometimes it is necessary to run a one-off command under an application. Dokku makes it easy to run a fresh container via the `run` command.

## Usage

The `run` command can be used to run a one-off process for a specific command. This will start a new container and run the desired command within that container. Note that this container image will be stay around even after command completes. The container image will be the same container image as was used to start the currently deployed application.

```shell
# runs `ls -lah` in the `/app` directory of the application `node-js-app`
dokku run node-js-app ls -lah

# optionally, run can be passed custom environment variables
dokku run --env "NODE_ENV=development" --env "PATH=/custom/path" node-js-app npm run mytask
```

The `run` command can also be used to run a command defined in the app `Procfile`:

```
console: bundle exec racksh
```

```shell
# runs `bundle exec racksh` in the `/app` directory of the application `my-app`
dokku run my-app console
```

If the container running the command should be removed after exit, the `--rm-container` or `--rm` global flags can be specified to remove the containers automatically:

```shell
dokku --rm-container run node-js-app ls -lah
dokku --rm run node-js-app ls -lah
```

Alternatively, a global property can be set to always remove `run` containers.

```shell
# don't keep `run` containers around
dokku config:set --global DOKKU_RM_CONTAINER=1

# revert the above setting and keep containers around
dokku config:unset --global DOKKU_RM_CONTAINER
```

Containers may have specific labels attached. In order to avoid issues with dokku internals, do not use any labels beginning with either `com.dokku` or `org.label-schema`.

```shell
dokku --label=com.example.test-label=value run node-js-app ls -lah
```

Finally, a container can be run in "detached" mode via the `--detach` Dokku flag. Running a process in detached mode will immediately return a `CONTAINER_ID`. It is up to the user to then further manage this container in whatever manner they see fit, as Dokku will *not* automatically terminate the container.

```shell
dokku --detach run node-js-app ls -lah
# returns the ID of the new container
```

> Note that the `--rm-container` or `--rm` flags cannot be used when running containers in detached mode, and attempting to do so will result in the `--detach` flag being ignored.
