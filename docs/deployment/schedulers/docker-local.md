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

### Deploying Process Types in Parallel

> New as of 0.25.5

By default, Dokku deploys an app's processes one-by-one in order, with the `web` process being deployed first. Deployment parallelism may be achieved by setting the `parallel-schedule-count` property, which defaults to `1`. Increasing this number increases the number of process types that may be deployed in parallel (with the web process being the exception).

```shell
# Increase parallelism from 1 process type at a time to 4 process types at a time.
dokku scheduler-docker-local:set node-js-app parallel-schedule-count 4
```

Once set, you may reset it by setting a blank value for `parallel-schedule-count`:

```shell
dokku scheduler-docker-local:set node-js-app parallel-schedule-count
```

If the value of `parallel-schedule-count` is increased and a given process type fails to schedule successfully, then any in-flight process types will continue to be processed, while all process types that have not been scheduled will be skipped before the deployment finally fails.

Container scheduling output is shown in the order it is received, and thus may be out of order in case of output to stderr.

Note that increasing the value of `parallel-schedule-count` may significantly impact CPU utilization on your host as your app containers - and their respective processes - start up. Setting a value higher than the number of available CPUs is discouraged. It is recommended that users carefully set this value so as not to overburden their server.

#### Increasing parallelism within a process deploy

> New as of 0.26.0

By default, Dokku will deploy one instance of a given process type at a time. This can be increased by customizing the `app.json` `formation` key to include a `max_parallel` key for the given process type.

An `app.json` file can be committed to the root of the pushed app repository, and must be within the built image artifact in the image's working directory as shown below.

- Buildpacks: `/app/app.json`
- Dockerfile: `WORKDIR/app.json` or `/app.json` (if no working directory specified)
- Docker Image: `WORKDIR/app.json` or `/app.json` (if no working directory specified)

The `formation` key should be specified as follows in the `app.json` file:

```Procfile
{
  "formation": {
    "web": {
      "max_parallel": 1
    },
    "worker": {
      "max_parallel": 4
    }
  }
}
```

Omitting or removing the entry will result in parallelism for that process type to return to 1 entry at a time. This can be combined with the  `parallel-schedule-count` property to speed up deployments.

Note that increasing the value of `max_parallel` may significantly impact CPU utilization on your host as your app containers - and their respective processes - start up. Setting a value higher than the number of available CPUs is discouraged. It is recommended that users carefully set this value so as not to overburden their server.

## Implemented Triggers

This plugin implements various functionality through `plugn` triggers to integrate with Docker for running apps on a single server. The following functionality is supported by the `scheduler-docker-local` plugin.

- `apps:clone`
- `apps:destroy`
- `apps:rename`
- `deploy`
- `enter`
- `logs`
- `ps:inspect`
- `ps:stop`
- `run`

## Supported Resource Management Properties

The `docker-local` scheduler supports a minimal list of resource _limits_ and _reservations_. The following properties are supported:

### Resource Limits

- cpu: (docker option: `--cpus`), is specified in number of CPUs a process can access.
  - See the ["CPU" section](https://docs.docker.com/config/containers/resource_constraints/#cpu) of the Docker Runtime Options documentation for more information.
- memory: (docker option: `--memory`) should be specified with a suffix of `b` (bytes), `k` (kilobytes), `m` (megabytes), `g` (gigabytes). Default unit is `m` (megabytes).
  - See the ["Memory" section](https://docs.docker.com/config/containers/resource_constraints/#memory) of the Docker Runtime Options documentation for more information.
- memory-swap: (docker option: `--memory-swap`) should be specified with a suffix of `b` (bytes), `k` (kilobytes), `m` (megabytes), `g` (gigabytes)
  - See the ["Memory" section](https://docs.docker.com/config/containers/resource_constraints/#memory) of the Docker Runtime Options documentation for more information.
- nvidia-gpus: (docker option: `--gpus`), is specified in number of Nvidia GPUs a process can access.
  - See the ["GPU" section](https://docs.docker.com/config/containers/resource_constraints/#gpu) of the Docker Runtime Options documentation for more information.

### Resource Reservations

- memory: (docker option: `--memory-reservation`) should be specified with a suffix of `b` (bytes), `k` (kilobytes), `m` (megabytes), `g` (gigabytes). Default unit is `m` (megabytes).
  - See the ["Memory" section](https://docs.docker.com/config/containers/resource_constraints/#memory) of the Docker Runtime Options documentation for more information.
