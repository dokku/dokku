# Kubernetes Scheduler

> Warning: This scheduler is not in Dokku core and thus functionality may change over time as the API stabilizes.

The [Kubernetes Scheduler Plugin](https://github.com/dokku/dokku-scheduler-kubernetes) is available free as an external plugin. Please see the plugin's [issue tracker](https://github.com/dokku/dokku-scheduler-kubernetes/issues) for more information on the status of the plugin.

For users that require additional functionality, please refer to the [Sponsoring Documentation](https://github.com/dokku/.github/blob/master/SPONSORING.md).

## Scheduler Interface

The following sections describe implemented scheduler functionality for the `kubernetes` scheduler.

### Implemented Commands and Triggers

This plugin implements various functionality through `plugn` triggers to integrate with `kubectl` for running apps on a Kubernetes cluster. The following functionality is supported by the `scheduler-kubernetes` plugin.

- `apps:destroy`
- `deploy`: partial, does not implement failed deploy log capture
- `logs`: partial, does not implement failure logs
- `ps:stop`

### Logging support

App logs for the `logs` command are fetched from running pods via the `kubectl` cli. To persist logs across deployments, consider using [Vector](https://vector.dev/docs/setup/installation/platforms/kubernetes/) or a similar tool to ship logs to another service or a third-party platform.
