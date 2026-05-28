# OpenResty Proxy

> [!IMPORTANT]
> New as of 0.31.0

Dokku can provide integration with the [OpenResty](https://openresty.org/) proxy service by utilizing the Docker label-based integration implemented by [openresty-docker-proxy](https://github.com/dokku/openresty-docker-proxy).

```
openresty:report [<app>] [<flag>]            # Displays a openresty report for one or more apps
openresty:logs [--num num] [--tail]          # Display openresty log output
openresty:set <app> <property> (<value>)     # Set or clear an openresty property for an app
openresty:show-config <app>                  # Display openresty compose config
openresty:start                              # Starts the openresty server
openresty:stop                               # Stops the openresty server
```

## Requirements

Using the `openresty` plugin integration requires the `docker-compose-plugin` for Docker. See [this document](https://docs.docker.com/compose/install/) from the Docker documentation for more information on the installation process for the `docker-compose-plugin`.

## Usage

> [!WARNING]
> As using multiple proxy plugins on a single Dokku installation can lead to issues routing requests to apps, doing so should be avoided. As the default proxy implementation is nginx, users are encouraged to stop the nginx service before switching to OpenResty.

The OpenResty plugin has specific rules for routing requests:

- OpenResty integration is exposed via docker labels attached to containers. Changes in labels require either app deploys or rebuilds.
- While OpenResty will respect labels associated with other containers, only `web` containers have OpenResty labels injected by the plugin.
- Only `http:80` and `https:443` port mappings are supported at this time.
- Requests are routed as soon as the container is running and passing healthchecks.

### Switching to OpenResty

To use the OpenResty plugin, use the `proxy:set` command for the app in question:

```shell
dokku proxy:set node-js-app type openresty
```

This will enable the docker label-based OpenResty integration. All future deploys will inject the correct labels for OpenResty to read and route requests to containers. Due to the docker label-based integration used by OpenResty, a single deploy or rebuild will be required before requests will route successfully.

```shell
dokku ps:rebuild node-js-app
```

Any changes to domains or port mappings will also require either a deploy or rebuild.

### Starting OpenResty container

OpenResty can be started via the `openresty:start` command. This will start a OpenResty container via the `docker compose up` command.

```shell
dokku openresty:start
```

### Stopping the OpenResty container

OpenResty may be stopped via the `openresty:stop` command.

```shell
dokku openresty:stop
```

The OpenResty container will be stopped and removed from the system. If the container is not running, this command will do nothing.

### Showing the OpenResty compose config

For debugging purposes, it may be useful to show the OpenResty compose config. This can be achieved via the `openresty:show-config` command.

```shell
dokku openresty:show-config
```

### Customizing the OpenResty container image

While the default OpenResty image is hardcoded, users may specify an alternative by setting the `image` property with the `--global` flag:

```shell
dokku openresty:set --global image dokku/openresty-docker-proxy:0.5.6
```

### Checking the OpenResty container's logs

It may be necessary to check the OpenResty container's logs to ensure that OpenResty is operating as expected. This can be performed with the `openresty:logs` command.

```shell
dokku openresty:logs
```

This command also supports the following modifiers:

```shell
--num NUM        # the number of lines to display
--tail           # continually stream logs
```

You can use these modifiers as follows:

```shell
dokku openresty:logs --tail --num 10
```

The above command will show logs continually from the openresty container, with an initial history of 10 log lines

### Customizing Openresty Settings for an app

#### OpenResty Properties

The OpenResty plugin supports all properties supported by the `nginx:set` command via `openresty:set`. At this time, please consult the nginx documentation for more information on what properties are available.

Please note that the oldest running container will be used for OpenResty configuration, and thus newer config may not apply until older app containers are retired during/after a deploy, depending on your zero-downtime settings.

#### Custom OpenResty Templates

At this time, the OpenResty plugin does not allow complete customization of the template used to manage an app's vhost. Apps will use a template provided by the OpenResty container to proxy requests. See the next section for documentation on how to configure portions of the template.

#### Injecting custom snippets into the OpenResty config

The OpenResty plugin allows users to specify templates in their repository for auto-injection into the OpenResty config. Please note that this configuration should be validated prior to deployment or may cause outages in your OpenResty proxy layer.

The following folders within an app repository may have `*.conf` files that will be automatically injected into the OpenResty config.

- `openresty/http-includes/`: Injected in the `server` block serving http(s) requests for the app.
- `openresty/http-location-includes/`: Injected in the `location` block that proxies to the app in the app's respective `server` block.

Custom snippets filenames may only include alphanumeric, underscore, and dot characters. For security reasons, filenames that contain other characters will be ignored.

### Label Management

The OpenResty plugin allows you to add custom container labels to apps. These labels are injected into containers during deployment and can be used to configure OpenResty behavior beyond what the plugin provides by default.

Refer to the upstream [openresty-docker-proxy](https://github.com/dokku/openresty-docker-proxy) documentation for more information on what labels are available.

#### Adding a label

To add a custom container label to an app, use the `openresty:labels:add` command:

```shell
dokku openresty:labels:add node-js-app openresty.directive value
```

This will add the label `openresty.directive=value` to the app's containers. After adding a label, you will need to rebuild or redeploy the app for the label to be applied to running containers.

```shell
dokku ps:rebuild node-js-app
```

#### Removing a label

To remove a custom container label from an app, use the `openresty:labels:remove` command:

```shell
dokku openresty:labels:remove node-js-app openresty.directive
```

This will remove the specified label from the app. After removing a label, you will need to rebuild or redeploy the app for the change to be applied to running containers.

```shell
dokku ps:rebuild node-js-app
```

#### Showing labels

To view all custom container labels for an app, use the `openresty:labels:show` command:

```shell
dokku openresty:labels:show node-js-app
```

To view a specific label value, provide the label name:

```shell
dokku openresty:labels:show node-js-app openresty.directive
```

### SSL Configuration

The OpenResty plugin only supports automatic ssl certificates from it's letsencrypt integration. Managed certificates provided by the `certs` plugin are ignored.

#### Enabling letsencrypt integration

By default, letsencrypt is disabled and https port mappings are ignored. To enable, set the `letsencrypt-email` property with the `--global` flag:

```shell
dokku openresty:set --global letsencrypt-email automated@dokku.sh
```

After enabling, the OpenResty container will need to be restarted and apps will need to be rebuilt. All http requests will then be redirected to https.

#### Customizing the letsencrypt server

The letsencrypt integration is set to the production letsencrypt server by default. To change this, set the `letsencrypt-server` property with the `--global` flag:

```shell
dokku openresty:set --global letsencrypt-server https://acme-staging-v02.api.letsencrypt.org/directory
```

After enabling, the OpenResty container will need to be restarted and apps will need to be rebuilt to retrieve certificates from the new server.

#### Limiting letsencrypt to certain domains

> [!WARNING]
> Changing this value may cause OpenResty to fail to start if the value is not valid. Caution should be exercised when changing this value from the defaults.

In cases where your server's IP may have invalid domains pointing at it, limiting letsencrypt to certain allowed domains may be desirable to reduce spam requests on the Letsencrypt servers. The default is to allow all domains to have certificates retrieved, but this can be limited by specifying the `allowed-letsencrypt-domains-func-base64` global property.

The default internal value for `allowed-letsencrypt-domains-func-base64` is the base64 representation of `return true`, and is meant to be the body of a lua function that return a boolean value.

```shell
value="$(echo 'return true' | base64 -w 0)"
dokku openresty:set --global allowed-letsencrypt-domains-func-base64 $value
```

As this is a global value, once changed, OpenResty should be stopped and started again for the value to take effect:

```shell
dokku openresty:stop
dokku openresty:start
```

A more complex example would be to limit provisioning of certificates to domains in a specific list. The body of the lua function has access to a variable `domain`, and we can use it like so:

```shell
body='allowed_domains = {"domain.com", "extra-domain.com"}

for index, value in ipairs(allowed_domains) do
  if value == domain then
    return true
  end
end

return false
'
value="$(echo "$body" | base64 -w 0)"
dokku openresty:set --global allowed-letsencrypt-domains-func-base64 $value
```

To reset the value to the default, simply specify a blank value prior to restarting OpenResty:

```shell
dokku openresty:set --global allowed-letsencrypt-domains-func-base64
```

## Displaying OpenResty reports for an app

You can get a report about the app's OpenResty config using the `openresty:report` command:

```shell
dokku openresty:report
```

```
=====> node-js-app openresty information
       Openresty image:                   dokku/openresty-docker-proxy:0.5.6
       Openresty letsencrypt email:       automated@dokku.sh
=====> python-app openresty information
       Openresty image:                   dokku/openresty-docker-proxy:0.5.6
       Openresty letsencrypt email:       automated@dokku.sh
=====> ruby-app openresty information
       Openresty image:                   dokku/openresty-docker-proxy:0.5.6
       Openresty letsencrypt email:       automated@dokku.sh
```

You can run the command for a specific app also.

```shell
dokku openresty:report node-js-app
```

```
=====> node-js-app openresty information
       Openresty image:                   dokku/openresty-docker-proxy:0.5.6
       Openresty letsencrypt email:       automated@dokku.sh
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku openresty:report node-js-app --openresty-computed-letsencrypt-email
```

## Properties

### Settable properties

Five properties (`image`, `log-level`, `letsencrypt-email`, `letsencrypt-server`, `allowed-letsencrypt-domains-func-base64`) are global only. The rest may be set per-app or globally with `--global`; a global value applies to any app that has no per-app value, otherwise the built-in default is used.

Global-only properties expose two report flags: `--openresty-global-<property>` returns the raw stored value (empty when the property has never been set), while `--openresty-computed-<property>` returns the effective value (the global value if set, otherwise the built-in default).

App-or-global properties expose three report flags: `--openresty-<property>` returns the raw per-app value (empty when unset), `--openresty-global-<property>` returns the raw global value (empty when unset), and `--openresty-computed-<property>` returns the effective value, resolving the per-app value first, then the global value, then the built-in default.

| Property | Scope | Default | Report flags | Description |
|---|---|---|---|---|
| `access-log-format` | app or global | none | `--openresty-access-log-format`, `--openresty-global-access-log-format`, `--openresty-computed-access-log-format` | Custom nginx `log_format` directive used for the access log |
| `access-log-path` | app or global | _`/var/log/nginx/{app}-access.log`_ | `--openresty-access-log-path`, `--openresty-global-access-log-path`, `--openresty-computed-access-log-path` | Path inside the openresty container where access logs are written |
| `allowed-letsencrypt-domains-func-base64` | global only | _allow-all stub_ | `--openresty-global-allowed-letsencrypt-domains-func-base64`, `--openresty-computed-allowed-letsencrypt-domains-func-base64` | Base64-encoded Lua function deciding which domains may request a letsencrypt certificate |
| `bind-address-ipv4` | app or global | none | `--openresty-bind-address-ipv4`, `--openresty-global-bind-address-ipv4`, `--openresty-computed-bind-address-ipv4` | IPv4 address the openresty server block binds to |
| `bind-address-ipv6` | app or global | `::` | `--openresty-bind-address-ipv6`, `--openresty-global-bind-address-ipv6`, `--openresty-computed-bind-address-ipv6` | IPv6 address the openresty server block binds to |
| `client-body-timeout` | app or global | `60s` | `--openresty-client-body-timeout`, `--openresty-global-client-body-timeout`, `--openresty-computed-client-body-timeout` | Time allowed to read the request body from the client |
| `client-header-timeout` | app or global | `60s` | `--openresty-client-header-timeout`, `--openresty-global-client-header-timeout`, `--openresty-computed-client-header-timeout` | Time allowed to read the request header from the client |
| `client-max-body-size` | app or global | `1m` | `--openresty-client-max-body-size`, `--openresty-global-client-max-body-size`, `--openresty-computed-client-max-body-size` | Maximum allowed request body size |
| `error-log-path` | app or global | _`/var/log/nginx/{app}-error.log`_ | `--openresty-error-log-path`, `--openresty-global-error-log-path`, `--openresty-computed-error-log-path` | Path inside the openresty container where error logs are written |
| `hsts` | app or global | `true` | `--openresty-hsts`, `--openresty-global-hsts`, `--openresty-computed-hsts` | When `true`, emits a `Strict-Transport-Security` header on HTTPS responses |
| `hsts-include-subdomains` | app or global | `true` | `--openresty-hsts-include-subdomains`, `--openresty-global-hsts-include-subdomains`, `--openresty-computed-hsts-include-subdomains` | Adds the `includeSubDomains` directive to the HSTS header |
| `hsts-max-age` | app or global | `15724800` | `--openresty-hsts-max-age`, `--openresty-global-hsts-max-age`, `--openresty-computed-hsts-max-age` | `max-age` value (seconds) in the HSTS header |
| `hsts-preload` | app or global | `false` | `--openresty-hsts-preload`, `--openresty-global-hsts-preload`, `--openresty-computed-hsts-preload` | Adds the `preload` directive to the HSTS header |
| `image` | global only | _parsed from `plugins/openresty-vhosts/Dockerfile`_ | `--openresty-global-image`, `--openresty-computed-image` | Docker image used to run the openresty container |
| `keepalive-timeout` | app or global | `75s` | `--openresty-keepalive-timeout`, `--openresty-global-keepalive-timeout`, `--openresty-computed-keepalive-timeout` | Time an idle keep-alive connection stays open |
| `letsencrypt-email` | global only | none | `--openresty-global-letsencrypt-email`, `--openresty-computed-letsencrypt-email` | Contact email enabling letsencrypt; empty disables https issuance |
| `letsencrypt-server` | global only | `https://acme-v02.api.letsencrypt.org/directory` | `--openresty-global-letsencrypt-server`, `--openresty-computed-letsencrypt-server` | ACME directory used when requesting certificates |
| `lingering-timeout` | app or global | `5s` | `--openresty-lingering-timeout`, `--openresty-global-lingering-timeout`, `--openresty-computed-lingering-timeout` | Time openresty waits for more client data when closing a connection |
| `log-level` | global only | `ERROR` | `--openresty-global-log-level`, `--openresty-computed-log-level` | Openresty log level |
| `proxy-buffer-size` | app or global | _system pagesize_ | `--openresty-proxy-buffer-size`, `--openresty-global-proxy-buffer-size`, `--openresty-computed-proxy-buffer-size` | Buffer size for reading the first part of the upstream response |
| `proxy-buffering` | app or global | `on` | `--openresty-proxy-buffering`, `--openresty-global-proxy-buffering`, `--openresty-computed-proxy-buffering` | Whether openresty buffers upstream responses (`on` or `off`) |
| `proxy-buffers` | app or global | _`8 {pagesize}`_ | `--openresty-proxy-buffers`, `--openresty-global-proxy-buffers`, `--openresty-computed-proxy-buffers` | Number and size of buffers used for an upstream response |
| `proxy-busy-buffers-size` | app or global | _`2 * pagesize`_ | `--openresty-proxy-busy-buffers-size`, `--openresty-global-proxy-busy-buffers-size`, `--openresty-computed-proxy-busy-buffers-size` | Maximum buffer size that can be busy sending a response to the client |
| `proxy-connect-timeout` | app or global | `60s` | `--openresty-proxy-connect-timeout`, `--openresty-global-proxy-connect-timeout`, `--openresty-computed-proxy-connect-timeout` | Time to establish a connection to the upstream |
| `proxy-read-timeout` | app or global | `60s` | `--openresty-proxy-read-timeout`, `--openresty-global-proxy-read-timeout`, `--openresty-computed-proxy-read-timeout` | Time to read a response from the upstream |
| `proxy-send-timeout` | app or global | `60s` | `--openresty-proxy-send-timeout`, `--openresty-global-proxy-send-timeout`, `--openresty-computed-proxy-send-timeout` | Time to transmit a request to the upstream |
| `send-timeout` | app or global | `60s` | `--openresty-send-timeout`, `--openresty-global-send-timeout`, `--openresty-computed-send-timeout` | Time between two successive write operations to the client |
| `underscore-in-headers` | app or global | `off` | `--openresty-underscore-in-headers`, `--openresty-global-underscore-in-headers`, `--openresty-computed-underscore-in-headers` | Whether to allow underscores in client request header field names |
| `x-forwarded-for-value` | app or global | `$remote_addr` | `--openresty-x-forwarded-for-value`, `--openresty-global-x-forwarded-for-value`, `--openresty-computed-x-forwarded-for-value` | Value used for the `X-Forwarded-For` header |
| `x-forwarded-port-value` | app or global | `$server_port` | `--openresty-x-forwarded-port-value`, `--openresty-global-x-forwarded-port-value`, `--openresty-computed-x-forwarded-port-value` | Value used for the `X-Forwarded-Port` header |
| `x-forwarded-proto-value` | app or global | `$scheme` | `--openresty-x-forwarded-proto-value`, `--openresty-global-x-forwarded-proto-value`, `--openresty-computed-x-forwarded-proto-value` | Value used for the `X-Forwarded-Proto` header |
| `x-forwarded-ssl` | app or global | none | `--openresty-x-forwarded-ssl`, `--openresty-global-x-forwarded-ssl`, `--openresty-computed-x-forwarded-ssl` | Value used for the `X-Forwarded-Ssl` header (e.g. `on`/`off`) |
