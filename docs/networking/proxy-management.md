# Proxy Management

> New as of 0.5.0, Enhanced in 0.6.0

```
proxy:disable <app>                      # Disable proxy for app
proxy:enable <app>                       # Enable proxy for app
proxy:ports <app>                        # List proxy port mappings for app
proxy:ports-add <app> <scheme>:<host-port>:<container-port> [<scheme>:<host-port>:<container-port>...]           # Set proxy port mappings for app
proxy:ports-clear <app>                  # Clear all proxy port mappings for app
proxy:ports-remove <app> <host-port> [<host-port>|<scheme>:<host-port>:<container-port>...]                      # Unset proxy port mappings for app
proxy:report [<app>] [<flag>]            # Displays a proxy report for one or more apps
proxy:set <app> <proxy-type>             # Set proxy type for app
```

In Dokku 0.5.0, port proxying was decoupled from the `nginx-vhosts` plugin into the proxy plugin. Dokku 0.6.0 introduced the ability to map host ports to specific container ports. In the future this will allow other proxy software - such as HAProxy or Caddy - to be used in place of nginx.

## Usage

### Container network interface binding

> Changed as of 0.11.0

From Dokku versions `0.5.0` until `0.11.0`, enabling or disabling an application's proxy would **also** control whether or not the application was bound to all interfaces - e.g. `0.0.0.0`. As of `0.11.0`, this is now controlled by the network plugin. Please see the [network documentation](/docs/networking/network.md#container-network-interface-binding) for more information.

### Displaying proxy reports about an app

> New as of 0.8.1

You can get a report about the app's proxy status using the `proxy:report` command:

```shell
dokku proxy:report
```

```
=====> node-js-app proxy information
       Proxy enabled:       true
       Proxy type:          nginx
       Proxy port map:      http:80:5000 https:443:5000
=====> python-sample proxy information
       Proxy enabled:       true
       Proxy type:          nginx
       Proxy port map:      http:80:5000
=====> ruby-sample proxy information
       Proxy enabled:       true
       Proxy type:          nginx
       Proxy port map:      http:80:5000
```

You can run the command for a specific app also.

```shell
dokku proxy:report node-js-app
```

```
=====> node-js-app proxy information
       Proxy enabled:       true
       Proxy type:          nginx
       Proxy port map:      http:80:5000 https:443:5000
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku proxy:report node-js-app --proxy-type
```

### Proxy port mapping

> New as of 0.6.0

You can now configure `host -> container` port mappings with the `proxy:ports-*` commands. This mapping is currently supported by the built-in nginx-vhosts plugin.

To inspect the port mapping for a given application, use the `proxy:ports` command:

```shell
dokku proxy:ports node-js-app
```

```
-----> Port mappings for node-js-app
-----> scheme             host port                 container port
http                      80                        5000
```

The above application is listening on the host's port `80`, which we can test via curl:

```shell
curl http://node-js-app.dokku.me
```

```
Hello World!
```

There are cases where we may wish for the service to be listening on more than one port, such as port 8080. Normally, this would not be possible:

```shell
curl http://node-js-app.dokku.me:8080
```

```
curl: (7) Failed to connect to node-js-app.dokku.me port 8080: Connection refused
```

However, we can use the `proxy:ports-add` command to add a second external port mapping - `8080` - to our application's port `5000`.

```shell
dokku proxy:ports-add node-js-app http:8080:5000
```

```
-----> Setting config vars
       DOKKU_PROXY_PORT_MAP: http:80:5000 http:8080:5000
-----> Configuring node-js-app.dokku.me...(using built-in template)
-----> Creating http nginx.conf
-----> Running nginx-pre-reload
       Reloading nginx
```

We can now test that port 80 still responds properly:

```shell
curl http://node-js-app.dokku.me
```

```
Hello World!
```

And our new listening port of `8080` also works:

```shell
curl http://node-js-app.dokku.me:8080
```

```
Hello World!
```

You can also remove a port mapping that is no longer necessary:

```shell
dokku proxy:ports-remove node-js-app http:80:5000
```

By default, buildpack apps and dockerfile apps **without** explicitly exposed ports (i.e. using the `EXPOSE` directive) will be configured with a listener on port `80` (and additionally a listener on 443 if ssl is enabled) that will proxy to the application container on port `5000`. Dockerfile apps **with** explicitly exposed ports will be configured with a listener on each exposed port and will proxy to that same port of the deployed application container.

> Note: This default behavior **will not** be automatically changed on subsequent pushes and must be manipulated with the `proxy:ports-*` syntax detailed above.

#### Proxy Port Scheme

The proxy port scheme is as follows:

- `SCHEME:HOST_PORT:CONTAINER_PORT`

The scheme metadata can be used by proxy implementations in order to properly handle proxying of requests. For example, the built-in `nginx-vhosts` proxy implementation supports both the `http` and `https` schemes.

Developers of proxy implementations are encouraged to use whatever schemes make the most sense, and ignore configurations which they do not support. For instance, a `udp` proxy implementation can safely ignore `http` and `https` port mappings.

To change the proxy implementation in use for an application, use the `proxy:set` command:

```shell
# no validation will be performed against
# the specified proxy implementation
dokku proxy:set node-js-app nginx
```
