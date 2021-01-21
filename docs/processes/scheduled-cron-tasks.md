# Scheduled Cron T

> New as of 0.23.0

```
cron:list <app>               # List scheduled cron tasks for an app
cron:report [<app>] [<flag>]  # Display report about an app
```

## Usage

### Dokku Managed Cron

Dokku automates scheduled `dokku run` commands via it's `app.json` cron integration.

#### Specifying commands

The `app.json` file for a given app can define a special `cron` key that contains a list of commands to run on given schedules. The following is a simple example `app.json` that simply runs the command `true` once a day:

```
{
  "cron": [
    {
      "command": "true",
      "schedule": "@daily"
    }
  ]
}
```

A cron entry takes the following properties:

- `command`: A command to be run within the built app image. Specified commands can also be `Procfile` entries.
- `schedule`: A [cron-compatible](https://en.wikipedia.org/wiki/Cron#Overview) scheduling definition upon which to run the command. Seconds are generally not supported.

Zero or more cron commands can be specified per app. Cron entries are validated after the build artifact is created but before the app is deployed, and the cron schedule is updated during the post-deploy phase.

#### Task Environment

When running scheduled cron tasks, there are a few items to be aware of:

- Scheduled cron tasks are performed within the app environment available at runtime. If the app image does not exist, the command may fail to execute.
- Schedules are performed on the hosting server's timezone, which is typically UTC.
- At this time, only the `PATH` and `SHELL` environment variables are specified in the cron template.
- Each scheduled task is executed within a one-off `run` container, and thus inherit any docker-options specified for `run` containers.Resources are never shared between scheduled tasks.
- Tasks are always executed using the `--rm` flag, and are never executed in detached mode.
- Scheduled cron tasks are supported on a per-scheduler basis, and are currently only implemented by the `docker-local` scheduler.
- Tasks for _all_ apps managed by the `docker-local` scheduler are written to a single crontab file owned by the `dokku` user. The `dokku` user's crontab should be considered reserved for this purpose.

#### Listing Cron tasks

Cron tasks for an app can be listed via the `cron:list` command. This command takes an `app` argument.

```shell
dokku cron:list node-js-app
```

```
ID                                    Schedule   Command
cGhwPT09cGhwIHRlc3QucGhwPT09QGRhaWx5  @daily     node index.js
cGhwPT09dHJ1ZT09PSogKiAqICogKg==      * * * * *  true
```

#### Displaying reports for an app

You can get a report about the cron configuration for apps using the `cron:report` command:

```shell
dokku cron:report node-js-app
```

```
=====> node-js-app cron information
       Cron task count:               2
=====> python-sample cron information
       Cron task count:               0
=====> ruby-sample cron information
       Cron task count:               10
```

You can run the command for a specific app also.

```shell
dokku cron:report node-js-app
```

```
=====> node-js-app cron information
       Cron task count:               2
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku cron:report node-js-app --cron-task-count
```

### Self Managed Cron

> Warning: Self-managed cron tasks should be considered advanced usage. While the instructions are available, users are highly encouraged to use the built-in scheduled cron task support unless absolutely necessary.

Some installations may require more fine-grained control over cron usage. The following are advanced instructions for configuring cron.

#### Using `run` for cron tasks

You can always use a one-off container to run an app task:

```shell
dokku --rm run node-js-app some-command
dokku --rm-container run node-js-app some-command
```

For tasks that should not be interrupted, run is the *preferred* method of handling cron tasks, as the container will continue running even during a deploy or scaling event. The trade-off is that there will be an increase in memory usage if there are multiple concurrent tasks running.


#### Using `enter` for cron tasks

Your Procfile can have the following entry:

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

For tasks that will properly resume, you *should* use the above method, as running tasks will be interrupted during deploys and scaling events, and subsequent commands will always run with the latest container. Note that if you scale the cron container down, this may interrupt proper running of the task.

#### General cron recommendations

Regularly scheduled tasks can be a bit of a pain with Dokku. The following are general recommendations to follow to help ensure successful task runs.

- Use the `dokku` user's crontab.
  - If you do not, the `dokku` binary will attempt to execute with `sudo`, and your cron run with fail with `sudo: no tty present and no askpass program specified`.
- Add a `MAILTO` environment variable to ship cron emails to yourself.
- Add a `PATH` environment variable or specify the full path to binaries on the host.
- Add a `SHELL` environment variable to specify Bash when running commands.
- Keep your cron tasks in time-sorted order.
- Keep your server time in UTC so you don't need to translate daylight savings time when reading the cronfile.
- Run tasks at the lowest traffic times if possible.
- Use cron to *trigger* jobs, not run them. Use a real queuing system such as rabbitmq to actually process jobs.
- Try to keep tasks quiet so that mails only send on errors.
- Do not silence standard error or standard out. If you silence the former, you will miss failures. Silencing the latter means you should actually make app changes to handle log levels.
- Use a service such as [Dead Man's Snitch](https://deadmanssnitch.com) to verify that cron tasks completed successfully.
- Add lots of comments to your cronfile, including what a task is doing, so that you don't spend time deciphering the file later.
- Place your cronfiles in a pattern such as `/etc/cron.d/APP`.
- Do not use non-ASCII characters in your cronfile names. cron is finicky.
- Remember to have trailing newlines in your cronfile! cron is finicky.

The following is a sample cronfile that you can use for your apps:

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
