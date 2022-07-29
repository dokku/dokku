# Log Management

```
logs <app> [-h|--help] [-t|--tail] [-n|--num num] [-q|--quiet] [-p|--ps process]  # Display recent log output
logs:failed [--all|<app>]                                                  # Shows the last failed deploy logs
logs:report [<app>] [<flag>]                                               # Displays a logs report for one or more apps
logs:set [--global|<app>] <key> <value>                                    # Set or clear a logs property for an app
logs:vector-logs [--num num] [--tail]                                      # Display vector log output
logs:vector-start                                                          # Start the vector logging container
logs:vector-stop                                                           # Stop the vector logging container
```

## Usage

### Application logs

You can easily get logs of an app using the `logs` command:

```shell
dokku logs node-js-app
```

Logs are pulled via integration with the scheduler for the specified application via "live tailing". As such, logs from previously running deployments are usually not available. Users that desire to see logs from previous deployments for debugging purposes should persist those logs to external services. Please see Dokku's [vector integration](deployment/logs.md#vector-logging-shipping) for more information on how to persist logs across deployments to ship logs to another service or a third-party platform.

#### Behavioral modifiers

Dokku also supports certain command-line arguments that augment the `log` command's behavior.

```
-n, --num NUM        # the number of lines to display
-p, --ps PS          # only display logs from the given process
-t, --tail           # continually stream logs
-q, --quiet          # display raw logs without colors, time and names
```

You can use these modifiers as follows:

```shell
dokku logs node-js-app -t -p web
```

The above command will show logs continually from the web process.

### Failed deploy logs

> Warning: The default [docker-local scheduler](/docs/advanced-usage/schedulers/docker-local.md) will "store" these until the next deploy or until the old containers are garbage collected - whichever runs first. If you require the logs beyond this point in time, please ship the logs to a centralized log server.

In some cases, it may be useful to retrieve the logs from a previously failed deploy.

You can retrieve these logs by using the `logs:failed` command.

```shell
dokku logs:failed node-js-app
```

You may also fetch all failed app logs by using the `--all` flag.

```shell
dokku logs:failed --all
```

### Docker Log Retention

Docker log retention can be specified via the `logs:set` command by specifying a value for `max-size`. Log retention is set via injected docker options for all applications, but is also available via the `logs-get-property` trigger for alternative schedulers.

```shell
dokku logs:set node-js-app max-size 20m
```

The default value may be set by passing an empty value for the option:

```shell
dokku logs:set node-js-app max-size
```

Valid values include any integer number followed by a unit of measure (`k`, `m`, or `g`) or the string `unlimited`. Setting to `unlimited` will result in Dokku omitting the log option.

The `max-size` property can also be set globally. The global default is `10m`, and the global value is used when no app-specific value is set.

```shell
dokku logs:set --global max-size 20m
```

The default value may be set by passing an empty value for the option.

```shell
dokku logs:set --global max-size
```

### Vector Logging Shipping

> New as of 0.22.6

Vector is an open-source, lightweight and ultra-fast tool for building observability pipelines. Dokku integrates with it for shipping container logs for the `docker-local` scheduler. Users may configure log-shipping on a per-app or global basis, neither of which interfere with the `dokku logs` commands.

#### Starting the Vector container

> Warning: While the default vector image may be updated over time, this will not impact running vector containers. Users are encouraged to view any Dokku and Vector changelogs to ensure their system will continue running as expected.

Vector may be started via the `logs:vector-start` command.

```shell
dokku logs:vector-start
```

This will start a new container named `vector` with Dokku's vector config mounted and ready for use. If a running container already exists, this command will do nothing. Additionally, if a container exists but is not running, this command will attempt to start the container.

While the default vector image is hardcoded, users may specify an alternative via the `--vector-image` flag:

```shell
dokku logs:vector-start --vector-image timberio/vector:latest-debian
```

The `vector` container will be started with the following volume mounts:

- `/var/lib/dokku/data/logs/vector.json:/etc/vector/vector.json`
- `/var/run/docker.sock:/var/run/docker.sock`
- `/var/log/dokku/apps:/var/log/dokku/apps`

The final volume mount - `/var/log/dokku/apps` - may be used for users that wish to ship logs to a file on disk that may be later logrotated. This directory is owned by the `dokku` user and group, with permissions set to `0755`. At this time, log-rotation is not configured for this directory.

#### Stopping the Vector container

Vector may be stopped via the `logs:vector-stop` command.

```shell
dokku logs:vector-stop
```

The `vector` container will be stopped and removed from the system. If the container is not running, this command will do nothing.

#### Checking Vector's Logs

It may be necessary to check the vector container's logs to ensure that vector is operating as expected. This can be performed with the `logs:vector-logs` command.

```shell
dokku logs:vector-logs
```

This command also supports the following modifiers:

```shell
--num NUM        # the number of lines to display
--tail           # continually stream logs
```

You can use these modifiers as follows:

```shell
dokku logs:vector-logs --tail --num 10
```

The above command will show logs continually from the vector container, with an initial history of 10 log lines

#### Configuring a log sink

Vector uses the concept of log "sinks" to send logs to a given endpoint. Log sinks may be configured globally or on a per-app basis by specifying a `vector-sink` in DSN form with the `logs:set` command. Specifying a sink value will reload any running vector container.

```shell
# setting the sink value in quotes is encouraged to avoid
# issues with ampersand encoding in shell commands
dokku logs:set node-js-app vector-sink "console://?encoding[codec]=json"
```

A sink may be removed by setting an empty value, which will also reload the running vector container.

```shell
dokku logs:set node-js-app vector-sink
```

Only one sink may be specified on a per-app basis at a given time.

Log sinks can also be specified globally by specifying the `--global` flag to `logs:set` with no app name specified:

```shell
dokku logs:set --global vector-sink "console://?encoding[codec]=json"
```

As with app-specific sink settings, the global value may also be cleared by setting no value.

```shell
dokku logs:set --global vector-sink
```

##### Log Sink DSN Format

The DSN form of a sink is as follows:

```
SINK_TYPE://?SINK_OPTIONS
```

Valid values for `SINK_TYPE` include all log vector log sinks, while `SINK_OPTIONS` is a query-string form for the sink's options. The following is a short description on how to set various values:

- `bool`: form: `key=bool`
- `string`: form: `key=string`
- `int`: form: `key=int`
- `[string]`: form: `key[]=string`
- `[int]`: form: `key[]=int`
- `table`: form: `option[key]=value`

For some sinks - such as the `http` sink - it may be useful to use special characters such as `&`. These characters must be url escaped as per [RFC 3986](https://datatracker.ietf.org/doc/html/rfc3986.html).

```shell
# the following command will set the `http` sink with a uri config value
# for a uri config value: https://loggerservice.com:1234/?token=abc1234&type=vector
# the url quoted version: https%3A//loggerservice.com%3A1234/%3Ftoken%3Dabc1234%26type%3Dvector
dokku logs:set test vector-sink "http://?uri=https%3A//loggerservice.com%3A1234/%3Ftoken%3Dabc1234%26type%3Dvector"
```

Please read the [sink documentation](https://vector.dev/docs/reference/sinks/) for your sink of choice to configure the sink as desired.
