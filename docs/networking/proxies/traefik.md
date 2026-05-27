# Traefik Proxy

> [!IMPORTANT]
> New as of 0.28.0

Dokku provides integration with the [Traefik](https://traefik.io/) proxy service by utilizing the Docker label-based integration implemented by Traefik.

```
traefik:report [<app>] [<flag>]          # Displays a traefik report for one or more apps
traefik:logs [--num num] [--tail]        # Display traefik log output
traefik:set <app> <property> (<value>)   # Set or clear an traefik property for an app
traefik:show-config <app>                # Display traefik compose config
traefik:start                            # Starts the traefik server
traefik:stop                             # Stops the traefik server
```

## Requirements

Using the `traefik` plugin integration requires the `docker-compose-plugin` for Docker. See [this document](https://docs.docker.com/compose/install/) from the Docker documentation for more information on the installation process for the `docker-compose-plugin`.

## Usage

> [!WARNING]
> As using multiple proxy plugins on a single Dokku installation can lead to issues routing requests to apps, doing so should be avoided. As the default proxy implementation is nginx, users are encouraged to stop the nginx service before switching to Traefik.

The Traefik plugin has specific rules for routing requests:

- Traefik integration is exposed via docker labels attached to containers. Changes in labels require either app deploys or rebuilds.
- While Traefik will respect labels associated with other containers, only `web` containers have Traefik labels injected by the plugin.
- Only `http:80` and `https:443` port mappings are supported.
- If no `http:80` mapping is found, the first `http` port mapping is used for http requests.
- If no `https:443` mapping is found, the first `https` port mapping is used for https requests.
- If no `https` mapping is found, the container port from `http:80` will be used for https requests.
- Requests are routed as soon as the container is running and passing healthchecks.
- Readiness healthchecks defined in `app.json` with a `path` property are automatically transformed into Traefik healthcheck labels.

### Healthchecks

When an app has a readiness healthcheck defined in its `app.json` file with a `path` property, Dokku automatically generates Traefik healthcheck labels. These labels configure Traefik to perform health checks on the container before routing traffic to it.

The following `app.json` healthcheck properties are mapped to Traefik labels:

| app.json Property | Traefik Label Property | Description |
|-------------------|------------------------|-------------|
| `path`            | `healthcheck.path`     | The HTTP path to check (required) |
| `scheme`          | `healthcheck.scheme`   | The scheme to use (`http` or `https`) |
| `port`            | `healthcheck.port`     | The port to check |
| `timeout`         | `healthcheck.timeout`  | Timeout in seconds (formatted as `Xs`) |
| `wait`            | `healthcheck.interval` | Interval between checks in seconds (formatted as `Xs`) |

Example `app.json` configuration:

```json
{
  "healthchecks": {
    "web": [
      {
        "name": "web readiness check",
        "path": "/health",
        "timeout": 5,
        "type": "readiness",
        "wait": 10
      }
    ]
  }
}
```

> [!NOTE]
> Only the first readiness healthcheck with a `path` property is used. Liveness, startup, and command-based healthchecks are not transformed into Traefik labels.

### Switching to Traefik

To use the Traefik plugin, use the `proxy:set` command for the app in question:

```shell
dokku proxy:set node-js-app type traefik
```

This will enable the docker label-based Traefik integration. All future deploys will inject the correct labels for Traefik to read and route requests to containers. Due to the docker label-based integration used by Traefik, a single deploy or rebuild will be required before requests will route successfully.

```shell
dokku ps:rebuild node-js-app
```

Any changes to domains or port mappings will also require either a deploy or rebuild.

### Starting Traefik container

Traefik can be started via the `traefik:start` command. This will start a Traefik container via the `docker compose up` command.

```shell
dokku traefik:start
```

### Stopping the Traefik container

Traefik may be stopped via the `traefik:stop` command.

```shell
dokku traefik:stop
```

The Traefik container will be stopped and removed from the system. If the container is not running, this command will do nothing.

### Changing the Traefik entrypoint names

When you use a self-hosted Traefik instance, your entrypoint names might be different from the default `http` and `https`

Use `traefik:set` to set both `http-entry-point` and `https-entry-point` to custom values

```shell
dokku traefik:set --global http-entry-point web
dokku traefik:set --global https-entry-point websecure
```

### Showing the Traefik compose config

For debugging purposes, it may be useful to show the Traefik compose config. This can be achieved via the `traefik:show-config` command.

```shell
dokku traefik:show-config
```

### Customizing the Traefik container image

While the default Traefik image is hardcoded, users may specify an alternative by setting the `image` property with the `--global` flag:

```shell
dokku traefik:set --global image traefik:v2.8
```

#### Checking the Traefik container's logs

It may be necessary to check the Traefik container's logs to ensure that Traefik is operating as expected. This can be performed with the `traefik:logs` command.

```shell
dokku traefik:logs
```

This command also supports the following modifiers:

```shell
--num NUM        # the number of lines to display
--tail           # continually stream logs
```

You can use these modifiers as follows:

```shell
dokku traefik:logs --tail --num 10
```

The above command will show logs continually from the traefik container, with an initial history of 10 log lines

### Changing the Traefik log level

Traefik log output is set to `ERROR` by default. It may be changed by setting the `log-level` property with the `--global` flag:

```shell
dokku traefik:set --global log-level DEBUG
```

After modifying, the Traefik container will need to be restarted.

### Label Management

The Traefik plugin allows you to add custom container labels to apps. These labels are injected into containers during deployment and can be used to configure Traefik behavior beyond what the plugin provides by default.

Refer to the upstream [Traefik](https://doc.traefik.io/traefik/) documentation for more information on what labels are available.

#### Adding a label

To add a custom container label to an app, use the `traefik:labels:add` command:

```shell
dokku traefik:labels:add node-js-app traefik.directive value
```

This will add the label `traefik.directive=value` to the app's containers. After adding a label, you will need to rebuild or redeploy the app for the label to be applied to running containers.

```shell
dokku ps:rebuild node-js-app
```

#### Removing a label

To remove a custom container label from an app, use the `traefik:labels:remove` command:

```shell
dokku traefik:labels:remove node-js-app traefik.directive
```

This will remove the specified label from the app. After removing a label, you will need to rebuild or redeploy the app for the change to be applied to running containers.

```shell
dokku ps:rebuild node-js-app
```

#### Showing labels

To view all custom container labels for an app, use the `traefik:labels:show` command:

```shell
dokku traefik:labels:show node-js-app
```

To view a specific label value, provide the label name:

```shell
dokku traefik:labels:show node-js-app traefik.directive
```

### SSL Configuration

The traefik plugin only supports automatic ssl certificates from it's letsencrypt integration. Managed certificates provided by the `certs` plugin are ignored.

#### Enabling letsencrypt integration

By default, letsencrypt is disabled and https port mappings are ignored. To enable, set the `letsencrypt-email` property with the `--global` flag:

```shell
dokku traefik:set --global letsencrypt-email automated@dokku.sh
```

After enabling, apps will need to be rebuilt and the Traefik container will need to be restarted. All http requests will then be redirected to https.

#### Customizing the letsencrypt server

The letsencrypt integration is set to the production letsencrypt server by default. To change this, set the `letsencrypt-server` property with the `--global` flag:

```shell
dokku traefik:set --global letsencrypt-server https://acme-staging-v02.api.letsencrypt.org/directory
```

After enabling, the Traefik container will need to be restarted and apps will need to be rebuilt to retrieve certificates from the new server.

#### Switching to DNS-01 challenge mode

By default, Traefik uses TLS-ALPN-01 challenge for obtaining certificates. To switch to DNS-01 challenge mode (useful for wildcard certificates or when port 443 is not accessible), you need to:

1. Set the challenge mode to `dns`:

```shell
dokku traefik:set --global challenge-mode dns
```

2. Set your DNS provider:

```shell
dokku traefik:set --global dns-provider cloudflare
```

3. Configure the required environment variables for your DNS provider. Each DNS provider requires specific environment variables. The variable names should be prefixed with `dns-provider-`:

```shell
dokku traefik:set --global dns-provider-cf_api_email user@example.com
dokku traefik:set --global dns-provider-cf_api_key your-api-key
```

The `dns-provider-` prefix will be stripped and the variable name will be uppercased when passed to the Traefik container. For example, `dns-provider-cf_api_email` becomes `CF_API_EMAIL`.

After configuring, the Traefik container will need to be restarted and apps will need to be rebuilt.

Refer to the [Traefik DNS Challenge documentation](https://doc.traefik.io/traefik/https/acme/#dnschallenge) for the list of supported DNS providers and their required environment variables.

To switch back to TLS challenge mode:

```shell
dokku traefik:set --global challenge-mode tls
```

### API Access

Traefik exposes an API and Dashboard, which Dokku disables by default for security reasons. It can be exposed and customized as described below.

#### Enabling the api

> [!WARNING]
> Users enabling the dashboard should also enable api basic auth.

By default, the api is disabled. To enable, set the `api-enabled` property with the `--global` flag:

```shell
dokku traefik:set --global api-enabled true
```

After enabling, the Traefik container will need to be restarted.

#### Enabling the dashboard

> [!WARNING]
> Users enabling the dashboard should also enable api basic auth.

By default, the dashboard is disabled. To enable, set the `dashboard-enabled` property with the `--global` flag:

```shell
dokku traefik:set --global dashboard-enabled true
```

After enabling, the Traefik container will need to be restarted.

#### Enabling api basic auth

Users enabling either the api or dashboard are encouraged to enable basic auth. This will apply _only_ to the api/dashboard, and not to apps. To enable, set the `basic-auth-username` and `basic-auth-password` properties with the `--global` flag:. Both must be set or basic auth will not be enabled.

```shell
dokku traefik:set --global basic-auth-username username
dokku traefik:set --global basic-auth-password password
```

After enabling, the Traefik container will need to be restarted.

#### Customizing the api hostname

The hostname used for the api and dashboard is set to `traefik.dokku.me` by default. It can be customized by setting the `api-vhost` property with the `--global` flag:

```shell
dokku traefik:set --global api-vhost lb.dokku.me
```

After enabling, the Traefik container will need to be restarted.

## Displaying Traefik reports for an app

You can get a report about the app's Traefik config using the `traefik:report` command:

```shell
dokku traefik:report
```

```
=====> node-js-app traefik information
       Traefik computed api enabled:      false
       Traefik computed api vhost:        traefik.dokku.me
       Traefik computed challenge mode:   tls
       Traefik computed dashboard enabled: false
       Traefik computed http entry point: http
       Traefik computed https entry point: https
       Traefik computed image:            traefik:v2.8
       Traefik computed letsencrypt email:
       Traefik computed letsencrypt server: https://acme-v02.api.letsencrypt.org/directory
       Traefik computed log level:        ERROR
       Traefik global api enabled:
       Traefik global api vhost:
       Traefik global challenge mode:
       Traefik global dashboard enabled:
       Traefik global http entry point:
       Traefik global https entry point:
       Traefik global image:
       Traefik global letsencrypt email:
       Traefik global letsencrypt server:
       Traefik global log level:
```

The `global-<prop>` keys hold the raw global value and are empty when nothing has been set globally. The `computed-<prop>` keys hold the effective value used at deploy time, falling back to the built-in default when the global value is empty.

You can run the command for a specific app also.

```shell
dokku traefik:report node-js-app
```

```
=====> node-js-app traefik information
       Traefik computed api enabled:      false
       Traefik computed api vhost:        traefik.dokku.me
       Traefik computed challenge mode:   tls
       Traefik computed dashboard enabled: false
       Traefik computed http entry point: http
       Traefik computed https entry point: https
       Traefik computed image:            traefik:v2.8
       Traefik computed letsencrypt email:
       Traefik computed letsencrypt server: https://acme-v02.api.letsencrypt.org/directory
       Traefik computed log level:        ERROR
       Traefik global api enabled:
       Traefik global api vhost:
       Traefik global challenge mode:
       Traefik global dashboard enabled:
       Traefik global http entry point:
       Traefik global https entry point:
       Traefik global image:
       Traefik global letsencrypt email:
       Traefik global letsencrypt server:
       Traefik global log level:
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku traefik:report node-js-app --traefik-computed-api-enabled
```

## Properties

### Settable properties

All traefik properties are global only. Set with `traefik:set --global <property> <value>`.

| Property | Scope | Default | Report flags | Description |
|---|---|---|---|---|
| `api-enabled` | global only | `false` | `--traefik-global-api-enabled`, `--traefik-computed-api-enabled` | When `true`, enables the Traefik HTTP API |
| `api-entry-point` | global only | none | `--traefik-global-api-entry-point`, `--traefik-computed-api-entry-point` | Name of the entry point used by the Traefik API |
| `api-entry-point-address` | global only | none | `--traefik-global-api-entry-point-address`, `--traefik-computed-api-entry-point-address` | Address (`host:port`) the Traefik API listens on |
| `api-vhost` | global only | `traefik.dokku.me` | `--traefik-global-api-vhost`, `--traefik-computed-api-vhost` | Virtual host that routes to the Traefik API |
| `basic-auth-password` | global only | none | `--traefik-global-basic-auth-password`, `--traefik-computed-basic-auth-password` | Password for basic auth in front of the API/dashboard |
| `basic-auth-username` | global only | none | `--traefik-global-basic-auth-username`, `--traefik-computed-basic-auth-username` | Username for basic auth in front of the API/dashboard |
| `challenge-mode` | global only | `tls` | `--traefik-global-challenge-mode`, `--traefik-computed-challenge-mode` | ACME challenge method used by Traefik (`tls`, `http`, or `dns`) |
| `dashboard-enabled` | global only | `false` | `--traefik-global-dashboard-enabled`, `--traefik-computed-dashboard-enabled` | When `true`, enables the Traefik dashboard |
| `dns-provider` | global only | none | `--traefik-global-dns-provider`, `--traefik-computed-dns-provider` | Lego DNS provider name used when `challenge-mode` is `dns` |
| `dns-provider-<ENV_VAR>` | global only | none | `--traefik-dns-provider-<env_var>` (masked as `*******` in the default stdout report; the raw value is returned when queried via `--format json` or when this flag is requested explicitly) | Per-provider environment variables passed to the Traefik container; `<ENV_VAR>` is the upstream variable name (e.g. `dns-provider-cloudflare-api-token`) |
| `http-entry-point` | global only | `http` | `--traefik-global-http-entry-point`, `--traefik-computed-http-entry-point` | Entry point name handling plaintext HTTP traffic |
| `https-entry-point` | global only | `https` | `--traefik-global-https-entry-point`, `--traefik-computed-https-entry-point` | Entry point name handling TLS-terminated HTTPS traffic |
| `image` | global only | _parsed from `plugins/traefik-vhosts/Dockerfile`_ | `--traefik-global-image`, `--traefik-computed-image` | Docker image used to run the Traefik container |
| `letsencrypt-email` | global only | none | `--traefik-global-letsencrypt-email`, `--traefik-computed-letsencrypt-email` | Contact email enabling letsencrypt; empty disables https issuance |
| `letsencrypt-server` | global only | `https://acme-v02.api.letsencrypt.org/directory` | `--traefik-global-letsencrypt-server`, `--traefik-computed-letsencrypt-server` | ACME directory used when requesting certificates |
| `log-level` | global only | `ERROR` | `--traefik-global-log-level`, `--traefik-computed-log-level` | Traefik log level |

### Internal properties

The following properties are not managed by `traefik:set` but are recorded internally by the plugin:

| Property | Description | Source |
|---|---|---|
| `proxy-status` | `started`/`stopped` state of the traefik compose project | `cmd-traefik-start`/`cmd-traefik-stop` in `plugins/traefik-vhosts/command-functions` |
