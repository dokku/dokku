# Zero Downtime Deploy Checks

> [!IMPORTANT]
> New as of 0.5.0

```
checks:disable <app> [process-type(s)]    Disable zero-downtime deployment for all processes (or comma-separated process-type list) ***WARNING: this will cause downtime during deployments***
checks:enable <app> [process-type(s)]     Enable zero-downtime deployment for all processes (or comma-separated process-type list)
checks:report [<app>] [<flag>]            Displays a checks report for one or more apps
checks:run <app> [process-type(s)]        Runs zero-downtime checks for all processes (or comma-separated process-type list)
checks:set [--global|<app>] <key> <value> Set or clear a logs property for an app
checks:skip <app> [process-type(s)]       Skip zero-downtime checks for all processes (or comma-separated process-type list)
```

By default, Dokku will wait `10` seconds after starting each container before assuming it is up and proceeding with the deploy. Once this has occurred for all containers started by an application, traffic will be switched to point to your new containers. Dokku will also wait a further `60` seconds *after* the deploy is complete before terminating old containers in order to give time for long running connections to terminate. In either case, you may have more than one container running for a given application.

You may both create user-defined checks for web processes using the `healthchecks` key in the `app.json` file, as well as customize any and all parts of this experience using the checks plugin.

> Web checks are performed via `curl` on Dokku host. Some application code - such
> as the Django framework - checks for specific hostnames or header values, these
> checks will fail. To avoid this:
>
> - Remove such checks from your code: Modify your application to remove the hostname check completely.
> - Allow checks from all hostnames: Modify your application to accept a dynamically provided hostname.
> - Specify the domain within the check: See below for further documentation.

## Configuring checks settings

### wait-to-retire

After a successful deploy, the grace period given to old containers before they are stopped/terminated is determined by the value of `wait-to-retire`. This is useful for ensuring completion of long-running HTTP connections.

```shell
dokku checks:set node-js-app wait-to-retire 30
```

Defaults to `60`.

## Configuring check settings using the `config` plugin

There are certain settings that can be configured via environment variables:

- `DOKKU_DEFAULT_CHECKS_WAIT`: (default: `10`) If no user-defined checks are specified - or if the process being checked is not a `web` process - this is the period of time Dokku will wait before checking that a container is still running.
- `DOKKU_DOCKER_STOP_TIMEOUT`: (default: `10`) Configurable grace period given to the `docker stop` command. If a container has not stopped by this time, a `kill -9` signal or equivalent is sent in order to force-terminate the container. Both the `ps:stop` and `apps:destroy` commands *also* respect this value. If not specified, the Docker defaults for the [`docker stop` command](https://docs.docker.com/engine/reference/commandline/stop/) will be used.

The following settings may also be specified in the `app.json` file, though are available as environment variables in order to ease application reuse.

- `DOKKU_CHECKS_WAIT`: (default: `5`) Wait this many seconds for the container to start before running checks.
- `DOKKU_CHECKS_TIMEOUT`: (default: `30`) Wait this many seconds for each response before marking it as a failure.
- `DOKKU_CHECKS_ATTEMPTS`: (default: `5`) Number of retries for to run for a specific check before marking it as a failure

## Skipping and Disabling Checks

> Note that `checks:disable` will now (as of 0.6.0) cause downtime for that process-type during deployments. Previously, it acted as `checks:skip` currently does.

You can choose to skip checks completely on a per-application/per-process basis. Skipping checks will avoid the default 10 second waiting period entirely, as well as any other user-defined checks.

```shell
# process type specification is optional
dokku checks:skip node-js-app worker,web
```

```
-----> Skipping zero downtime for app's (node-js-app) proctypes (worker,web)
-----> Unsetting node-js-app
-----> Unsetting DOKKU_CHECKS_DISABLED
-----> Setting config vars
       DOKKU_CHECKS_SKIPPED: worker,web
```

Zero downtime checks can also be disabled completely. This will stop old containers *before* new ones start, which may result in broken connections and downtime if your application fails to boot properly.

```shell
dokku checks:disable node-js-app worker
```

```
-----> Disabling zero downtime for app's (node-js-app) proctypes (worker)
-----> Setting config vars
       DOKKU_CHECKS_DISABLED: worker
-----> Setting config vars
       DOKKU_CHECKS_SKIPPED: web
```

### Displaying checks reports for an app

> [!IMPORTANT]
> New as of 0.8.1

You can get a report about the app's checks status using the `checks:report` command:

```shell
dokku checks:report
```

```
=====> node-js-app checks information
       Checks disabled list: none
       Checks skipped list: none
       Checks computed wait to retire: 60
       Checks global wait to retire: 60
       Checks wait to retire:
=====> python-app checks information
       Checks disabled list: none
       Checks skipped list: none
       Checks computed wait to retire: 60
       Checks global wait to retire: 60
       Checks wait to retire:
=====> ruby-app checks information
       Checks disabled list: _all_
       Checks skipped list: none
       Checks computed wait to retire: 60
       Checks global wait to retire: 60
       Checks wait to retire:
```

You can run the command for a specific app also.

```shell
dokku checks:report node-js-app
```

```
=====> node-js-app checks information
       Checks disabled list: none
       Checks skipped list: none   
       Checks computed wait to retire: 60
       Checks global wait to retire: 60
       Checks wait to retire:       
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku checks:report node-js-app --checks-disabled-list
```

## Customizing checks

> [!IMPORTANT]
> New as of 0.31.0

If your application needs a longer period to boot up - perhaps to load data into memory, or because of slow boot time - you may also use Dokku's `checks` functionality to more precisely check whether an application can serve traffic or not.

Healthchecks are run against all process from your application's `Procfile`. When no healthcheck is defined, Dokku will fallback to a process uptime check.

One or more healthchecks can be defined in the `app.json` file - see the [deployment task documentation](/docs/advanced-usage/deployment-tasks.md) for more information on how this is extracted - under the `healthchecks.web` path:

```json
{
  "healthchecks": {
    "web": [
        {
            "type":        "startup",
            "name":        "web check",
            "description": "Checking if the app responds to the /health/ready endpoint",
            "path":        "/health/ready",
            "attempts": 3
        }
    ]
  }
}
```

A healthcheck entry takes the following properties:

- `attempts`: (default: `3`) Number of retry attempts to perform on failure.
- `command`: (default `''` - empty string) Command to execute within container.
- `content`: (default: `''` - empty string) Content to search in http path check output.
- `initialDelay`: (default: 0, unit: seconds) Number of seconds to wait after a container has started before triggering the healthcheck.
- `name`: (default: autogenerated) The name of the healthcheck. If unspecified, it will be autogenerated from the rest of the healthcheck information.
- `path`: (default: `/` - for http checks): An http path to check.
- `port`: (default: `5000`): Port to run healthcheck against.
- `timeout`: (default: `5` seconds): Number of seconds to wait before timing out a healthcheck.
- `type`: (default: `""` - none): Type of the healthcheck. Options: liveness, readiness, startup.
- `uptime`: (default: `""` - none): Amount of time the container must be alive before the container is considered healthy. Any restarts will cause this to check to fail, and this check does not respect retries.
- `wait`: (default: `5` seconds): Number of seconds to wait between healthcheck attempts.

> [!WARNING]
> Healthchecks are implemented by specific scheduler plugins, and not all plugins support all options. Please consult the scheduler documentation for further details on what is supported.

See the [docker-container-healthchecker](https://github.com/dokku/docker-container-healthchecker) documentation for more details on how healthchecks are interpreted.

## Manually invoking checks

Checks can also be manually invoked via the `checks:run` command. This can be used to check the status of an application via cron to provide integration with external healthchecking software.

Checks are run against a specific application:

```shell
dokku checks:run APP
```

```
-----> Running pre-flight checks
-----> Running checks for app (APP.web.1)
       For more efficient zero downtime deployments, create a file CHECKS.
       See https://dokku.com/docs/deployment/zero-downtime-deploys/ for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
-----> Running checks for app (APP.web.2)
       For more efficient zero downtime deployments, create a file CHECKS.
       See https://dokku.com/docs/deployment/zero-downtime-deploys/ for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
-----> Running checks for app (APP.worker.1)
       For more efficient zero downtime deployments, create a file CHECKS.
       See https://dokku.com/docs/deployment/zero-downtime-deploys/ for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
```

Checks can be scoped to a particular process type:

```shell
dokku checks:run node-js-app worker
```

```
-----> Running pre-flight checks
-----> Running checks for app (APP.worker.1)
       For more efficient zero downtime deployments, create a file CHECKS.
       See https://dokku.com/docs/deployment/zero-downtime-deploys/ for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
```

An app process ID may also be specified:

```shell
dokku checks:run node-js-app web.2
```

```
-----> Running pre-flight checks
-----> Running checks for app (APP.web.2)
       For more efficient zero downtime deployments, create a file CHECKS.
       See https://dokku.com/docs/deployment/zero-downtime-deploys/ for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
```

Non-existent process types will result in an error:

```shell
dokku checks:run node-js-app non-existent
```

```
-----> Running pre-flight checks
Invalid process type specified (APP.non-existent)
```

Non-existent process IDs will *also* result in an error

```shell
dokku checks:run node-js-app web.3
```

```
-----> Running pre-flight checks
Invalid container id specified (APP.web.3)
```
