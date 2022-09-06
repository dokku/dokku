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

## Implementing a Scheduler

Custom plugins names _must_ have the prefix `scheduler-` or scheduler overriding via `scheduler:set` may not function as expected.

At this time, the following dokku commands are used to implement a complete scheduler.

- `apps:destroy`: stops the app processes on the scheduler
  - triggers: post-delete, scheduler-register-retired, scheduler-retire
- `apps:rename`: handles app renaming
  - triggers: post-app-rename-setup
- `apps:clone`: handles app cloning
  - triggers: post-app-clone-setup
- `deploy`: deploys app proceses and checks the status of a deploy
  - triggers: scheduler-app-status, scheduler-deploy, scheduler-is-deployed, scheduler-logs-failed
- `enter`: enters a running container
  - triggers: scheduler-enter
- `logs`: fetches app logs
  - triggers: scheduler-logs
- `run`: starts one-off run containers (detached and non-detached) as well as listing run processes
  - triggers: scheduler-run, scheduler-run-list
- `ps:stop`: stops app processes
  - triggers: scheduler-stop
- `ps:inspect`: outputs inspect output for processes in an app
  - triggers: scheduler-inspect

Schedulers may decide to omit some functionality here, or use plugin triggers to supplement config with information from other plugins. Additionally, a scheduler may implement other triggers in order handle any extra processes needed during a deploy.

Schedulers can use any tools available on the system to build the docker image, and may even be used to interact with off-server systems. The only current requirement is that the scheduler must have access to the image built in the build phase. If this is not the case, the registry plugin can be used to push the image to a registry that the scheduler software can access.

Deployment tasks are currently executed directly on the primary Dokku server.
