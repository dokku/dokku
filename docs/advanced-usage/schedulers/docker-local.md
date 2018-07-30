# Docker Local Scheduler

> New as of 0.12.0

Dokku natively includes functionality to manage application lifecycles for a single server using the `scheduler-docker-local` plugin. It is the default scheduler, but as with all schedulers, it is set on a per-application basis. The scheduler can currently be overridden by running the following command:

```shell
dokku config:set node-js-app DOCKER_SCHEDULER=docker-local
```

As it is the default, unsetting the `DOCKER_SCHEDULER` config variable is also a valid way to reset the scheduler.

```shell
dokku config:unset node-js-app DOCKER_SCHEDULER
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
