# Kubernetes Scheduler

> Warning: This scheduler is not in Dokku core and thus functionality may change over time as the API stabilizes.

The [Kubernetes Scheduler Plugin](https://github.com/dokku/dokku-scheduler-kubernetes) implements the following functionality:

- `apps:destroy`
- `deploy`: partial, does not implement failed deploy log capture
- `logs`: partial, does not implement failure logs
- `ps:stop`

Please see the plugin's [issue tracker](https://github.com/dokku/dokku-scheduler-kubernetes/issues) for more information on the status of the plugin.

For users that require additional functionality, please refer to the [Sponsoring Documentation](https://github.com/dokku/.github/blob/master/SPONSORING.md).
