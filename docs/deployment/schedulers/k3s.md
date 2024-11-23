# K3s Scheduler

> [!IMPORTANT]
> New as of 0.33.0

```
scheduler-k3s:annotations:set <app|--global> <property> (<value>) [--process-type PROCESS_TYPE] <--resource-type RESOURCE_TYPE>, Set or clear an annotation for a given app/process-type/resource-type combination
scheduler-k3s:autoscaling-auth:set <app|--global> <trigger> [<--metadata key=value>...], Set or clear a scheduler-k3s autoscaling keda trigger authentication resource for an app
scheduler-k3s:autoscaling-auth:report <app|--global> [--format stdout|json] [--include-metadata] # Displays a scheduler-k3s autoscaling auth report for an app
scheduler-k3s:cluster-add [ssh://user@host:port]    # Adds a server node to a Dokku-managed cluster
scheduler-k3s:cluster-list                          # Lists all nodes in a Dokku-managed cluster
scheduler-k3s:cluster-remove [node-id]              # Removes client node to a Dokku-managed cluster
scheduler-k3s:ensure-charts                         # Ensures the k3s charts are installed
scheduler-k3s:initialize                            # Initializes a cluster
scheduler-k3s:labels:set <app|--global> <property> (<value>) [--process-type PROCESS_TYPE] <--resource-type RESOURCE_TYPE>, Set or clear a label for a given app/process-type/resource-type combination
scheduler-k3s:report [<app>] [<flag>]               # Displays a scheduler-k3s report for one or more apps
scheduler-k3s:set [<app>|--global] <key> (<value>)  # Set or clear a scheduler-k3s property for an app or the scheduler
scheduler-k3s:show-kubeconfig                       # Displays the kubeconfig for remote usage
scheduler-k3s:uninstall                             # Uninstalls k3s from the Dokku server
```

> [!NOTE]
> The k3s plugin replaces the external [scheduler-kubernetes](https://github.com/dokku/dokku-scheduler-kubernetes) plugin. Users can continue to use the external plugin as necessary, but all future development will occur on the official core k3s plugin.

For multi-server support, Dokku provides the ability for users to setup a K3s cluster. As with all schedulers, it is set on a per-app basis. The scheduler can currently be overridden by running the following command:

```shell
dokku scheduler:set node-js-app selected k3s
```

As it is the default, unsetting the `selected` scheduler property is also a valid way to reset the scheduler.

```shell
dokku scheduler:set node-js-app k3s
```

## Usage

> [!IMPORTANT]
> The k3s plugin requires usage of a docker registry to store deployed image artifacts. See the [registry documentation](/docs/advanced-usage/registry-management.md) for more details on how to configure a registry.

### Initializing a cluster

> [!WARNING]
> This command must be run as root

Clusters can be initialized via the `scheduler-k3s:initialize` command. This will start a k3s cluster on the Dokku node itself.

```shell
dokku scheduler-k3s:initialize
```

By default, the k3s installation can run both app and system workloads. For clusters where app workloads are run on distinct worker nodes, initialize the cluster with the `--taint-scheduling` flag, which will allow _only_ Critical cluster components on the k3s control-plane nodes.

```shell
dokku scheduler-k3s:initialize --taint-scheduling
```

By default, Dokku will attempt to auto-detect the IP address of the server. In cases where the auto-detected IP address is incorrect, an override may be specified via the `--server-ip` flag:

```shell
dokku scheduler-k3s:initialize --server-ip 192.168.20.15
```

Dokku's k3s integration natively uses `nginx` as it's ingress load balancer via [ingress-nginx](https://github.com/kubernetes/ingress-nginx). Properties set by the `nginx` plugin will be respected, either by turning them into annotations or creating a custom server/location snippet that the `ingress-nginx` project can use. A `ps:restart` is required after changing nginx properties in order to have them apply to running resources.

Dokku can also use Traefik on cluster initialization via the [Traefik's CRDs](https://doc.traefik.io/traefik/providers/kubernetes-crd/). To change the ingress, set the `--ingress-class` flag:

```shell
dokku scheduler-k3s:initialize --ingress-class traefik
```

### Adding nodes to the cluster

> [!WARNING]
> The `dokku` user _must_ be able to ssh onto the server in order to connect nodes to the cluster. The remote user must be root or have sudo enabled, or the install will fail.

#### Adding a worker node

Nodes that run app workloads can be added via the `scheduler-k3s:cluster-add` command. This will ssh onto the specified server, install k3s, and join it to the current Dokku node in worker mode. Workers are typically used to run app workloads.

```shell
dokku scheduler-k3s:cluster-add  ssh://root@worker-1.example.com
```

If the server isn't in the `known_hosts` file, the connection will fail. This can be bypassed by setting the `--insecure-allow-unknown-hosts` flag:

```shell
dokku scheduler-k3s:cluster-add --insecure-allow-unknown-hosts ssh://root@worker-1.example.com
```

By default, Dokku will attempt to auto-detect the IP address of the Dokku server for the remote server to connect to. In cases where the auto-detected IP address is incorrect, an override may be specified via the `--server-ip` flag:

```shell
dokku scheduler-k3s:cluster-add --server-ip 192.168.20.15 ssh://root@worker-1.example.com
```

#### Adding a server node

> [!NOTE]
> Only the initial Dokku server will be properly configured for push deployment, and should be considered your git remote. Additional server nodes are for ensuring high-availability of the K3s etcd state. Ensure this server is properly backed up and restorable or deployments will not work.

Server nodes are typically used to replicate the cluster state, and it is recommended to have an odd number of nodes spread across several availability zones (datacenters in close proximity within a region). This allows for higher availability in the event of a cluster failure. Server nodes run control-plane services such as the traefik load balancer and the etcd backing store.

Server nodes can also be added with the `scheduler-k3s:cluster-add` command by specifying `--role server`. This will ssh onto the specified server, install k3s, and join it to the current Dokku node in server mode.

```shell
dokku scheduler-k3s:cluster-add  --role server ssh://root@server-1.example.com
```

Server nodes allow any workloads to be scheduled on them by default, in addition to the control-plane, etcd, and the scheduler itself. To avoid app workloads being scheduled on your control-plane, use the `--taint-scheduling` flag:

```shell
dokku scheduler-k3s:cluster-add --role server --taint-scheduling ssh://root@server-1.example.com
```

If the server isn't in the `known_hosts` file, the connection will fail. This can be bypassed by setting the `--insecure-allow-unknown-hosts` flag:

```shell
dokku scheduler-k3s:cluster-add --role server --insecure-allow-unknown-hosts ssh://root@server-1.example.com
```

By default, Dokku will attempt to auto-detect the IP address of the Dokku server for the remote server to connect to. In cases where the auto-detected IP address is incorrect, an override may be specified via the `--server-ip` flag:

```shell
dokku scheduler-k3s:cluster-add --role server --server-ip 192.168.20.15 ssh://root@server-1.example.com
```

#### Changing the network interface

When attaching an worker or server node, the K3s plugin will look at the IP associated with the `eth0` interface and use that to connect the new node to the cluster. To change this, set the `network-interface` property to the appropriate value.

```shell
dokku scheduler-k3s:set --global network-interface eth1
```

### Changing deploy timeouts

By default, app deploys will timeout after 300s. To customize this value, set the `deploy-timeout` property via `scheduler-k3s:set`:

```shell
dokku scheduler-k3s:set node-js-app deploy-timeout 60s
```

The default value may be set by passing an empty value for the option:

```shell
dokku scheduler-k3s:set node-js-app deploy-timeout
```

The `deploy-timeout` property can also be set globally. The global default is `300s`.

```shell
dokku scheduler-k3s:set --global deploy-timeout 60s
```

The default value may be set by passing an empty value for the option.

```shell
dokku scheduler-k3s:set --global deploy-timeout
```

### Customizing the namespace

By default, app deploys will run against the `default` Kubernetes namespace. To customize this value, set the `namespace` property via `scheduler-k3s:set`:

```shell
dokku scheduler-k3s:set node-js-app namespace lollipop
```

The default value may be set by passing an empty value for the option:

```shell
dokku scheduler-k3s:set node-js-app namespace
```

The `namespace` property can also be set globally. The global default is `default`.

```shell
dokku scheduler-k3s:set --global namespace 60s
```

The default value may be set by passing an empty value for the option.

```shell
dokku scheduler-k3s:set --global namespace
```

### Enabling rollback on failure

By default, app deploys do not rollback on failure. To enable this functionality, set the `rollback-on-failure` property via `scheduler-k3s:set`:

```shell
dokku scheduler-k3s:set node-js-app rollback-on-failure true
```

The default value may be set by passing an empty value for the option:

```shell
dokku scheduler-k3s:set node-js-app rollback-on-failure
```

The `rollback-on-failure` property can also be set globally. The global default is `false`.

```shell
dokku scheduler-k3s:set --global rollback-on-failure false
```

The default value may be set by passing an empty value for the option.

```shell
dokku scheduler-k3s:set --global rollback-on-failure
```

### Using image pull secrets

When authenticating against a registry via `registry:login`, the scheduler-k3s plugin will authenticate all servers in the cluster against the registry specified. If desired, an image pull secret can be used instead. To customize this value, set the `image-pull-secrets` property via `scheduler-k3s:set`:

```shell
dokku scheduler-k3s:set node-js-app image-pull-secrets lollipop
```

The default value may be set by passing an empty value for the option:

```shell
dokku scheduler-k3s:set node-js-app image-pull-secrets
```

The `image-pull-secrets` property can also be set globally. The global default is empty string, and k3s will use Dokku's locally configured `~/.docker/config.json` for any private registry pulls.

```shell
dokku scheduler-k3s:set --global image-pull-secrets 60s
```

The default value may be set by passing an empty value for the option.

```shell
dokku scheduler-k3s:set --global image-pull-secrets
```

### SSL Certificates

#### Enabling letsencrypt integration

By default, letsencrypt is disabled and https port mappings are ignored. To enable, set the `letsencrypt-email-prod` or `letsencrypt-email-stag` property with the `--global` flag:

```shell
# set the value for prod
dokku scheduler-k3s:set --global letsencrypt-email-prod automated@dokku.sh

# set the value for stag
dokku scheduler-k3s:set --global letsencrypt-email-stag automated@dokku.sh
```

After enabling and rebuilding, all apps with an `http:80` port mapping will have a corresponding `https:443` added and ssl will be automatically enabled. All http requests will then be redirected to https.

#### Customizing the letsencrypt server

The letsencrypt integration is set to the production letsencrypt server by default. This can be changed on an app-level by setting the `letsencrypt-server` property with the `scheduler-k3s:set` command

```shell
dokku scheduler-k3s:set node-js-app letsencrypt-server staging
```

The default value may be set by passing an empty value for the option:

```shell
dokku scheduler-k3s:set node-js-app letsencrypt-server
```

The `image-pull-secrets` property can also be set globally. The global default is `production`.

```shell
dokku scheduler-k3s:set --global letsencrypt-server staging
```

The default value may be set by passing an empty value for the option.

```shell
dokku scheduler-k3s:set --global letsencrypt-server staging
```

### Customizing Annotations and Labels

> [!NOTE]
> The cron ID is used as the process type if your app deploys any cron tasks

#### Setting Annotations

Dokku injects certain annotations into each created resource by default, but it may be necessary to inject others for tighter integration with third-party tools. The `scheduler-k3s:annotations:set` command can be used to perform this task. The command takes an app name and a required `--resource-type` flag.

```shell
dokku scheduler-k3s:annotations:set node-js-app annotation.key annotation.value --resource-type deployment
```

If not specified, the annotation will be applied to all processes within an app, though it may be further scoped to a specific process type via the `--process-type` flag.

```shell
dokku scheduler-k3s:annotations:set node-js-app annotation.key annotation.value --resource-type deployment --process-type web
```

The following resource types are supported:

- `certificate`
- `cronjob`
- `deployment`
- `ingress`
- `job`
- `pod`
- `secret`
- `service`
- `serviceaccount`
- `traefik_ingressroute`
- `traefik_middleware`

A `ps:restart` is required after setting annotations in order to have them apply to running resources.

#### Removing an annotation

To unset an annotation, pass an empty value:

```shell
dokku scheduler-k3s:annotations:set node-js-app annotation.key --resource-type deployment
dokku scheduler-k3s:annotations:set node-js-app annotation.key --resource-type deployment --process-type web
```

A `ps:restart` is required after removing annotations in order to remove them from running resources.

#### Setting Labels

Dokku injects certain labels into each created resource by default, but it may be necessary to inject others for tighter integration with third-party tools. The `scheduler-k3s:labels:set` command can be used to perform this task. The command takes an app name and a required `--resource-type` flag.

```shell
dokku scheduler-k3s:labels:set node-js-app label.key label.value --resource-type deployment
```

If not specified, the label will be applied to all processes within an app, though it may be further scoped to a specific process type via the `--process-type` flag.

```shell
dokku scheduler-k3s:labels:set node-js-app label.key label.value --resource-type deployment --process-type web
```

The following resource types are supported:

- `certificate`
- `cronjob`
- `deployment`
- `ingress`
- `job`
- `pod`
- `secret`
- `service`
- `serviceaccount`
- `traefik_ingressroute`
- `traefik_middleware`

A `ps:restart` is required after setting labels in order to have them apply to running resources.

#### Removing a label

To unset an label, pass an empty value:

```shell
dokku scheduler-k3s:annotations:set node-js-app label.key --resource-type deployment
dokku scheduler-k3s:labels:set node-js-app label.key --resource-type deployment --process-type web
```

A `ps:restart` is required after removing labels in order to remove them from running resources.

### Autoscaling

#### Workload Autoscaling

> [!IMPORTANT]
> New as of 0.33.8
> Users with older installations will need to manually install Keda.

Autoscaling in k3s is managed by [Keda](https://keda.sh/), which integrates with a variety of external metric providers to allow for autoscaling application workloads.

To enable autoscaling, use the `app.json` `formation.$PROCESS_TYPE.autoscaling` key to manage rules. In addition to the existing configuration used for process management, each process type in the `formation.$PROCESS_TYPE.autoscaling` key can have the following keys:

- `min_quantity`: The minimum number of instances the application can run. If not specified, the `quantity` specified for the app is used.
- `max_quantity`: The maximum number of instances the application can run. If not specified, the higher value of `quantity` and the `min_quantity` is used.
- `polling_interval_seconds`: (default: 30) The interval to wait for polling each of the configured triggers
- `cooldown_seconds`: (default: 300) The number of seconds to wait in between each scaling event
- `triggers`: A list of autoscaling triggers.

Autoscaling triggers are passed as is to Keda, and should match the configuration keda uses for a given [scaler](https://keda.sh/docs/2.13/scalers/). Below is an example for [datadog](https://keda.sh/docs/2.13/scalers/datadog/#example-2---driving-scale-directly):

```json
{
    "formation": {
        "web": {
            "autoscaling": {
                "min_quantity": 1,
                "max_quantity": 10,
                "triggers": [
                    {
                        "name": "name-for-trigger",
                        "type": "datadog",
                        "metadata": {
                            "query": "per_second(sum:http.requests{service:myservice1}).rollup(max, 300))/180,per_second(sum:http.backlog{service:myservice1}).rollup(max, 300)/30",
                            "queryValue": "1",
                            "queryAggregator": "max"
                        }
                    }
                ]
            }
        }
    }
}
```

Each value in the `metadata` stanza can use the following interpolated strings:

- `DOKKU_DEPLOYMENT_NAME`: The name of the deployment being scaled
- `DOKKU_PROCESS_TYPE`: The name of the process being scaled
- `DOKKU_APP_NAME`: The name of the app being scaled

##### HTTP Autoscaling

In addition to the built-in scalers that Keda provides, Dokku also supports Keda's HTTP Add On. This requires that the addon be properly installed and configured. For existing k3s clusters, this can be performed by the `scheduler-k3s:ensure-charts` command:

```shell
dokku scheduler-k3s:ensure-charts
```

> [!NOTE]
> Users who wish to use this functionality on a cluster not managed by Dokku will need to manually install the `keda-http-add-on` into the `keda` namespace. Please consult the `keda-http-add-on` [install documentation](https://kedacore.github.io/http-add-on/install.html) for further details.

> [!WARNING]
> If the `keda-http-add-on` chart is not installed, then this trigger will be ignored.

Once the chart is configured, an `http` trigger can be specified like so:


```json
{
    "formation": {
        "web": {
            "autoscaling": {
                "min_quantity": 1,
                "max_quantity": 10,
                "triggers": [
                    {
                        "type": "http",
                        "metadata": {
                            "scaledown_period_seconds": "150",
                            "request_rate_target_value": "50"
                        }
                    }
                ]
            }
        }
    }
}
```

The following metadata properties are supported with the http autoscaler:

- `scale_by`: (default: `request_rate`) whether to scale by `concurrency` or `request_rate`.
- `scaledown_period_seconds`: (default: `300`) period to wait after the last reported active before scaling the resource back to 0.
- `request_rate_granularity_seconds`: (default: `1`) granualarity of the aggregated requests for the request rate calculation.
- `request_rate_target_value`: (default: `100`) target value for the request rate.
- `request_rate_window_seconds`: (default: `60`) aggregation window for the request rate calculation.
- `concurrency_target_value`: (default: `100`) target value for the request concurrency.

Note that due to Keda limitations, scaling is done by _either_ `concurrency` or `request_rate`.

#### Workload Autoscaling Authentication

Most Keda triggers require some form of authentication to query for data. In the Kubernetes API, they are represented by `TriggerAuthentication` and `ClusterTriggerAuthentication` resources. Dokku can manage these via the `scheduler-k3s:autoscaling-auth` commands, and includes generated resources with each helm release generated by a deploy.

If no app-specific authentication is provided for a given trigger type, Dokku will fallback to any globally defined `ClusterTriggerAuthentication` resources. Autoscaling triggers within an app all share the same `TriggerAuthentication` resources, while `ClusterTriggerAuthentication` resources can be shared across all apps deployed by Dokku within a given cluster.

##### Creating Authentication Resources

Users can specify custom authentication resources directly via the Kubernetes api _or_ use the `scheduler-k3s:autoscaling-auth:set` command to create the resources in the Kubernetes cluster.

```shell
dokku scheduler-k3s:autoscaling-auth:set $APP $TRIGGER --metadata apiKey=some-api-key --metadata appKey=some-app-key
```

For example, the following will configure the authentication for all datadog triggers on the specified app:

```shell
dokku scheduler-k3s:autoscaling-auth:set node-js-app datadog --metadata apiKey=1234567890 --metadata appKey=asdfghjkl --metadata datadogSite=us5.datadoghq.com
```

After execution, Dokku will include the following resources for each specified trigger with the helm release generated on subsequent app deploys:

- `Secret`: an Opaque `Secret` resource storing the authentication credentials
- `TriggerAuthentication`: A `TriggerAuthentication` resource that references the secret for use by triggers

If the `--global` flag is specified instead of an app name, a custom helm chart is created on the fly with the above resources.

##### Removing Authentication Resources

To remove a configured authenticatin resource, run the `scheduler-k3s:autoscaling-auth:set` command with no metadata specified. Subsequent deploys will not include these resources.

```shell
dokku scheduler-k3s:autoscaling-auth:set $APP $TRIGGER_TYPE
```

##### Displaying an Authentication Resource report

To see a list of authentication resources managed by Dokku, run the `scheduler-k3s:autoscaling-auth:report` command.

```shell
dokku scheduler-k3s:autoscaling-auth:report node-js-app
```

```
====> $APP autoscaling-auth report
      datadog: configured
```

By default, the report will not display configured metadata - making it safe to include in Dokku report output. To include metadata and their values, add the `--include-metadata` flag:

```shell
dokku scheduler-k3s:autoscaling-auth:report node-js-app --include-metadata
```

```
====> node-js-app autoscaling-auth report
      Datadog:                       configured
      Datadog apiKey:                1234567890
      Datadog appKey:                asdfghjkl
      Datadog datadogSite:           us5.datadoghq.com
```

### Using kubectl remotely

> [!WARNING]
> Certain ports must be open for interacting with the remote kubernets api. Refer to the [K3s networking documentation](https://docs.k3s.io/installation/requirements?os=debian#networking) for the required open ports between servers prior to running the command.

By default, Dokku assumes that all it controls all actions on the cluster, and thus does not expose the `kubectl` binary for administrators. To interact with kubectl, you will need to retrieve the `kubeconfig` for the cluster and configure your client to use that configuration.

```shell
dokku scheduler-k3s:show-kubeconfig
```

### Interacting with an external Kubernetes cluster

While the k3s scheduler plugin is designed to work with a Dokku-managed k3s cluster, Dokku can be configured to interact with any Kubernetes cluster by setting the global `kubeconfig-path` to a path to a custom kubeconfig on the Dokku server. This property is only available at a global level.

```shell
dokku scheduler-k3s:set --global kubeconfig-path /path/to/custom/kubeconfig
```

To set the default value, omit the value from the `scheduler-k3s:set` call:

```shell
dokku scheduler-k3s:set --global kubeconfig-path
```

The default value for the `kubeconfig-path` is the k3s kubeconfig located at `/etc/rancher/k3s/k3s.yaml`.

### Customizing the Kubernetes context

When interacting with a custom Kubeconfig, the `kube-context` property can be set to specify a specific context within the kubeconfig to use. This property is available only at the global leve.

```shell
dokku scheduler-k3s:set --global kube-context lollipop
```

To set the default value, omit the value from the `scheduler-k3s:set` call:

```shell
dokku scheduler-k3s:set --global kube-context
```

The default value for the `kube-context` is an empty string, and will result in Dokku using the current context within the kubeconfig.

### Customizing Helm Chart Properties

Dokku includes a number of helm charts by default with settings that are optimized for Dokku. That said, it may be useful to further customize the charts for a given environment. Users can customize which charts are installed by setting properties prefixed with `chart.$CHART_NAME.` with the `--global` flag.

```shell
dokku scheduler-k3s:set --global chart.cert-manager.version 1.13.3
```

> [!NOTE]
> Properties follow dot-notation, and are expanded according to Helm's internal logic. See the [Helm documentation](https://helm.sh/docs/helm/helm_install/#helm-install) for `helm install` for further details.


To unset a chart property, omit the value from the `scheduler-k3s:set` call:

```shell
dokku scheduler-k3s:set --global chart.cert-manager.version
```

A `scheduler-k3s:ensure-charts` command with the `--force` flag is required after changing any chart properties in order to have them apply. This will install all charts, not just the ones that have changed.

```shell
dokku scheduler-k3s:ensure-charts --force
```

Alternatively, a comma separated list of chart names can be specified to only force install the specified charts:

```shell
dokku scheduler-k3s:ensure-charts --force --chart-names cert-manager
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
- `nginx`
    - Properties set by the `nginx` plugin will be respected, either by turning them into annotations or creating a custom server/location snippet that the `ingress-nginx` project can use. A `ps:restart` after changing any nginx properties is required in order to have them apply.
    - The `nginx:access-logs` and `nginx:error-logs` commands will fetch logs from one running `ingress-nginx` pod.
    - The `nginx:show-config` command will retrieve any `server` blocks associated with a domain attached to the app from one running `ingress-nginx` pod.
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

The `k3s` scheduler supports a minimal list of resource _limits_ and _reservations_:

- cpu: is specified in number of CPUs a process can access.
- memory: should be specified with a suffix of `b` (bytes), `Ki` (kilobytes), `Mi` (megabytes), `Gi` (gigabytes). Default unit is `Mi` (megabytes).

If unspecified for any task, the default reservation will be `.1` CPU and `128Mi` RAM, with no limit set for either CPU or RAM. This is to avoid issues with overscheduling pods on a cluster. To avoid issues, set more specific values for at least resource reservations. If unbounded utilization is desired, set CPU and Memory to `0m` and `0Mi`, respectively.

> [!NOTE]
> Cron tasks retrieve resource limits based on the computed cron task ID.
