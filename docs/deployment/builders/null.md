# Null Builder

> New as of 0.25.0

The `null` builder does nothing, and is useful for routing to services not managed by Dokku. It should not be used in normal operation. Please see the [network documentation](/docs/networking/network.md#routing-an-app-to-a-known-ip:port-combination) for more information on the aforementioned use case.

## Usage

### Detection

This builder is _never_ auto-detected. The builder _must_  be specified via the `builder:set` command:

```shell
dokku builder:set node-js-app selected null
```
