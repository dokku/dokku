# Proxy plugin

> New as of 0.5.0, Enhanced in 0.6.0

```
proxy <app>                                                                                              Show proxy settings for app
proxy:disable <app>                                                                                      Disable proxy for app
proxy:enable <app>                                                                                       Enable proxy for app
proxy:ports <app>                                                                                        List proxy port mappings for app
proxy:ports-add <app> <scheme>:<host-port>:<container-port> [<scheme>:<host-port>:<container-port>...]   Set proxy port mappings for app
proxy:ports-clear <app>                                                                                  Clear all proxy port mappings for app
proxy:ports-remove <app> <host-port> [<host-port>|<scheme>:<host-port>:<container-port>...]              Unset proxy port mappings for app
proxy:set <app> <proxy-type>                                                                             Set proxy type for app
```

In Dokku 0.5.0, port proxying was decoupled from the `nginx-vhosts` plugin into the proxy plugin. Dokku 0.6.0 introduced the ability to map host ports to specific container ports. In the future this will allow other proxy software - such as HAProxy or Caddy - to be used in place of nginx.

## Container network interface binding

> New as of 0.5.0

By default, the deployed docker container running your app's web process will bind to the internal docker network interface (i.e. `docker inspect --format '{{ .NetworkSettings.IPAddress }}' $CONTAINER_ID`). This behavior can be modified per app by disabling the proxy (i.e. `dokku proxy:disable <app>`). In this case, the container will bind to an external interface (i.e. `0.0.0.0`) and your app container will be directly accessible by other hosts on your network.

> If a proxy is disabled, dokku will bind your container's port to a random port on the host for every deploy, e.g. `0.0.0.0:32771->5000/tcp`.

```shell
# container bound to docker interface
$ docker ps
CONTAINER ID        IMAGE                      COMMAND                CREATED              STATUS              PORTS               NAMES
1b88d8aec3d1        dokku/node-js-app:latest   "/bin/bash -c '/star   About a minute ago   Up About a minute                       node-js-app.web.1

# internal IP address for the container
$ docker inspect --format '{{ .NetworkSettings.IPAddress }}' node-js-app.web.1
172.17.0.6

# disable the proxy so it listens on a host ip address
$ dokku proxy:disable node-js-app

# container bound to all interfaces
$ docker ps
CONTAINER ID        IMAGE                      COMMAND                CREATED              STATUS              PORTS                     NAMES
d6499edb0edb        dokku/node-js-app:latest   "/bin/bash -c '/star   About a minute ago   Up About a minute   0.0.0.0:49153->5000/tcp   node-js-app.web.1
```

## Proxy port mapping

> New as of 0.6.0

You can now configure `host -> container` port mappings with the `proxy:ports-*` commands. This mapping is currently supported by the built-in nginx-vhosts plugin.

```shell
$ dokku proxy:ports node-js-app
-----> Port mappings for node-js-app
-----> scheme             host port                 container port
http                      80                        5000

$ curl http://node-js-app.dokku.me
Hello World!

$ curl http://node-js-app.dokku.me:8080
curl: (7) Failed to connect to node-js-app.dokku.me port 8080: Connection refused

$ dokku proxy:ports-add node-js-app http:8080:5000
-----> Setting config vars
       DOKKU_PROXY_PORT_MAP: http:80:5000 http:8080:5000
-----> Configuring node-js-app.dokku.me...(using built-in template)
-----> Creating http nginx.conf
-----> Running nginx-pre-reload
       Reloading nginx

$ curl http://node-js-app.dokku.me
Hello World!

$ curl http://node-js-app.dokku.me:8080
Hello World!
```

By default, buildpack apps and dockerfile apps **without** explicitly exposed ports (i.e. using the `EXPOSE` directive) will be configured with a listener on port `80` (and additionally a listener on 443 if ssl is enabled) that will proxy to the application container on port `5000`. Dockerfile apps **with** explicitly exposed ports will be configured with a listener on each exposed port and will proxy to that same port of the deployed application container.

> NOTE: This default behavior **will not** be automatically changed on subsequent pushes and must be manipulated with the `proxy:ports-*` syntax detailed above.
