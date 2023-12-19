# K3s Scheduler

```
scheduler-k3s:initialize [<--local>|<--remote user:password@host:port>] # initializes a cluster
scheduler-k3s:add-server [user:password@host:port]                      # adds a server node to a Dokku-managed cluster
scheduler-k3s:add-client [user:password@host:port]                      # adds a client node to a Dokku-managed cluster
scheduler-k3s:set [<app>|--global] <key> (<value>)                      # Set or clear a scheduler-k3s property for an app or the scheduler
```

> [!IMPORTANT]
> New as of 0.33.0

For multi-server support, Dokku provides the ability for users to setup a K3s cluster. As with all schedulers, it is set on a per-app basis. The scheduler can currently be overridden by running the following command:

```shell
dokku scheduler:set node-js-app selected k3s
```

As it is the default, unsetting the `selected` scheduler property is also a valid way to reset the scheduler.

```shell
dokku scheduler:set node-js-app k3s
```

## Tutorial

### Single-node cluster

Initialize the cluster in single-node mode. This will start k3s on the Dokku node itself.

```shell
dokku scheduler-k3s:initialize
```

The above command will initialize a cluster with an automatically generated token for authentication. This token should be stored securely for later recovery, and can be displayed with via the `scheduler-k3s:report` command:

```shell
dokku scheduler-k3s:report --global
```

```
=====> --global scheduler-k3s information
       scheduler k3s token: topsecret:server:token
```

Setup registry authentication. The K3s scheduler plugin will automatically pick up any newly configured registry backends and ensure nodes in the cluster have the credentials in place for pulling images from the cluster.

```shell
# hub.docker.com
dokku registry:login hub.docker.com $USERNAME $PASSWORD
```

To ensure images are pushed and pulled from the correct registry, set the correct `server` registry property. This can be set on a per-app basis, but we will set it globally here for this tutorial.

```shell
dokku registry:set --global server hub.docker.com
```

If using docker hub, you'll need to use a custom repository name. As it is unique, this must be set on a per-app basis.

```shell
dokku registry:set node-js-app image-repo my-awesome-prefix/node-js-app
```

Additionally, apps should be configured to push images on the release phase via the `push-on-release` registry property.

```shell
dokku registry:set --global push-on-release true
```

As routing is handled by traefik managed on the k3s plugin, set the proxy plugin to `k3s` as well. This implements a traefik-plugin compatible layer, allowing users to otherwise interact with the `traefik` plugin as expected.

```shell
dokku proxy:set --global k3s
```

Finally, set the scheduler to `k3s` so that app deploys will work on k3s.

```shell
dokku scheduler:set --global selected k3s
```

At this point, all app deploys will be performed against the k3s cluster.

### Running a multi-cluster node

For high-availability, it is recommended to add both agent and server nodes to the cluster. Dokku will default to starting the cluster with an embedded Etcd database backend, and is ready to add new agent or server nodes immediately.

Multiple server nodes can be added with the `scheduler-k3s:add-server` command. This will ssh onto the specified server, install k3s, and join it to the current Dokku node in server mode.

> [!WARNING]
> Certain ports must be open for cross-server communication. Refer to the [K3s networking documentation](https://docs.k3s.io/installation/requirements?os=debian#networking) for the required open ports between servers prior to running the command.

```shell
dokku scheduler-k3s:add-server root@server-1.example.com
```

Server nodes are typically used to replicate the cluster state, and it is recommended to have an odd number of nodes spread across several availability zones (datacenters in close proximity within a region). This allows for higher availability in the event of a cluster failure. Server nodes do not run any app workloads, but run control-plane services such as the traefik load balancer and the etcd backing store.

> [!NOTE]
> Only the initial Dokku server will be properly configured for deployment, and should be considered your git remote. Additional server nodes are for ensuring high-availability of the K3s etcd state. Ensure this server is properly backed up and restorable or deployments will not work.

Agent nodes are used to run. To add an agent, run the `add-agent` command. This will ssh onto the specified server, install k3s, and join it to the current Dokku node in agent mode. Agents are typically used to run app workloads.

```shell
dokku scheduler-k3s:add-agent root@server-1.example.com
```

When attaching an agent or server node, the K3s plugin will look at the IP associated with the `eth0` interface and use that to connect the new node to the cluster. To change this, set the `network-interface` property to the appropriate value.

```shell
dokku scheduler-k3s:set --global network-interface eth1
```

Dokku does not manage the ssh key of the server, but the value that should be added to the remote root user's `/root/.ssh/authorized_keys` file can be checked with the `git:public-key` command:

```shell
dokku git:public-key
```

### Joining an existing cluster

> [!IMPORTANT]
> This is future planned behavior. Commands here will not work as expected.

In some cases, Dokku must be setup from scratch and joined to an existing cluster. To do so, first set the k3s token value:

```shell
dokku scheduler-k3s:set --global token topsecret:server:token
```

Next, run the `scheduler-k3s:join-cluster` command. This command takes a single server node in the cluster, will SSH onto the specified server, retrieve the kubeconfig, and then configure kubectl on the Dokku node to speak with the specified k3s cluster. Dokku will also periodically retrieve all server nodes in the cluster so that server nodes can be safely replaced as needed.

```shell
dokku scheduler-k3s:join-cluster root@server-1.example.com
```

### Using kubectl remotely

> [!WARNING]
> Certain ports must be open for interacting with the remote kubernets api. Refer to the [K3s networking documentation](https://docs.k3s.io/installation/requirements?os=debian#networking) for the required open ports between servers prior to running the command.

By default, Dokku assumes that all it controls all actions on the cluster, and thus does not expose the `kubectl` binary for administrators. To interact with kubectl, you will need to retrieve the `kubeconfig` for the cluster and configure your client to use that configuration.

```shell
dokku scheduler-k3s:show-kubeconfig
```

## Scheduler Interface

The following sections describe implemented and unimplemented scheduler functionality for the `k3s` scheduler.

### Implemented Commands and Triggers

This plugin implements various functionality through `plugn` triggers to integrate with Docker for running apps on a single server. The following functionality is supported by the `scheduler-docker-local` plugin.

- `apps:destroy`
- `deploy`
- `logs`
- `ps:stop`

### Unimplemented command functionality

- `apps:clone`
- `apps:rename`
- `ps:inspect`
- `enter`
- `run`
- `run:detached`

The following Dokku functionality is not implemented at this time.

- `cron`
- `vector` log integration
- one-off tasks
- persistent storage

### Logging support

App logs for the `logs` command are fetched by Dokku from running containers via the `kubectl` cli. Persisting logs via Vector is not implemented at this time.

### Supported Resource Management Properties

The `docker-local` scheduler supports a minimal list of resource _limits_ and _reservations_. The following properties are supported:

#### Resource Limits

- cpu: is specified in number of CPUs a process can access.
- memory: should be specified with a suffix of `b` (bytes), `k` (kilobytes), `m` (megabytes), `g` (gigabytes). Default unit is `m` (megabytes).

#### Resource Reservations

- cpu: is specified in number of CPUs a process can access.
- memory: should be specified with a suffix of `b` (bytes), `k` (kilobytes), `m` (megabytes), `g` (gigabytes). Default unit is `m` (megabytes).
