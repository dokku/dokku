# Traefik Proxy

Dokku provides integration with the [Traefik](https://traefik.io/) proxy service by utilizing the Docker label-based integration implemented by Traefik.

```
traefik:report [<app>] [<flag>]          # Displays a traefik report for one or more apps
traefik:logs [--num num] [--tail]        # Display traefik log output
traefik:set <app> <property> (<value>)   # Set or clear an traefik property for an app
traefik:show-config <app>                # Display traefik compose config
traefik:start                            # Starts the traefik server
traefik:stop                             # Stops the traefik server
```

## Usage

> Warning: As using multiple proxy plugins on a single Dokku installation can lead to issues routing requests to apps, doing so should be avoided. As the default proxy implementation is nginx, users are encouraged to stop the nginx service before switching to Traefik.

The Traefik plugin has specific rules for routing requests:

- Traefik integration is exposed via docker labels attached to containers. Changes in labels require either app deploys or rebuilds.
- While Traefik will respect labels associated with other containers, only `web` containers have Traefik labels injected by the plugin.
- Only `http:80` and `https:443` port mappings are supported.
- If no `http:80` mapping is found, the first `http` port mapping is used for http requests.
- If no `https:443` mapping is found, the first `https` port mapping is used for https requests.
- If no `https` mapping is found, the container port from `http:80` will be used for https requests.
- Requests are routed as soon as the container is running and passing healthchecks.

### Switching to Traefik

To use the Traefik plugin, use the `proxy:set` command for the app in question:

```shell
dokku proxy:set node-js-app traefik
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

The above command will show logs continually from the vector container, with an initial history of 10 log lines

### Changing the Traefik log level

Traefik log output is set to `ERROR` by default. It may be changed by setting the `log-level` property with the `--global` flag:

```shell
dokku traefik:set --global log-level DEBUG
```

After modifying,  the Traefik container will need to be restarted.

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

### API Access

Traefik exposes an API and Dashboard, which Dokku disables by default for security reasons. It can be exposed and customized as described below.

#### Enabling the api

> Warning: Users enabling the dashboard should also enable api basic auth.

By default, the api is disabled. To enable, set the `api` property with the `--global` flag:

```shell
dokku traefik:set --global api true
```

After enabling, the Traefik container will need to be restarted.

#### Enabling the dashboard

> Warning: Users enabling the dashboard should also enable api basic auth.

By default, the dashboard is disabled. To enable, set the `dashboard` property with the `--global` flag:

```shell
dokku traefik:set --global dashboard true
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
       Traefik api enabled:           false
       Traefik api vhost:             traefik.dokku.me
       Traefik basic auth password:   password
       Traefik basic auth username:   user
       Traefik dashboard enabled:     false
       Traefik image:                 traefik:v2.8
       Traefik letsencrypt email:
       Traefik letsencrypt server:
       Traefik log level:             ERROR
=====> python-app traefik information
       Traefik api enabled:           false
       Traefik api vhost:             traefik.dokku.me
       Traefik basic auth password:   password
       Traefik basic auth username:   user
       Traefik dashboard enabled:     false
       Traefik image:                 traefik:v2.8
       Traefik letsencrypt email:
       Traefik letsencrypt server:
       Traefik log level:             ERROR
=====> ruby-app traefik information
       Traefik api enabled:           false
       Traefik api vhost:             traefik.dokku.me
       Traefik basic auth password:   password
       Traefik basic auth username:   user
       Traefik dashboard enabled:     false
       Traefik image:                 traefik:v2.8
       Traefik letsencrypt email:
       Traefik letsencrypt server:
       Traefik log level:             ERROR
```

You can run the command for a specific app also.

```shell
dokku traefik:report node-js-app
```

```
=====> node-js-app traefik information
       Traefik api enabled:           false
       Traefik api vhost:             traefik.dokku.me
       Traefik basic auth password:   password
       Traefik basic auth username:   user
       Traefik dashboard enabled:     false
       Traefik image:                 traefik:v2.8
       Traefik letsencrypt email:
       Traefik letsencrypt server:
       Traefik log level:             ERROR
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku traefik:report node-js-app --traefik-api-enabled
```
