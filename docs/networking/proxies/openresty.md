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
dokku proxy:set node-js-app openresty
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
dokku openresty:report node-js-app --openresty-letsencrypt-email
```
