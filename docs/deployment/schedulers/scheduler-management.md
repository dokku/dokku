# Scheduler Management

> New as of 0.26.0

```
scheduler:report [<app>] [<flag>]   # Displays a scheduler report for one or more apps
scheduler:set <app> <key> (<value>) # Set or clear a scheduler property for an app
```

Schedulers are a way of customizing how an app image is deployed, and can be used to interact with non-local systems such as Kubernetes and Nomad.

## Usage

### Scheduler selection

Dokku supports the following built-in schedulers:

- `scheduler-docker-local`: Schedules apps against the local docker socket and runs containers directly on the Dokku host. See the [docker-local scheduler documentation](/docs/deployment/schedulers/docker-local.md) for more information on how this scheduler functions.
- `scheduler-null`: Does nothing during the scheduler phase. See the [null scheduler documentation](/docs/deployment/schedulers/null.md) for more information on how this scheduler functions.

See the [alternate schedulers documentation](/docs/deployment/schedulers/alternate-schedulers.md) for more information on other scheduler plugins.

### Overriding the auto-selected scheduler

If desired, the scheduler can be specified via the `scheduler:set` command by speifying a value for `selected`. The selected scheduler will always be used.

```shell
dokku scheduler:set node-js-app selected docker-local
```

The default value may be set by passing an empty value for the option:

```shell
dokku scheduler:set node-js-app selected
```

The `selected` property can also be set globally. The global default is an empty string, and auto-detection will be performed when no value is set per-app or globally.

```shell
dokku scheduler:set --global selected docker-local
```

The default value may be set by passing an empty value for the option.

```shell
dokku scheduler:set --global selected
```

### Displaying scheduler reports for an app

You can get a report about the app's scheduler status using the `scheduler:report` command:

```shell
dokku scheduler:report
```

```
=====> node-js-app scheduler information
       Scheduler computed selected:  herokuish
       Scheduler global selected: herokuish
       Scheduler selected: herokuish
=====> python-sample scheduler information
       Scheduler computed selected: dockerfile
       Scheduler global selected: herokuish
       Scheduler selected: dockerfile
=====> ruby-sample scheduler information
       Scheduler computed selected: herokuish
       Scheduler global selected: herokuish
       Scheduler selected:
```

You can run the command for a specific app also.

```shell
dokku scheduler:report node-js-app
```

```
=====> node-js-app scheduler information
       Scheduler selected: herokuish
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku scheduler:report node-js-app --scheduler-selected
```

### Custom schedulers

To create a custom scheduler, the following triggers may be implemented:

- `check-deploy`
- `core-post-deploy`
- `post-app-clone-setup`
- `post-app-rename-setup`
- `post-create`
- `post-delete`
- `pre-deploy`
- `pre-restore`
- `scheduler-app-status`
- `scheduler-deploy`
- `scheduler-docker-cleanup`
- `scheduler-inspect`
- `scheduler-is-deployed`
- `scheduler-logs`
- `scheduler-logs-failed`
- `scheduler-retire`
- `scheduler-run`
- `scheduler-stop`
- `scheduler-tags-create`
- `scheduler-tags-destroy`

Custom plugins names _must_ have the prefix `scheduler-` or scheduler overriding via `scheduler:set` may not function as expected.

Schedulers can use any tools available on the system to build the docker image, and may even be used to interact with off-server systems. The only current requirement is that the scheduler must have access to the image built in the build phase. If this is not the case, the registry plugin can be used to push the image to a registry that the scheduler software can access.
