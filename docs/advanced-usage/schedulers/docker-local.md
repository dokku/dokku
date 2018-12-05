# Docker Local Scheduler

> Subcommands new as of 0.12.12

```
scheduler-docker-local:report [<app>] [<flag>]              # Displays a scheduler-docker-local report for one or more apps
scheduler-docker-local:set <app> <key> (<value>)            # Set or clear a scheduler-docker-local property for an app
```

> New as of 0.12.0

Dokku natively includes functionality to manage application lifecycles for a single server using the `scheduler-docker-local` plugin. It is the default scheduler, but as with all schedulers, it is set on a per-application basis. The scheduler can currently be overridden by running the following command:

```shell
dokku config:set node-js-app DOCKER_SCHEDULER=docker-local
```

As it is the default, unsetting the `DOCKER_SCHEDULER` config variable is also a valid way to reset the scheduler.

```shell
dokku config:unset node-js-app DOCKER_SCHEDULER
```

## Usage

### Disabling chown of persistent storage

The `scheduler-docker-local` plugin will ensure your storage mounts are owned by either `herokuishuser` or the overridden value you have set in `DOKKU_APP_USER`. You may disable this by running the following `scheduler-docker-local:set` command for your application:

```shell
dokku scheduler-docker-local:set node-js-app disable-chown true
```

Once set, you may re-enable it by setting a blank value for `disable-chown`:

```shell
dokku scheduler-docker-local:set node-js-app disable-chown
```

## Implemented Triggers

This plugin implements various functionality through `plugn` triggers to integrate with Docker for running apps on a single server. The following functionality is supported by the `scheduler-docker-local` plugin.

- `check-deploy`
- `core-post-deploy`
- `post-delete`
- `pre-deploy`
- `pre-restore`
- `scheduler-deploy`
- `scheduler-docker-cleanup`
- `scheduler-logs-failed`
- `scheduler-run`
- `scheduler-stop`
- `scheduler-tags-create`
- `scheduler-tags-destroy`
