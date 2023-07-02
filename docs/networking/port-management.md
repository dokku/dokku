# Port Management

> New as of 0.31.0, replaces the previous `proxy:ports*` commands

```
ports:list <app>                        # List port mappings for an app
ports:add <app> <scheme>:<host-port>:<container-port> [<scheme>:<host-port>:<container-port>...]           # Add port mappings to an app
ports:clear <app>                       # Clear all port mappings for an app
ports:remove <app> <host-port> [<host-port>|<scheme>:<host-port>:<container-port>...]                      # Remove specific port mappings from an app
ports:set <app> <scheme>:<host-port>:<container-port> [<scheme>:<host-port>:<container-port>...]           # Set port mappings for an app
```

## Usage

> Warning: Mapping alternative ports may conflict with the active firewall installed on your server or hosting provider. Such software includes - but is not limited to - AWS Security Groups, iptables, and UFW. Please consult the documentation for those softwares as applicable.
>
> Users should also avoid setting the `PORT` environment variable. Dokku will use port mappings to set this value. Overriding this manually may cause issues in application routing.

You can now configure `host -> container` port mappings with the `ports:*` commands. This mapping is currently supported by the built-in nginx-vhosts plugin.

By default, buildpack apps and dockerfile apps **without** explicitly exposed ports (i.e. using the `EXPOSE` directive) will be configured with a listener on port `80` (and additionally a listener on 443 if ssl is enabled) that will proxy to the application container on port `5000`. Dockerfile apps **with** explicitly exposed ports will be configured with a listener on each exposed port and will proxy to that same port of the deployed application container.

> Note: This default behavior **will not** be automatically changed on subsequent pushes and must be manipulated with the `ports:*` commands detailed below.

### Port Scheme

The proxy port scheme is as follows:

- `SCHEME:HOST_PORT:CONTAINER_PORT`

The scheme metadata can be used by proxy implementations in order to properly handle proxying of requests. For example, the built-in `nginx-vhosts` proxy implementation supports the `http`, `https`, `grpc` and `grpcs` schemes. 
For the `grpc` and `grpcs` see [nginx blog post on grpc](https://www.nginx.com/blog/nginx-1-13-10-grpc/).

Developers of proxy implementations are encouraged to use whatever schemes make the most sense, and ignore configurations which they do not support. For instance, a `udp` proxy implementation can safely ignore `http` and `https` port mappings.

To change the proxy implementation in use for an application, use the `proxy:set` command:

```shell
# no validation will be performed against
# the specified proxy implementation
dokku proxy:set node-js-app nginx
```

### Listing port mappings

To inspect the port mapping for a given application, use the `ports:list` command:

```shell
dokku ports:list node-js-app
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

### Adding a custom port mapping

There are cases where we may wish for the service to be listening on more than one port, such as port 8080. Normally, this would not be possible:

```shell
curl http://node-js-app.dokku.me:8080
```

```
curl: (7) Failed to connect to node-js-app.dokku.me port 8080: Connection refused
```

However, we can use the `ports:add` command to add a second external port mapping - `8080` - to our application's port `5000`.

```shell
dokku ports:add node-js-app http:8080:5000
```

```
-----> Setting config vars
       DOKKU_PROXY_PORT_MAP: http:80:5000 http:8080:5000
-----> Configuring node-js-app.dokku.me...(using built-in template)
-----> Creating http nginx.conf
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

### Setting all port mappings at once

Port mappings can also be force set using the `ports:set` command.

```shell
dokku ports:set node-js-app http:8080:5000
```

```
-----> Setting config vars
       DOKKU_PROXY_PORT_MAP: http:80:5000 http:8080:5000
-----> Configuring node-js-app.dokku.me...(using built-in template)
-----> Creating http nginx.conf
       Reloading nginx
```

### Removing a port mapping

A port mapping can be removed using the `ports:remove` command if it no longer necessary:

```shell
dokku ports:remove node-js-app http:80:5000
```

Ports may also be removed by specifying only the `host-port` value. This effectively acts as a wildcard and removes all mappings for that particular host port.

```shell
dokku ports:remove node-js-app http:80
```

## Port management by Deployment Method

> Warning: If you set a port map but _do not have a global domain set_, Dokku will reset that map upon first deployment.

### Buildpacks

For buildpack deployments, your application _must_ respect the `PORT` environment variable. We will typically set this to port `5000`, but this is not guaranteed. If you do not respect the `PORT` environment variable, your containers may start but your services will not be accessible outside of that container.

### Dockerfile

> Changed as of 0.5.0

Dokku's default proxy implementation - nginx - supports HTTP and GRPC request proxying. At this time, we do not support proxying plain TCP or UDP ports. UDP ports can be exposed by disabling the nginx proxy with `dokku proxy:disable myapp` and manually exposing the ports via the `docker-options` plugin. For example, `dokku docker-options:add myapp deploy "-p 2456:2456/udp"`. If you would like to investigate alternative proxy methods, please refer to our [proxy management documentation](/docs/networking/proxy-management.md).

#### Applications using EXPOSE

Dokku will extract all tcp ports exposed using the `EXPOSE` directive (one port per line) and setup nginx to proxy the same port numbers to listen publicly. If you would like to change the exposed port, you should do so within your `Dockerfile`.

For example, if the Dokku installation is configured with the domain `dokku.me` and an application named `node-js-app` is deployed with following Dockerfile:

```
FROM ubuntu:20.04
EXPOSE 1234
RUN python -m SimpleHTTPServer 1234
```

The application would be exposed to the user at `node-js-app.dokku.me:1234`. If this is not desired, the following application configuration may be applied:

```shell
# add a port mapping to port 80
dokku ports:add node-js-app http:80:1234

# remove the incorrect port mapping
dokku ports:remove node-js-app http:1234:1234
```

#### Applications not using EXPOSE

Any application that does not use an `EXPOSE` directive will result in Dokku defaulting to port `5000`. This behavior mimics the behavior of a Buildpack deploy. If your application _does not_ support the `PORT` environment variable, then you will either need to:

- modify your application to support the `PORT` environment variable.
- switch to using an `EXPOSE` directive in your Dockerfile.

#### Switching between `EXPOSE` usage modes

When switching between `EXPOSE` usage modes, it is important to reset your port management. The following two commands can be used to reset your state and redeploy your application.

```shell
# assuming your application is called `node-js-app`
dokku config:unset --no-restart node-js-app DOKKU_DOCKERFILE_PORTS PORT
dokku ports:clear node-js-app
```

### Docker Image

When deploying an image, we will use `docker inspect` to extract the `ExposedPorts` configuration and if defined, use that to populate port mapping. If this behavior is not desired, you can override that configuration variable with the following commands.

```shell
# assuming your application is called `node-js-app`
dokku config:set node-js-app DOKKU_DOCKERFILE_PORTS="1234/tcp 80/tcp"
dokku ports:clear node-js-app
```

All other port-related behavior is the same as when deploying via Dockerfile.

### Displaying ports reports for an app

You can get a report about the app's ports status using the `ports:report` command:

```shell
dokku ports:report
```

```
=====> node-js-app ports information
       Port map:      http:80:5000 https:443:5000
=====> python-sample ports information
       Port map:      http:80:5000
=====> ruby-sample ports information
       Port map:      http:80:5000
```

You can run the command for a specific app also.

```shell
dokku ports:report node-js-app
```

```
=====> node-js-app ports information
       Port map:      http:80:5000 https:443:5000
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku ports:report node-js-app --ports-map
```
