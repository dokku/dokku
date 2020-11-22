# Log Management

```
logs <app> [-h] [-t] [-n num] [-q] [-p process]  # Display recent log output
logs:failed  [--parallel count] [--all|<app>]    # Shows the last failed deploy logs
```

## Usage

### Application logs

You can easily get logs of an app using the `logs` command:

```shell
dokku logs node-js-app
```

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

will show logs continually from the web process.

### Failed deploy logs

> Warning: The default [docker-local scheduler](/docs/advanced-usage/schedulers/docker-local.md) will "store" these until the next deploy or until the old containers are garbage collected - whichever runs first. If you require the logs beyond this point in time, please ship the logs to a centralized log server.

In some cases, it may be useful to retrieve the logs from a previously failed deploy.

You can retrieve these logs by using the `logs:failed` command

```shell
dokku logs:failed node-js-app
```

You may also fetch all failed app logs by using the `--all` flag.

```shell
dokku logs:failed --all
```
