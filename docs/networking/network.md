# Network Management

> New as of 0.11.0

```
network:rebuild <app>                    # Rebuilds network settings for an app
network:rebuildall                       # Rebuild network settings for all apps
network:set <app> <key> (<value>)        # Set or clear a network property for an app
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

### Container network interface binding

> This functionality does not control the `--network` docker flag. Please use the [docker-options plugin](docs/advanced-usage/docker-options.md) to manage this flag.

By default, an application will only bind to the internal interface. This behavior can be modified per app by changing the `bind-all-interfaces` network property.

```shell
# bind to the default docker interface (`docker0`) with a random internal ip
# this is the default behavior
dokku network:set node-js-app bind-all-interfaces false

# bind to all interfaces (`0.0.0.0`) on a random port for each upstream port
# this will make the app container directly accessible by other hosts on your network
# ports are randomized for every deploy, e.g. `0.0.0.0:32771->5000/tcp`.
dokku network:set node-js-app bind-all-interfaces true
```

By way of example, in the default case, each container is bound to the docker interface:

```shell
docker ps
```

```
CONTAINER ID        IMAGE                      COMMAND                CREATED              STATUS              PORTS               NAMES
1b88d8aec3d1        dokku/node-js-app:latest   "/bin/bash -c '/star   About a minute ago   Up About a minute                       node-js-app.web.1
```

As such, the container's IP address will be an internal IP, and thus it is only accessible on the host itself:

```
docker inspect --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' node-js-app.web.1
```

```
172.17.0.6
```

However, you can disable the internal proxying via the `network:set` command so that it will listen on the host's IP address:

```shell
dokku network:set node-js-app bind-all-interfaces true

# container bound to all interfaces
docker ps
```

```
CONTAINER ID        IMAGE                      COMMAND                CREATED              STATUS              PORTS                     NAMES
d6499edb0edb        dokku/node-js-app:latest   "/bin/bash -c '/star   About a minute ago   Up About a minute   0.0.0.0:49153->5000/tcp   node-js-app.web.1
```
