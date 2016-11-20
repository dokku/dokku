# Log Management

```
logs <app>           # display recent log output
```

## Usage

### Application Logs
You can easily get logs of an application using the `logs` command:

```shell
dokku logs node-js-app
```

## Behavioral modifiers

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
