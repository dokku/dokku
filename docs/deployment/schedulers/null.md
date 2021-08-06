# Null Scheduler

> New as of 0.25.0

The `null` scheduler does nothing, and is useful for routing to services not managed by Dokku. It should not be used in normal operation. Please see the [network documentation](/docs/networking/network.md#routing-an-app-to-a-known-ip:port-combination) for more information on the aforementioned use case.

## Usage

### Detection

This scheduler is _never_ auto-detected. The scheduler _must_  be specified via the `config:set` command:

```shell
dokku config:set node-js-app DOCKER_SCHEDULER=null
```
