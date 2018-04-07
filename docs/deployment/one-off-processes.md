# One-off Processes and Cron

Sometimes you need to either inspect running containers or run a one-off command under an application. In those cases, Dokku makes it easy to either connect to a running container or run a fresh container.

## Run a command in an app environment

```
run <app> <cmd>                                # Run a command in a new container using the current application image
```

The `run` command can be used to run a one-off process for a specific command. This will start a new container and run the desired command within that container. Note that this container will be stay around even after command completes. The container will be the same container as was used to start the currently deployed application.

```shell
# runs `ls -lah` in the `/app` directory of the application `node-js-app`
dokku run node-js-app ls -lah
```

The `run` command can also be used to run a command defined in your Procfile:

```
console: bundle exec racksh
```

```shell
# runs `bundle exec racksh` in the `/app` directory of the application `my-app`
dokku run my-app console
```

If you want to remove the container after a command has started, you can run the following command:

```shell
# don't keep `run` containers around
dokku config:set --global DOKKU_RM_CONTAINER=1

# revert the above setting and keep containers around
dokku config:unset --global DOKKU_RM_CONTAINER
```

You may also use the `--rm-container` or `--rm` Dokku flags to remove the containers automatically:

```shell
dokku --rm-container run node-js-app ls -lah
dokku --rm run node-js-app ls -lah
```

Finally, you may wish to run a container in "detached" mode via the `--detach` Dokku flag. Running a process in detached mode will immediately return a `CONTAINER_ID`. It is up to the user to then further manage this container in whatever manner they see fit, as Dokku will *not* automatically terminate the container.

```shell
dokku --detach run node-js-app ls -lah
# returns the ID of the new container
```

> Note that you may not use the `--rm-container` or `--rm` flags when running containers in detached mode, and attempting to do so will result in the `--detach` flag being ignored.

### Using `run` for cron tasks

You can always use a one-off container to run an application task:

```shell
dokku --rm run node-js-app some-command
dokku --rm-container run node-js-app some-command
```

For tasks that should not be interrupted, run is the **preferred** method of handling cron tasks, as the container will continue running even during a deploy or scaling event. The trade-off is that there will be an increase in memory usage if there are multiple concurrent tasks running.

## Entering existing containers

> New as of 0.4.0

```
enter <app>  [<container-type> || --container-id <container-id>]  # Connect to a specific app container
```

The `enter` command can be used to enter a running container. The following variations of the command exist:

```shell
dokku enter node-js-app web
dokku enter node-js-app web.1
dokku enter node-js-app --container-id ID
```

Additionally, you can run `enter` with no container-type. If only a single container-type is defined in your app, you will be dropped into the only running container. This behavior is not supported when specifying a custom command; as described below.

By default, it runs a `/bin/bash`, but can also be used to run a custom command:

```shell
# just echo hi
dokku enter node-js-app web echo hi

# run a long-running command, as one might for a cron task
dokku enter node-js-app web python script/background-worker.py
```

### Using `enter` for cron tasks

Your procfile can have the following entry:

```Procfile
cron: sleep infinity
```

With the `cron` process scaled to `1`:

```shell
dokku ps:scale node-js-app cron=1
```

You can now run all your commands in that container:

```shell
dokku enter node-js-app cron some-command
```

Note that you can also run multiple commands at the same time to reduce memory usage, though that may result in polluting the container environment.

For tasks that will properly resume, you **should** use the above method, as running tasks will be interrupted during deploys and scaling events, and subsequent commands will always run with the latest container. Note that if you scale the cron container down, this may interrupt proper running of the task.

## General Cron Recommendations

Regularly scheduled tasks can be a bit of a pain with dokku. The following are general recommendations to follow to help ensure successful task runs.

- Use the `dokku` user's crontab
  - If you do not, the `dokku` binary will attempt to execute with `sudo` dokku, and your cron run with fail with `sudo: no tty present and no askpass program specified`
- Add a `MAILTO` environment variable to ship cron emails to yourself.
- Add a `PATH` environment variable or specify the full path to binaries on the host.
- Add a `SHELL` environment variable to specify bash when running commands.
- Keep your cron tasks in time-sorted order.
- Keep your server time in UTC so you don't need to translate daylight saving's time when reading the cronfile.
- Run tasks at the lowest traffic times if possible.
- Use cron to **trigger** jobs, not run them. Use a real queuing system such as rabbitmq to actually process jobs.
- Try to keep tasks quiet so that mails only send on errors.
- Do not silence standard error or standard out. If you silence the former, you will miss failures. Silencing the latter means you should actually make application changes to handle log levels.
- Use a service such as [Dead Man's Snitch](https://deadmanssnitch.com) to verify that cron tasks completed successfully.
- Add lots of comments to your cronfile, including what a task is doing, so that you don't spend time deciphering the file later.
- Place your cronfiles in a pattern such as `/etc/cron.d/APP`.
- Do not use non-ascii characters in your cronfile names. Cron is finicky.
- Remember to have trailing newlines in your cronfile! Cron is finicky.

The following is a sample cronfile that you can use for your applications:

```cron
# server cron jobs
MAILTO="mail@dokku.me"
PATH=/usr/local/bin:/usr/bin:/bin
SHELL=/bin/bash

# m   h   dom mon dow   username command
# *   *   *   *   *     dokku    command to be executed
# -   -   -   -   -
# |   |   |   |   |
# |   |   |   |   +----- day of week (0 - 6) (Sunday=0)
# |   |   |   +------- month (1 - 12)
# |   |   +--------- day of month (1 - 31)
# |   +----------- hour (0 - 23)
# +----------- min (0 - 59)

### HIGH TRAFFIC TIME IS B/W 00:00 - 04:00 AND 14:00 - 23:59
### RUN YOUR TASKS FROM 04:00 - 14:00
### KEEP SORTED IN TIME ORDER

### PLACE ALL CRON TASKS BELOW

# removes unresponsive users from the subscriber list to decrease bounce rates
0 0 * * * dokku dokku --rm run node-js-app some-command

# sends out our email alerts to users
0 1 * * * dokku dokku ps:scale node-js-app cron=1 && dokku enter node-js-app cron some-other-command && dokku ps:scale node-js-app cron=0

### PLACE ALL CRON TASKS ABOVE, DO NOT REMOVE THE WHITESPACE AFTER THIS LINE

```
