# Network Management

> [!IMPORTANT]
> New as of 0.11.0, Enhanced in 0.20.0

```
network:create <network>                    # Creates an attachable docker network
network:destroy <network>                   # Destroys a docker network
network:exists <network>                    # Checks if a docker network exists
network:info <network> [--format text|json] # Outputs information about a docker network
network:list [--format text|json]           # Lists all docker networks
network:report [<app>] [<flag>]             # Displays a network report for one or more apps
network:rebuild <app>                       # Rebuilds network settings for an app
network:rebuildall                          # Rebuild network settings for all apps
network:set <app> <key> (<value>)           # Set or clear a network property for an app
```

The Network plugin allows developers to abstract the concept of container network management, allowing developers to both change what networks a given container is attached to as well as rebuild the configuration on the fly.

## Usage

### Listing networks

> [!IMPORTANT]
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

The `network:list` command also takes a `--format` flag, with the valid options including `text` (default) and `json`. The `json` output format can be used for automation purposes:

```shell
dokku network:list --format json
```

```
[
    {"CreatedAt":"2024-02-25T01:55:24.275184461Z","Driver":"bridge","ID":"d18df2d21433","Internal":false,"IPv6":false,"Labels":{},"Name":"bridge","Scope":"local"},
    {"CreatedAt":"2024-02-25T01:55:24.275184461Z","Driver":"bridge","ID":"f50fa882e7de","Internal":false,"IPv6":false,"Labels":{},"Name":"test-network","Scope":"local"},
    {"CreatedAt":"2024-02-25T01:55:24.275184461Z","Driver":"host","ID":"ab6a59291443","Internal":false,"IPv6":false,"Labels":{},"Name":"host","Scope":"local"},
    {"CreatedAt":"2024-02-25T01:55:24.275184461Z","Driver":"null","ID":"e2506bc8b7d7","Internal":false,"IPv6":false,"Labels":{},"Name":"none","Scope":"local"}
]
```

### Creating a network

> [!IMPORTANT]
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

> [!IMPORTANT]
> New as of 0.20.0, Requires Docker 1.21+

A Docker network without any associated containers may be destroyed via the `network:destroy` command. Docker will refuse to destroy networks that have containers attached.

```shell
dokku network:destroy test-network
```

```
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

```
-----> Destroying network test
```

### Checking if a network exists

> [!IMPORTANT]
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

> [!IMPORTANT]
> New as of 0.35.3

Network information can be retrieved via the `network:info` command. This is a slightly different version of the `docker network` command.

```shell
dokku network:info bridge
```

```
=====> bridge network information
       ID:       d18df2d21433
       Name:     bridge
       Driver:   bridge
       Scope:    local
```

The `network:info` command also takes a `--format` flag, with the valid options including `text` (default) and `json`. The `json` output format can be used for automation purposes:

```shell
dokku network:info bridge --format json
```

```
{"CreatedAt":"2024-02-25T01:55:24.275184461Z","Driver":"bridge","ID":"d18df2d21433","Internal":false,"IPv6":false,"Labels":{},"Name":"bridge","Scope":"local"}
```

### Routing an app to a known ip:port combination

In some cases, it may be necessary to route an app to an existing `$IP:$PORT` combination. This is particularly the case for internal admin tools or services that aren't run by Dokku but have a web ui that would benefit from being exposed by Dokku. This can be done by using a proxy application and routing requests through that.

```shell
# for a service listening on:
# - ip address: 127.0.0.1
# - port: 8080

# create the app
dokku apps:create local-app

# add an extra host that maps host.docker.internal to the docker gateway
dokku docker-options:add local-app deploy "--add-host=host.docker.internal:host-gateway"

# set the SERVICE_HOST to the mapped hostname
dokku config:set local-app SERVICE_HOST=host.docker.internal

# set the SERVICE_PORT to the port combination for your app
dokku config:set local-app SERVICE_PORT=8080

# set the domains desired
dokku domains:set local-app local-app.dokku.me

# deploy the service-proxy image
dokku git:from-image local-app dokku/service-proxy:latest
```

Only a single `$IP:$PORT` combination can be routed to for a given app, and that `$IP:$PORT` combination _must_ be accessible to the service proxy on initial deploy, or the service proxy won't start.

### Attaching an app to a network

> [!IMPORTANT]
> New as of 0.20.0, Requires Docker 1.21+

Apps will default to being associated with the default `bridge` network or a network specified by the `initial-network` network property. Additionally, an app can be attached to `attachable` networks by changing the `attach-post-create` or `attach-post-deploy` network properties when using the [docker-local scheduler](/docs/deployment/schedulers/docker-local.md). A change in these values will require an app deploy or rebuild.

```shell
# associates the network after a container is created but before it is started
# commonly used for cross-app networking
dokku network:set node-js-app attach-post-create test-network

# associates the network after the deploy is successful but before the proxy is updated
# used for cross-app networking when healthchecks must be invoked first
dokku network:set node-js-app attach-post-deploy other-test-network

# associates the network at container creation
# typically blocks access to services and external routing
dokku network:set node-js-app initial-network global-network
```

Multiple networks can also be specified for the `attach-post-create` and `attach-post-deploy` phases.

```shell
# one or more networks can be specified
dokku network:set node-js-app attach-post-create test-network test-network-2
```

Setting the `attach` network property to an empty value will de-associate the container with the network.

```shell
dokku network:set node-js-app attach-post-create
dokku network:set node-js-app attach-post-deploy
dokku network:set node-js-app initial-network
```

The network properties can also be set globally. The global default value is an empty string, and the global value is used when no app-specific value is set.

```shell
dokku network:set --global attach-post-create global-create-network
dokku network:set --global attach-post-deploy global-deploy-network
dokku network:set --global initial-network global-network
```

The default value may be set by passing an empty value for the option.

```shell
dokku network:set --global attach-post-create
dokku network:set --global attach-post-deploy
dokku network:set --global initial-network
```

#### Network Aliases

> [!NOTE]
> This feature is only available when an app has been attached to a network other than the default `bridge` network.

When a container created for a deployment is being attached to a network - regardless of which network property was used - a network alias of the pattern `APP.PROC_TYPE` will be added to all containers. This can be used to load-balance requests between containers. For an application named `node-js-app` with a process type of web, the network alias - or resolvable DNS record within the network - will be:

```
node-js-app.web
```

The fully-qualified URL for the resource will depend upon the `PORT` being listened to by the application. Applications built via buildpacks will have their `PORT` environment variable set to `5000`, and as such internal network requests for the above example should point to the following:

```
http://node-js-app.web:5000
```

Dockerfile-based applications may listen on other ports. For more information on how ports are specified for applications, please refer to the [port management documentation](/docs/networking/port-management.md).

#### Specifying a custom TLD

When attaching applications to networks, a custom TLD can be specified via the `network:set` command. This TLD is suffixed to the network alias for the application/process-type combination for _all_ networks to which the application is attached, and cannot be customized per network. The default value is an empty string.

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

The default value may be set by passing an empty value for the option:

```shell
dokku network:set node-js-app tld
```

The `tld` property can also be set globally. The global default is empty string, and the global value is used when no app-specific value is set.

```shell
dokku network:set --global tld svc.cluster.local
```

The default value may be set by passing an empty value for the option.

```shell
dokku network:set --global tld
```

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

> [!WARNING]
> If the attachment fails during the `running` container state, this may result in your application failing to respond to proxied requests once older containers are removed.

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

> This functionality does not control the `--network` docker flag. Please use the [docker-options plugin](/docs/advanced-usage/docker-options.md) to manage this flag.

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

```shell
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

The `bind-all-interfaces` property can also be set globally. The global default is `false`, and the global value is used when no app-specific value is set.

```shell
dokku network:set --global bind-all-interfaces true
```

The default value may be set by passing an empty value for the option.

```shell
dokku network:set --global bind-all-interfaces
```

### Displaying network reports for an app

You can get a report about the app's network status using the `network:report` command:

```shell
dokku network:report
```

```
=====> node-js-app network information
       Network attach post create:
       Network attach post deploy:
       Network bind all interfaces:   false
       Network computed attach post create:
       Network computed attach post deploy:
       Network computed bind all interfaces:false
       Network computed initial network:
       Network computed tld:
       Network global attach post create:
       Network global attach post deploy:
       Network global bind all interfaces:false
       Network global initial network:
       Network global tld:
       Network initial network:
       Network tld:
       Network web listeners: 172.17.0.1:5000
=====> python-sample network information
       Network attach post create:
       Network attach post deploy:
       Network bind all interfaces:   false
       Network computed attach post create:
       Network computed attach post deploy:
       Network computed bind all interfaces:false
       Network computed initial network:
       Network computed tld:
       Network global attach post create:
       Network global attach post deploy:
       Network global bind all interfaces:false
       Network global initial network:
       Network global tld:
       Network initial network:
       Network tld:
       Network web listeners:          172.17.0.2:5000
=====> ruby-sample network information
       Network attach post create:
       Network attach post deploy:
       Network bind all interfaces:   false
       Network computed attach post create:
       Network computed attach post deploy:
       Network computed bind all interfaces:false
       Network computed initial network:
       Network computed tld:
       Network global attach post create:
       Network global attach post deploy:
       Network global bind all interfaces:false
       Network global initial network:
       Network global tld:
       Network initial network:
       Network tld:
       Network web listeners:
```

You can run the command for a specific app also.

```shell
dokku network:report node-js-app
```

```
=====> node-js-app network information
       Network attach post create:
       Network attach post deploy:
       Network bind all interfaces:   false
       Network computed attach post create:
       Network computed attach post deploy:
       Network computed bind all interfaces:false
       Network computed initial network:
       Network computed tld:
       Network global attach post create:
       Network global attach post deploy:
       Network global bind all interfaces:false
       Network global initial network:
       Network global tld:
       Network initial network:
       Network tld:
       Network web listeners: 172.17.0.1:5000
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku network:report node-js-app --network-bind-all-interfaces
```
