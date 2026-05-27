# Proxy Management

> [!IMPORTANT]
> New as of 0.5.0, Enhanced in 0.6.0

```
proxy:build-config [--parallel count] [--all|<app>] # (Re)builds config for given app
proxy:clear-config [--all|<app>] # Clears config for given app
proxy:disable [--parallel count] [--all|<app>]      # Disable proxy for app
proxy:enable [--parallel count] [--all|<app>]       # Enable proxy for app
proxy:report [<app>] [<flag>]                       # Displays a proxy report for one or more apps
proxy:set [<app>|--global] <key> (<value>)          # Set or clear a proxy property for an app
```

In Dokku 0.5.0, port proxying was decoupled from the `nginx-vhosts` plugin into the proxy plugin. Dokku 0.6.0 introduced the ability to map host ports to specific container ports. This allows other proxy software - such as HAProxy or Caddy - to be used in place of nginx.

## Usage

### Changing the proxy

The default proxy shipped with Dokku is `nginx`. It can be changed via the `proxy:set` command.

```shell
dokku proxy:set node-js-app type caddy
```

```
=====> Setting type to caddy
```

The proxy may also be set on a global basis. This is usually preferred as running multiple proxy implementations may cause port collision issues.

```shell
dokku proxy:set --global type caddy
```

```
=====> Setting type to caddy
```

Changing the proxy does not stop or start any given proxy implementation. Please see the documentation for your proxy implementation for details on how to perform a change.

### Disabling and enabling the proxy

The proxy can be disabled on a per-app basis via the `proxy:disable` command. While disabled, no proxy configuration will be generated for the app and incoming requests will not be routed through the configured proxy.

```shell
dokku proxy:disable node-js-app
```

```
-----> Disabling proxy for app
```

The proxy can be re-enabled with the `proxy:enable` command:

```shell
dokku proxy:enable node-js-app
```

```
-----> Enabling proxy for app
```

### Setting proxy ports

When an app's domains are disabled (see [domains documentation](/docs/configuration/domains.md)), Dokku still exposes the app on a high port so that internal services can reach it at a stable port across deploys. The non-SSL and SSL ports used for this can be configured via the `proxy:set` command:

```shell
dokku proxy:set node-js-app proxy-port 5000
dokku proxy:set node-js-app proxy-ssl-port 5443
```

Either property may also be set globally. Per-app values take precedence over the global value.

```shell
dokku proxy:set --global proxy-port 5000
dokku proxy:set --global proxy-ssl-port 5443
```

The default value for either property may be restored by passing an empty value:

```shell
dokku proxy:set node-js-app proxy-port
```

### Regenerating proxy config

In certain cases, your app proxy configs may drift from the correct config for your app. You may regenerate the config at any point via the `proxy:build-config` command. This command will trigger a rebuild for the configured proxy implementation (default: nginx) for a given app. The command may fail if there are no current web listeners for your app.

```shell
dokku proxy:build-config node-js-app
```

All apps may have their proxy config rebuilt by using the `--all` flag.

```shell
dokku proxy:build-config --all
```

By default, rebuilding proxy configs for all apps happens serially. The parallelism may be controlled by the `--parallel` flag.

```shell
dokku proxy:build-config --all --parallel 2
```

Finally, the number of parallel workers may be automatically set to the number of CPUs available by setting the `--parallel` flag to `-1`

```shell
dokku proxy:build-config --all --parallel -1
```

### Clearing the generated proxy config

> [!IMPORTANT]
> New as of 0.27.0

Generated proxy configurations can also be cleared using the `proxy:clear-config` command.

```shell
dokku proxy:clear-config node-js-app
```

All apps may have their proxy config cleared by using the `--all` flag.

```shell
dokku proxy:clear-config --all
```

Clearing a proxy configuration has different effects depending on the proxy plugin in use. Consul the documentation for your proxy implementation for further details.

### Displaying proxy reports for an app

> [!IMPORTANT]
> New as of 0.8.1

You can get a report about the app's proxy status using the `proxy:report` command:

```shell
dokku proxy:report
```

```
=====> node-js-app proxy information
       Proxy computed type: nginx
       Proxy enabled:       true
       Proxy global type:
       Proxy type:
=====> python-sample proxy information
       Proxy computed type: nginx
       Proxy enabled:       true
       Proxy global type:
       Proxy type:
=====> ruby-sample proxy information
       Proxy computed type: nginx
       Proxy enabled:       true
       Proxy global type:
       Proxy type:
```

The `type` and `global-type` keys hold the raw per-app and global value respectively, and are empty when nothing has been set. The `computed-type` key holds the effective value used at deploy time, falling back to the global value (where one has been set) and then to the built-in default of `nginx`.

You can run the command for a specific app also.

```shell
dokku proxy:report node-js-app
```

```
=====> node-js-app proxy information
       Proxy computed type: nginx
       Proxy enabled:       true
       Proxy global type:
       Proxy type:
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku proxy:report node-js-app --proxy-computed-type
```

#### Proxy Port Scheme

See the [port scheme documentation](/docs/networking/port-management.md#port-scheme) for more information on the port mapping scheme used by dokku.

### Proxy port mapping

See the [port management documentation](/docs/networking/port-management.md) for more information on how port mappings are managed for an application.

### Container network interface binding

> Changed as of 0.11.0

From Dokku versions `0.5.0` until `0.11.0`, enabling or disabling an application's proxy would **also** control whether or not the application was bound to all interfaces - e.g. `0.0.0.0`. As of `0.11.0`, this is now controlled by the network plugin. Please see the [network documentation](/docs/networking/network.md#container-network-interface-binding) for more information.

## Implementing a Proxy

Custom plugins names _must_ have the suffix `-vhosts` or scheduler overriding via `proxy:set` may not function as expected.

At this time, the following dokku commands are used to interact with a complete proxy implementation.

- `domains:add`: Adds a given domain to an app.
    - triggers: `post-domains-update`
- `domains:clear`: Clears out an app's associated domains.
    - triggers: `post-domains-update`
- `domains:disable`: Disables domains for an app.
    - triggers: `pre-disable-vhost`
- `domains:enable`: Enables domains for an app.
    - triggers: `pre-enable-vhost`
- `domains:remove`: Removes a domain from an app.
    - triggers: `post-domains-update`
- `domains:reset`: Reset app domains to global-configured domains.
    - triggers: `post-domains-update`
- `domains:set`: Sets all domains for a given app.
    - triggers: `post-domains-update`
- `proxy:build-config`: Builds - or rebuilds - external proxy configuration.
    - triggers: `proxy-build-config`
- `proxy:clear-config`: Clears out external proxy configuration.
    - triggers: `proxy-clear-config`
- `proxy:disable`: Disables the proxy configuration for an app.
    - triggers: `proxy-disable`
- `proxy:enable`: Enables the proxy configuration for an app.
    - triggers: `proxy-enable`
- `ports:add`: Adds one or more port mappings to an app
    - triggers: `post-proxy-ports-update`
- `ports:clear`: Clears out all port mappings for an app.
    - triggers: `post-proxy-ports-update`
- `ports:remove`: Removes one or more port mappings from an app.
    - triggers: `post-proxy-ports-update`
- `ports:set`: Sets all port mappings for an app.
    - triggers: `post-proxy-ports-update`

Proxy implementations may decide to omit some functionality here, or use plugin triggers to supplement config with information from other plugins.

Individual proxy implementations _may_ trigger app rebuilds, depending on how proxy metadata is exposed for the proxy implementation.

Finally, proxy implementations _may_ install extra software needed for the proxy itself in whatever manner deemed fit. Proxy software can run on the host itself or within a running Docker container with either exposed ports or host networking.

## Properties

### Settable properties

| Property | Scope | Default | Report flags | Description |
|---|---|---|---|---|
| `disabled` | app only | `false` | `--proxy-disabled`, `--proxy-computed-disabled` (also exposed inverted as `--proxy-enabled`) | When `true`, disables proxy integration for this app (`proxy:enable`/`proxy:disable` write this) |
| `proxy-port` | app + global | none | `--proxy-proxy-port`, `--proxy-global-proxy-port`, `--proxy-computed-proxy-port` | Override port used for the HTTP listener in the generated proxy config |
| `proxy-ssl-port` | app + global | none | `--proxy-proxy-ssl-port`, `--proxy-global-proxy-ssl-port`, `--proxy-computed-proxy-ssl-port` | Override port used for the HTTPS listener in the generated proxy config |
| `type` | app + global | `nginx` | `--proxy-type`, `--proxy-global-type`, `--proxy-computed-type` | Proxy implementation handling traffic for the app (`nginx`, `caddy`, `haproxy`, `traefik`, `openresty`, or a custom plugin) |

### Read-only flags

The following flags surface in `proxy:report` but are not managed by `proxy:set`:

| Flag | Description |
|---|---|
| `--proxy-enabled` | `true` when the app's `disabled` property is not `true` |
