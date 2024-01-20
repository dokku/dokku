# K3s Scheduler

```
scheduler-k3s:initialize                           # initializes a cluster
scheduler-k3s:add-server [user:password@host:port] # adds a server node to a Dokku-managed cluster
scheduler-k3s:add-client [user:password@host:port] # adds a client node to a Dokku-managed cluster
scheduler-k3s:set [<app>|--global] <key> (<value>) # Set or clear a scheduler-k3s property for an app or the scheduler
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
# must be run as root
dokku scheduler-k3s:initialize
```

The above command will initialize a cluster with the following configuration:

- etcd distributed backing store
- Wireguard as the networking flannel
- K3s automatic upgrader
- Longhorn distributed block storage
- Traefik configured to run on all nodes in the cluster

Additionally, an internal token for authentication will be automatically generated. This token should be stored securely for later recovery, and can be displayed with via the `scheduler-k3s:report` command:

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

If using docker hub, you'll need to use a custom repository name. This can be set via a global template, allowing users access to the app name as the variable `AppName` as shown below.

```shell
dokku registry:set --global image-repo-template "my-awesome-prefix/{{ .AppName }}"
```

Additionally, apps should be configured to push images on the release phase via the `push-on-release` registry property.

```shell
dokku registry:set --global push-on-release true
```

As routing is handled by traefik managed on the k3s plugin, set the proxy plugin to `k3s` as well.

```shell
dokku proxy:set --global k3s
```

Ensure any other proxy implementations are disabled. Typically at least nginx will be running on the Dokku host and should be stopped if the host is used as load balancer.

```shell
dokku nginx:stop
```

Finally, set the scheduler to `k3s` so that app deploys will work on k3s.

```shell
dokku scheduler:set --global selected k3s
```

At this point, all app deploys will be performed against the k3s cluster.

> [!NOTE]
> HTTP requests for apps can be performed against any node in the cluster. Without extra configuration, many other ports may also be available on the host. For security reasons, it may be desirable to place the k3s cluster behind one or more TCP load balancers while shutting off traffic to all cluster ports. Please consult your hosting provider for more information on how to provision a TCP load balancer and shut off all ports other than 22/80/443 access to the outside world.

### Running a multi-cluster node

> [!WARNING]
> Certain ports must be open for cross-server communication. Refer to the [K3s networking documentation](https://docs.k3s.io/installation/requirements?os=debian#networking) for the required open ports between servers prior to running the command.

For high-availability, it is recommended to add both agent and server nodes to the cluster. Dokku will default to starting the cluster with an embedded Etcd database backend, and is ready to add new agent or server nodes immediately.

When attaching an agent or server node, the K3s plugin will look at the IP associated with the `eth0` interface and use that to connect the new node to the cluster. To change this, set the `network-interface` property to the appropriate value.

```shell
dokku scheduler-k3s:set --global network-interface eth1
```

Dokku will connect to remote servers via the `root` user with the `dokku` user's SSH key pair. Dokku servers may not have an ssh key pair by default, but they can be generated as needed via the `git:generate-deploy-key` command.

```shell
dokku git:generate-deploy-key
```

This key can then be displayed with the `git:public-key` command, and copied to the remote server's `/root/.ssh/authorized_keys` file.

```shell
dokku git:public-key
```

Multiple server nodes can be added with the `scheduler-k3s:cluster-add` command. This will ssh onto the specified server, install k3s, and join it to the current Dokku node in server mode.

```shell
dokku scheduler-k3s:cluster-add  --role server ssh://root@server-1.example.com
```

Server nodes allow any workloads to be scheduled on them by default, in addition to the control-plane, etcd, and the scheduler itself. To avoid app workloads being scheduled on your control-plane, use the `--taint-scheduling` flag:

```shell
dokku scheduler-k3s:cluster-add --role server --taint-scheduling ssh://root@cluster-1.example.com
```

If the server isn't in the `known_hosts` file, the connection will fail. This can be bypassed by setting the `--insecure-allow-unknown-hosts` flag:

```shell
dokku scheduler-k3s:cluster-add --role server --insecure-allow-unknown-hosts ssh://root@worker-1.example.com
```

Server nodes are typically used to replicate the cluster state, and it is recommended to have an odd number of nodes spread across several availability zones (datacenters in close proximity within a region). This allows for higher availability in the event of a cluster failure. Server nodes run control-plane services such as the traefik load balancer and the etcd backing store.

> [!NOTE]
> Only the initial Dokku server will be properly configured for push deployment, and should be considered your git remote. Additional server nodes are for ensuring high-availability of the K3s etcd state. Ensure this server is properly backed up and restorable or deployments will not work.

Agent nodes are used to run. To add an agent, run the `scheduler-k3s:cluster-add` with the `--role worker` flag. This will ssh onto the specified server, install k3s, and join it to the current Dokku node in agent mode. Agents are typically used to run app workloads.

```shell
dokku scheduler-k3s:cluster-add --role worker ssh://root@worker-1.example.com
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

This plugin implements various functionality through `plugn` triggers to integrate with Docker for running apps on a single server. The following functionality is supported by the `scheduler-k3s` plugin.

- `apps:clone`
- `apps:destroy`
- `apps:rename`
- `cron`
- `enter`
- `deploy`
- healthchecks
       - Due to Kubernetes limitations, only a single healthcheck is supported for each of the `liveness`, `readiness`, and `startup` healthchecks
       - Due to Kubernetes limitations, content checks are not supported
       - Ports specified in the `app.json` are ignored in favor of the container port on the port mapping detected
- `logs`
- `ps:stop`
- `run`
       - The `scheduler-post-run` trigger is not always triggered
- `run:detached`
- `run:list`

### Unimplemented command functionality

- `run:logs`
- `ps:inspect`

The following Dokku functionality is not implemented at this time.

- `vector` log integration
- persistent storage

### Logging support

App logs for the `logs` command are fetched by Dokku from running containers via the `kubectl` cli. Persisting logs via Vector is not implemented at this time. Users may choose to configure the Vector Kubernetes integration directly by following [this guide](https://vector.dev/docs/setup/installation/platforms/kubernetes/).

### Supported Resource Management Properties

The `k3s` scheduler supports a minimal list of resource _limits_ and _reservations_. The following properties are supported:

#### Resource Limits

> [!NOTE]
> Cron tasks retrieve resource limits based on the computed cron task ID. If unspecified, the default will be 1 CPU and 512m RAM.

- cpu: is specified in number of CPUs a process can access.
- memory: should be specified with a suffix of `b` (bytes), `k` (kilobytes), `m` (megabytes), `g` (gigabytes). Default unit is `m` (megabytes).

#### Resource Reservations

> [!NOTE]
> Cron tasks retrieve resource reservations based on the computed cron task ID. If unspecified, the default will be 1 CPU and 512m RAM.

- cpu: is specified in number of CPUs a process can access.
- memory: should be specified with a suffix of `b` (bytes), `k` (kilobytes), `m` (megabytes), `g` (gigabytes). Default unit is `m` (megabytes).
