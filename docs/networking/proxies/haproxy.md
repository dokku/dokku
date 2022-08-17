# Haproxy Proxy

> New as of 0.28.0

Dokku provides integration with the [Haproxy](http://www.haproxy.org) proxy service by utilizing the Docker label-based integration implemented by [EasyHaproxy](https://github.com/byjg/docker-easy-haproxy).

```
haproxy:report [<app>] [<flag>]            # Displays a haproxy report for one or more apps
haproxy:logs [--num num] [--tail]          # Display haproxy log output
haproxy:set <app> <property> (<value>)     # Set or clear an haproxy property for an app
haproxy:show-config <app>                  # Display haproxy compose config
haproxy:start                              # Starts the haproxy server
haproxy:stop                               # Stops the haproxy server
```

## Usage

> Warning: As using multiple proxy plugins on a single Dokku installation can lead to issues routing requests to apps, doing so should be avoided. As the default proxy implementation is nginx, users are encouraged to stop the nginx service before switching to Haproxy.

The Haproxy plugin has specific rules for routing requests:

- Haproxy integration is exposed via docker labels attached to containers. Changes in labels require either app deploys or rebuilds.
- While Haproxy will respect labels associated with other containers, only `web` containers have Haproxy labels injected by the plugin.
- Only `http:80` port mappings are supported at this time.
- Requests are routed as soon as the container is running and passing healthchecks.

### Switching to Haproxy

To use the Haproxy plugin, use the `proxy:set` command for the app in question:

```shell
dokku proxy:set node-js-app haproxy
```

This will enable the docker label-based Haproxy integration. All future deploys will inject the correct labels for Haproxy to read and route requests to containers. Due to the docker label-based integration used by Haproxy, a single deploy or rebuild will be required before requests will route successfully.

```shell
dokku ps:rebuild node-js-app
```

Any changes to domains or port mappings will also require either a deploy or rebuild.

### Starting Haproxy container

Haproxy can be started via the `haproxy:start` command. This will start a Haproxy container via the `docker compose up` command.

```shell
dokku haproxy:start
```

### Stopping the Haproxy container

Haproxy may be stopped via the `haproxy:stop` command.

```shell
dokku haproxy:stop
```

The Haproxy container will be stopped and removed from the system. If the container is not running, this command will do nothing.

### Showing the Haproxy compose config

For debugging purposes, it may be useful to show the Haproxy compose config. This can be achieved via the `haproxy:show-config` command.

```shell
dokku haproxy:show-config
```

### Customizing the Haproxy container image

While the default Haproxy image is hardcoded, users may specify an alternative by setting the `image` property with the `--global` flag:

```shell
dokku haproxy:set --global image byjg/easy-haproxy:4.0.0
```

#### Checking the Haproxy container's logs

It may be necessary to check the Haproxy container's logs to ensure that Haproxy is operating as expected. This can be performed with the `haproxy:logs` command.

```shell
dokku haproxy:logs
```

This command also supports the following modifiers:

```shell
--num NUM        # the number of lines to display
--tail           # continually stream logs
```

You can use these modifiers as follows:

```shell
dokku haproxy:logs --tail --num 10
```

The above command will show logs continually from the vector container, with an initial history of 10 log lines

## Displaying Haproxy reports for an app

You can get a report about the app's Haproxy config using the `haproxy:report` command:

```shell
dokku haproxy:report
```

```
=====> node-js-app haproxy information
       Haproxy image:                   byjg/easy-haproxy:4.0.0
=====> python-app haproxy information
       Haproxy image:                   byjg/easy-haproxy:4.0.0
=====> ruby-app haproxy information
       Haproxy image:                   byjg/easy-haproxy:4.0.0
```

You can run the command for a specific app also.

```shell
dokku haproxy:report node-js-app
```

```
=====> node-js-app haproxy information
       Haproxy image:                   byjg/easy-haproxy:4.0.0
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku haproxy:report node-js-app --haproxy-image
```
