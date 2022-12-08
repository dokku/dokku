# Nomad Scheduler

> Warning: This scheduler is not in Dokku core and thus functionality may change over time as the API stabilizes.

The [Nomad Scheduler Plugin](https://github.com/dokku/dokku-scheduler-nomad) is available free as an external plugin. Please see the plugin's [issue tracker](https://github.com/dokku/dokku-scheduler-nomad/issues) for more information on the status of the plugin.

For users that require additional functionality, please refer to the [Sponsoring Documentation](https://github.com/dokku/.github/blob/master/SPONSORING.md).

## Scheduler Interface

The following sections describe implemented scheduler functionality for the `nomad` scheduler.

### Implemented Commands and Triggers

This plugin implements various functionality through `plugn` triggers to integrate with the `nomad` cli for running apps on a Nomad cluster. The following functionality is supported by the `scheduler-nomad` plugin.

- `apps:destroy`
- `deploy`
- `ps:stop`

### Logging support

> Warning: Fetching app logs for the `logs` command is currently not implemented. Please consider using [Vector](https://vector.dev/docs/setup/installation/platforms/kubernetes/) or a similar tool to ship logs to another service or a third-party platform.
