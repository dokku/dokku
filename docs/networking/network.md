# Network Management

> New as of 0.10.0

```
network:rebuild <app>                    # Rebuilds network settings for an app
network:rebuildall                       # Rebuild network settings for all apps
```

The Network plugin allows developers to abstract the concept of container network management, allowing developers to both change what networks a given container is attached to as well as rebuild the configuration on the fly.

## Usage

### Rebuilding network settings

There are cases where you may need to rebuild the network configuration for an app, such as on app boot or container restart. In these cases, you can use the `network:rebuild` command:

```shell
dokku network:rebuild node-js-app
```

> This command will exit a non-zero number that depends on the number of containers for which configuration could not be built

### Rebuilding all network settings

In some cases, a docker upgrade may reset container IPs or Ports. In both cases, you can quickly rewrite those files by using the `network:rebuildall` command:

```shell
dokku network:rebuildall
```

> This command will exit a non-zero number that depends on the number of containers for which configuration could not be built
