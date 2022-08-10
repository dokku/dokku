# Caddy Proxy

Dokku provides integration with the [Caddy](https://caddyserver.com/) proxy service by utilizing the Docker label-based integration implemented by Caddy.

```
caddy:report [<app>] [<flag>]            # Displays a caddy report for one or more apps
caddy:logs [--num num] [--tail]          # Display caddy log output
caddy:set <app> <property> (<value>)     # Set or clear an caddy property for an app
caddy:show-config <app>                  # Display caddy compose config
caddy:start                              # Starts the caddy server
caddy:stop                               # Stops the caddy server
```

## Usage

> Warning: As using multiple proxy plugins on a single Dokku installation can lead to issues routing requests to apps, doing so should be avoided. As the default proxy implementation is nginx, users are encouraged to stop the nginx service before switching to Caddy.

The Caddy plugin has specific rules for routing requests:

- Caddy integration is exposed via docker labels attached to containers. Changes in labels require either app deploys or rebuilds.
- While Caddy will respect labels associated with other containers, only `web` containers have Caddy labels injected by the plugin.
- Only `http:80` and `https:443` port mappings are supported.
- Caddy will automatically enable SSL if the letsencrypt email property is set. SSL will be disabled otherwise.
- If no `http:80` mapping is found, the first `http` port mapping is used for http requests.
- If no `https:443` mapping is found, the first `https` port mapping is used for https requests.
- If no `https` mapping is found, the container port from `http:80` will be used for https requests.
- Requests are routed as soon as the container is running and passing healthchecks.

### Switching to Caddy

To use the Caddy plugin, use the `proxy:set` command for the app in question:

```shell
dokku proxy:set node-js-app caddy
```

This will enable the docker label-based Caddy integration. All future deploys will inject the correct labels for Caddy to read and route requests to containers. Due to the docker label-based integration used by Caddy, a single deploy or rebuild will be required before requests will route successfully.

```shell
dokku ps:rebuild node-js-app
```

Any changes to domains or port mappings will also require either a deploy or rebuild.

### Starting Caddy container

Caddy can be started via the `caddy:start` command. This will start a Caddy container via the `docker compose up` command.

```shell
dokku caddy:start
```

### Stopping the Caddy container

Caddy may be stopped via the `caddy:stop` command.

```shell
dokku caddy:stop
```

The Caddy container will be stopped and removed from the system. If the container is not running, this command will do nothing.

### Showing the Caddy compose config

For debugging purposes, it may be useful to show the Caddy compose config. This can be achieved via the `caddy:show-config` command.

```shell
dokku caddy:show-config
```

### Customizing the Caddy container image

While the default Caddy image is hardcoded, users may specify an alternative by setting the `image` property with the `--global` flag:

```shell
dokku caddy:set --global image lucaslorentz/caddy-docker-proxy:2.7
```

#### Checking the Caddy container's logs

It may be necessary to check the Caddy container's logs to ensure that Caddy is operating as expected. This can be performed with the `caddy:logs` command.

```shell
dokku caddy:logs
```

This command also supports the following modifiers:

```shell
--num NUM        # the number of lines to display
--tail           # continually stream logs
```

You can use these modifiers as follows:

```shell
dokku caddy:logs --tail --num 10
```

The above command will show logs continually from the vector container, with an initial history of 10 log lines

### Changing the Caddy log level

Caddy log output is set to `ERROR` by default. It may be changed by setting the `log-level` property with the `--global` flag:

```shell
dokku caddy:set --global log-level DEBUG
```

After modifying,  the Caddy container will need to be restarted.

### SSL Configuration

The caddy plugin only supports automatic ssl certificates from it's letsencrypt integration. Managed certificates provided by the `certs` plugin are ignored.

#### Enabling letsencrypt integration

By default, letsencrypt is disabled and https port mappings are ignored. To enable, set the `letsencrypt-email` property with the `--global` flag:

```shell
dokku caddy:set --global letsencrypt-email automated@dokku.sh
```

After enabling, the Caddy container will need to be restarted and apps will need to be rebuilt. All http requests will then be redirected to https.

#### Customizing the letsencrypt server

The letsencrypt integration is set to the production letsencrypt server by default. To change this, set the `letsencrypt-server` property with the `--global` flag:

```shell
dokku caddy:set --global letsencrypt-server https://acme-staging-v02.api.letsencrypt.org/directory
```

After enabling, the Caddy container will need to be restarted and apps will need to be rebuilt to retrieve certificates from the new server.

### Using Caddy's Internal TLS server

To switch to Caddy's internal TLS server for certificate provisioning, set the `tls-internal` property. This can only be set on a per-app basis.

```shell
dokku caddy:set node-js-app tls-internal true
```

## Displaying Caddy reports for an app

You can get a report about the app's Caddy config using the `caddy:report` command:

```shell
dokku caddy:report
```

```
=====> node-js-app caddy information
       Caddy image:                   lucaslorentz/caddy-docker-proxy:2.7
       Caddy letsencrypt email:
       Caddy letsencrypt server:
       Caddy log level:               ERROR
       Caddy polling interval:        5s
       Caddy tls internal:            false
=====> python-app caddy information
       Caddy image:                   lucaslorentz/caddy-docker-proxy:2.7
       Caddy letsencrypt email:
       Caddy letsencrypt server:
       Caddy log level:               ERROR
       Caddy polling interval:        5s
       Caddy tls internal:            false
=====> ruby-app caddy information
       Caddy image:                   lucaslorentz/caddy-docker-proxy:2.7
       Caddy letsencrypt email:
       Caddy letsencrypt server:
       Caddy log level:               ERROR
       Caddy polling interval:        5s
       Caddy tls internal:            false
```

You can run the command for a specific app also.

```shell
dokku caddy:report node-js-app
```

```
=====> node-js-app caddy information
       Caddy image:                   lucaslorentz/caddy-docker-proxy:2.7
       Caddy letsencrypt email:
       Caddy letsencrypt server:
       Caddy log level:               ERROR
       Caddy polling interval:        5s
       Caddy tls internal:            false
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku caddy:report node-js-app --caddy-image
```
