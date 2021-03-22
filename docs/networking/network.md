# Network Management

> New as of 0.11.0, Enhanced in 0.20.0

```
network:create <network>                 # Creates an attachable docker network
network:destroy <network>                # Destroys a docker network
network:exists <network>                 # Checks if a docker network exists
network:info <network>                   # Outputs information about a docker network
network:list                             # Lists all docker networks
network:report [<app>] [<flag>]          # Displays a network report for one or more apps
network:rebuild <app>                    # Rebuilds network settings for an app
network:rebuildall                       # Rebuild network settings for all apps
network:set <app> <key> (<value>)        # Set or clear a network property for an app
```

The Network plugin allows developers to abstract the concept of container network management, allowing developers to both change what networks a given container is attached to as well as rebuild the configuration on the fly.

## Usage

### Listing networks

> New as of 0.20.0, Requires Docker 1.21+

You can easily list all available networks using the `network:list` command:

```shell
dokku network:list
```

```
=====> Networks
bridge
host
none
test-network
```

Note that you can easily hide extra output from Dokku commands by using the `--quiet` flag, which makes it easier to parse on the command line.

```shell
dokku --quiet network:list
```

```
bridge
host
none
test-network
```

### Creating a network

> New as of 0.20.0, Requires Docker 1.21+

Docker networks can be created via the `network:create` command. Executing this command will create an attachable `bridge` network. This can be used to route requests between containers without going through any public network.

```shell
dokku network:create test-network
```

```
-----> Creating network test-network
```

Specifying other additional flags or other types of networks can be created directly via the `docker` command.

### Destroying a network

> New as of 0.20.0, Requires Docker 1.21+

A Docker network without any associated containers may be destroyed via the `network:destroy` command. Docker will refuse to destroy networks that have containers attached.

```shell
dokku network:destroy test-network
```

```shell
 !     WARNING: Potentially Destructive Action
 !     This command will destroy network test.
 !     To proceed, type "test"
> test
-----> Destroying network test
```

As the command is destructive, it will default to asking for confirmation before executing the removal of the network. This may be avoided by providing the `--force` flag:

```shell
dokku --force network:destroy test-network
```

```shell
-----> Destroying network test
```

### Checking if a network exists

> New as of 0.20.0, Requires Docker 1.21+

For CI/CD pipelines, it may be useful to see if an network exists before creating a new network. You can do so via the `network:exists` command:

```shell
dokku network:exists nonexistent-network
```

```
Network does not exist
```

The `network:exists` command will return non-zero if the network does not exist, and zero if it does.

### Checking network info

> New as of 0.20.0, Requires Docker 1.21+

Network information can be retrieved via the `network:info` command. This is a slightly different version of the `docker network` command.

```shell
dokku network:info test-network
```

```
// TODO
```

### Attaching an app to a network

> New as of 0.20.0, Requires Docker 1.21+

Apps will default to being associated with the `bridge` network, but can be attached to `attachable` networks by changing the `attach-post-create` or `attach-post-deploy` network properties when using the [docker-local scheduler](/docs/advanced-usage/schedulers/docker-local.md). Additionally, it can be attached to an initial network via the `initial-network` property. A change in these values will require an app deploy or rebuild.

```shell
# associates the network after a container is created but before it is started
dokku network:set node-js-app attach-post-create test-network

# associates the network after the deploy is successful but before the proxy is updated
dokku network:set node-js-app attach-post-deploy other-test-network

# associates the network at container creation
dokku network:set node-js-app initial-network global-network
```

Setting the `attach` network property to an empty value will de-associate the container with the network.

```shell
dokku network:set node-js-app attach-post-create
dokku network:set node-js-app attach-post-deploy
dokku network:set node-js-app initial-network
```

Finally, the `initial-network` property can be set globally by using the `--global` flag in place of the app name.

```shell
dokku network:set --global initial-network global-network
```

#### Network Aliases

When a container created for a deployment is being attached to a network - regardless of which `attach` property was used - a network alias of the pattern `APP.PROC_TYPE` will be added to all containers. This can be used to load-balance requests between containers. For an application named `node-js-app` with a process type of web, the network alias - or resolvable DNS record within the network - will be:

```
node-js-app.web
```

The fully-qualified URL for the resource will depend upon the `PORT` being listened to by the application. Applications built via buildpacks will have their `PORT` environment variable set to `5000`, and as such internal network requests for the above example should point to the following:

```
http://node-js-app.web:5000
```

Dockerfile-based applications may listen on other ports. For more information on how ports are specified for applications, please refer to the [port management documentation](/docs/networking/port-management.md).

#### Specifying a custom TLD

When attaching applications to networks, a custom TLD can be specified via the `network:set` command. This TLD is suffixed to the network alias for the application/process-type combination for _all_ networks to which the application is attached, and cannot be customized per network.

To specify a TLD of `svc.cluster.local` for your application, run the following command:

```shell
# replace node-js-app with your application name
dokku network:set node-js-app tld svc.cluster.local
```

With an application named `node-js-app` and a process-type named `web`, the above command will turn the network alias into:

```shell
node-js-app.web.svc.cluster.local
```

Note that this has no impact on container port handling, and users must still specify the container port when making internal network requests.

#### When to attach containers to a network

Containers can be attached to a network for a variety of reasons:

- A background process in one app needs to communicate to a webservice in another app
- An app needs to talk to a container not managed by Dokku in a secure manner
- A custom network that allows transparent access to another host exists and is necessary for an app to run

Whatever the reason, the semantics of the two network hooks are important and are outlined before.

- `attach-post-create`:
  - Phase it applies to:
    - `build`: Intermediate containers created during the build process.
    - `deploy`: Deployed app containers.
    - `run`: Containers created by the `run` command.
  - Container state on attach: `created` but not `running`
  - Use case: When the container needs to access a resource on the network.
  - Example: The app needs to talk to a database on the same network when it first boots.
- `attach-post-deploy`
  - Phase it applies to:
    - `deploy`: Deployed app containers.
  - Container state on attach: `running`
  - Use case: When another container on the network needs to access _this_ container.
  - Example: A background process needs to communicate with the web process exposed by this container.
- `initial-network`:
  - Phase it applies to:
    - `build`: Intermediate containers created during the build process.
    - `deploy`: Deployed app containers.
    - `run`: Containers created by the `run` command.
  - Container state on attach: `created`
  - Use case: When another container on the network is already running and needed by this container.
  - Example: A key-value store exposing itself to all your apps may be on the `initial-network`.


> Warning: If the attachment fails during the `running` container state, this may result in your application failing to respond to proxied requests once older containers are removed.

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

By default, an app will only bind to the internal interface. This behavior can be modified per app by changing the `bind-all-interfaces` network property.

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

### Displaying network reports for an app

You can get a report about the app's network status using the `network:report` command:

```shell
dokku network:report
```

```
=====> node-js-app network information
       Network bind all interfaces: false
       Network listeners: 172.17.0.1:5000
=====> python-sample network information
       Network bind all interfaces: false
       Network listeners: 172.17.0.2:5000
=====> ruby-sample network information
       Network bind all interfaces: true
       Network listeners:
```

You can run the command for a specific app also.

```shell
dokku network:report node-js-app
```

```
=====> node-js-app network information
       Network bind all interfaces: false
       Network listeners: 172.17.0.1:5000
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku network:report node-js-app --network-bind-all-interfaces
```
