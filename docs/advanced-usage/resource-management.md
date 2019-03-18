# Resource Management

> New as of 0.15.0

```
resource:limit [--process-type <process-type>] [RESOURCE_OPTS...] <app>    # Limit resources for a given app/process-type combination
resource:limit-clear [--process-type <process-type>] <app>                 # Limit resources for a given app/process-type combination
resource:report [<app>] [<flag>]                                           # Displays a resource report for one or more apps
resource:reserve [--process-type <process-type>] [RESOURCE_OPTS...] <app>  # Reserve resources for a given app/process-type combination
resource:reserve-clear [--process-type <process-type>] <app>               # Reserve resources for a given app/process-type combination
```

The resource plugin is meant to allow users to limit or reserve resources for a given app/process-type combination.

## Usage

By default, Dokku allows unlimited resource access to apps deployed on a server. In some cases, it may be desirable to limit this on a per-app or per-process-type basis. The `resource` plugin allows management of both resource "limits" and resource "reservations", where each resource request type has specific meaning to the scheduler in use for a given app.

> The meaning of a values and it's units are specific to the scheduler in use for a given app. If a scheduler does not support a given resource type combination, it will be ignored. All resource commands require an app rebuild or deploy in order to take effect.

Valid resource options include:

- `--cpu`
- `--memory`
- `--memory-swap`
- `--network`
- `--network-ingress`
- `--network-egress`

### Resource Limits

When specified and supported, a resource limit will ensure that your app does not go _over_ the specified value. If this occurs, the underlying scheduler may either cap resource utilization, or it may decide to terminate and reschedule your process.

Resource limits may be set via the `resource:limit` command:

```shell
dokku resource:limit --memory 100 node-js-app
```

```
=====> Setting resource limits for node-js-app
       memory: 100
```

Multiple resources can be limited in a single call:

```shell
dokku resource:limit --cpu 100 --memory 100 node-js-app
```

```
=====> Setting resource limits for node-js-app
       cpu: 100
       memory: 100
```

Resources can also be limited on a per-process type basis. If specified, this will override any generic limits set for the app.

```shell
dokku resource:limit --cpu 100 --memory 100 --process-type worker node-js-app
```

```
=====> Setting resource limits for node-js-app (worker)
       cpu: 100
       memory: 100
```

#### Displaying Resource Limits

Running the `resource:limits` command without any flags will display the currently configured app reservation.

```shell
dokku resource:limits node-js-app
=====> resource limits node-js-app information
       cpu: 100
       memory: 100
       memory-swap:
       network:
       network-ingress:
       network-egress:
```

#### Global Resource Limits

By default, Dokku does not limit any resources, though these can be specified at the server level via the `--global` flag on the `resource:limit` command.

```shell
dokku resource:limit --global --memory 200
```

```
=====> Setting default resource limits
       memory: 200
```

This may also be combined with the `--process-type` flag to set global limits on a process-type level:

```shell
dokku resource:limit --global --memory 150 --process-type web node-js-app
```

```
=====> Setting default resource limits (web)
       memory: 150
```

#### Displaying Global Resource Limits

Running the `resource:limit` command with only the `--global` flag will display the currently configured default limits.

```shell
dokku resource:limit --global
```

```
=====> resource limits default information
       cpu:
       memory: 200
       memory-swap:
       network:
       network-ingress:
       network-egress:
```

This may also be combined with the `--process-type` flag to see global limits on a process-type level:

```shell
dokku resource:limit --global --process-type web
```

```
=====> resource limits default information (web)
       cpu:
       memory-swap:
       memory: 150
       network-egress:
       network-ingress:
       network:
```

#### Clearing Resource Limits

In cases where the values are incorrect - or there is no desire to limit resources - resource limits may be cleared using the `resource:limit-clear` command.

```shell
dokku resource:limit-clear node-js-app
```

```
-----> Clearing resource limit for node-js-app
```

Defaults can also be cleared by leaving the app unspecified.

```shell
dokku resource:limit-clear
```

```
-----> Clearing default resource limits
```

### Resource Reservations

When specified and supported, a resource reservation will ensure that your server has _at least_ the specified resources before placing a given app's process. If there a resource exhaustion, future rebuilds and deploys may fail.

Resource reservations may be set via the `resource:reserve` command:

```shell
dokku resource:reserve --memory 100 node-js-app
```

```
=====> Setting resource reservation for node-js-app
       memory: 100
```

Multiple resources can be limited in a single call:

```shell
dokku resource:reserve --cpu 100 --memory 100 node-js-app
```

```
=====> Setting resource reservation for node-js-app
       cpu: 100
       memory: 100
```

Resources can also be limited on a per-process type basis. If specified, this will override any generic limits set for the app.

```shell
dokku resource:reserve --cpu 100 --memory 100 --process-type worker node-js-app
```

```
=====> Setting resource reservation for node-js-app (worker)
       cpu: 100
       memory: 100
```

#### Displaying Resource Reservations

Running the `resource:reserve` command without any flags will display the currently configured app reservation.

```shell
dokku resource:reserve node-js-app
=====> resource reservation node-js-app information
       cpu: 100
       memory: 100
       memory-swap:
       network:
       network-ingress:
       network-egress:
```

#### Global Resource Reservations

By default, Dokku does not reserve any resources, though these can be specified at the server level via the `--global` flag on the `resource:reserve` command.

```shell
dokku resource:reserve --global --memory 200
```

```
=====> Setting default resource reservation
       memory: 200
```

This may also be combined with the `--process-type` flag to set global reservations on a process-type level:

```shell
dokku resource:reserve --global --memory 150 --process-type web node-js-app
```

```
=====> Setting default resource reservation (web)
       memory: 150
```

#### Displaying Global Resource Reservations

Running the `resource:reserve` command with only the `--global` flag will display the currently configured default reservation.

```shell
dokku resource:reserve --global
```

```
=====> resource reservation default information
       cpu:
       memory: 200
       memory-swap:
       network:
       network-ingress:
       network-egress:
```

This may also be combined with the `--process-type` flag to see global reservations on a process-type level:

```shell
dokku resource:reserve --global --process-type web
```

```
=====> resource reservation default information (web)
       cpu:
       memory-swap:
       memory: 150
       network-egress:
       network-ingress:
       network:
```

#### Clearing Resource Reservations

In cases where the values are incorrect - or there is no desire to reserve resources - resource reservations may be cleared using the `resource:reserve-clear` command.

```shell
dokku resource:reserve-clear node-js-app
```

```
-----> Clearing resource reservation for node-js-app
```

Defaults can also be cleared by leaving the app unspecified.

```shell
dokku resource:reserve-clear
```

```
-----> Clearing default resource reservation
```

### Displaying resource reports for an app

You can get a report about the app's resource status using the `resource:report` command:

```shell
dokku resource:report
```

```
=====> node-js-app resource information
       Resource limit cpu:
       Resource limit memory: 1024
       Resource limit memory swap: 0
       Resource limit network: 10
       Resource limit network ingress:
       Resource limit network egress:
       Resource reservation cpu:
       Resource reservation memory: 512
       Resource reservation memory swap:
       Resource reservation network: 8
       Resource reservation network ingress:
       Resource reservation network egress:
=====> python-sample resource information
       Resource limit cpu:
       Resource limit memory:
       Resource limit memory swap:
       Resource limit network:
       Resource limit network ingress:
       Resource limit network egress:
       Resource reservation cpu:
       Resource reservation memory:
       Resource reservation memory swap:
       Resource reservation network:
       Resource reservation network ingress:
       Resource reservation network egress:
=====> ruby-sample resource information
       Resource limit cpu:
       Resource limit memory:
       Resource limit memory swap:
       Resource limit network:
       Resource limit network ingress:
       Resource limit network egress:
       Resource reservation cpu:
       Resource reservation memory:
       Resource reservation memory swap:
       Resource reservation network:
       Resource reservation network ingress:
       Resource reservation network egress:
```

You can run the command for a specific app also.

```shell
dokku resource:report node-js-app
```

```
=====> node-js-app resource information
       Resource limit cpu:
       Resource limit memory: 1024
       Resource limit memory swap: 0
       Resource limit network: 10
       Resource limit network ingress:
       Resource limit network egress:
       Resource reservation cpu:
       Resource reservation memory: 512
       Resource reservation memory swap:
       Resource reservation network: 8
       Resource reservation network ingress:
       Resource reservation network egress:
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku resource:report node-js-app --resource-memory
```
