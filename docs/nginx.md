# Nginx Configuration

Dokku uses nginx as its server for routing requests to specific applications. By default, access and error logs are written for each app to `/var/log/nginx/${APP}-access.log` and `/var/log/nginx/${APP}-error.log` respectively

```
nginx:access-logs <app> [-t]                     Show the nginx access logs for an application (-t follows)
nginx:build-config <app>                         (Re)builds nginx config for given app
nginx:disable <app>                              Disable nginx for an application (forces container binding to external interface)
nginx:enable <app>                               Enable nginx for an application
nginx:error-logs <app> [-t]                      Show the nginx error logs for an application (-t follows)
```

## Customizing the nginx configuration

> New as of 0.3.10.

Dokku currently templates out an nginx configuration that is included in the `nginx-vhosts` plugin. If you'd like to provide a custom template for your application, there are a few options:

- Copy the existing pertinent (ssl or non-ssl) template found (by default) in `/var/lib/dokku/plugins/available/nginx-vhosts/templates`, place it either in `/home/dokku/APP` or check it into the root of your app repo, and name it one of the following:
  - `nginx.conf.template` (since 0.3.10)
  - `nginx.conf.ssl_terminated.template` (since 0.4.0)
  - `nginx.ssl.conf.template` (since 0.4.2)

> If placed on the dokku server, the template file **must** be owned by user and group `dokku:dokku`.

For instance - assuming defaults - to customize the nginx template in use for the `myapp` application, create the file `nginx.conf.template` in your repo or on disk with the with the following contents:

```
server {
  listen      [::]:80;
  listen      80;
  server_name $NOSSL_SERVER_NAME;
  access_log  /var/log/nginx/${APP}-access.log;
  error_log   /var/log/nginx/${APP}-error.log;

  # set a custom header for requests
  add_header X-Served-By www-ec2-01;

  location    / {
    proxy_pass  http://$APP;
    proxy_http_version 1.1;
    proxy_set_header Upgrade \$http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host \$http_host;
    proxy_set_header X-Forwarded-Proto \$scheme;
    proxy_set_header X-Forwarded-For \$remote_addr;
    proxy_set_header X-Forwarded-Port \$server_port;
    proxy_set_header X-Request-Start \$msec;
  }
  include $DOKKU_ROOT/$APP/nginx.conf.d/*.conf;
}
```

The above is a sample http configuration that adds an `X-Served-By` header to requests.

A few tips for custom nginx templates:

- Special characters - dollar signs, spaces inside of quotes, etc. - should be escaped with a single backslash or can cause deploy failures.
- Templated files will be validated via `nginx -t`.
- Application environment variables can be used within your nginx configuration.

After your changes a `dokku deploy myapp` will regenerate the `/home/dokku/myapp/nginx.conf` file which is then used.

### Customizing via configuration files included by the default templates

The default nginx.conf- templates will include everything from your apps `nginx.conf.d/` subdirectory in the main `server {}` block (see above):

```
include $DOKKU_ROOT/$APP/nginx.conf.d/*.conf;
```

That means you can put additional configuration in separate files, for example to limit the uploaded body size to 50 megabytes, do

```shell
mkdir /home/dokku/myapp/nginx.conf.d/
echo 'client_max_body_size 50M;' > /home/dokku/myapp/nginx.conf.d/upload.conf
chown dokku:dokku /home/dokku/myapp/nginx.conf.d/upload.conf
service nginx reload
```

## Customizing hostnames

Applications typically have the following structure for their hostname:

```
scheme://subdomain.domain.tld
```

The `subdomain` is inferred from the pushed application name, while the `domain` is set during initial configuration in the `$DOKKU_ROOT/VHOST` file.

You can optionally override this in a plugin by implementing the `nginx-hostname` plugin trigger. For example, you can reverse the subdomain with the following sample `nginx-hostname` plugin trigger:

```shell
#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"; SUBDOMAIN="$2"; VHOST="$3"

NEW_SUBDOMAIN=`echo $SUBDOMAIN | rev`
echo "$NEW_SUBDOMAIN.$VHOST"
```

If the `nginx-hostname` has no output, the normal hostname algorithm will be executed.

You can also use the built-in `domains` plugin to handle:

### Disabling VHOSTS

If desired, it is possible to disable vhosts by setting the environment variable `NO_VHOST=1`:

```shell
dokku config:set myapp NO_VHOST=1
```

On subsequent deploys, the nginx virtualhost will be discarded. This is useful when deploying internal-facing services that should not be publicly routeable. As of 0.4.0, nginx will still be configured to proxy your app on some random high port. This allows internal services to maintain the same port between deployments. You may change this port by setting `DOKKU_NGINX_PORT` and/or `DOKKU_NGINX_SSL_PORT` (for services configured to use SSL.)

### Domains plugin

> New as of 0.3.10

```shell
domains:add <app> DOMAIN                         Add a custom domain to app
domains <app>                                    List custom domains for app
domains:clear <app>                              Clear all custom domains for app
domains:remove <app> DOMAIN                      Remove a custom domain from app
```

The domains plugin allows you to specify custom domains for applications. This plugin is aware of any ssl certificates that are imported via `nginx:import-ssl`. Be aware that setting `NO_VHOST` will override any custom domains.

Custom domains are also backed up via the built-in `backup` plugin

```shell
# where `myapp` is the name of your app

# add a domain to an app
dokku domains:add myapp example.com

# list custom domains for app
dokku domains myapp

# clear all custom domains for app
dokku domains:clear myapp

# remove a custom domain from app
dokku domains:remove myapp example.com
```

## Container network interface binding

> New as of 0.3.13

The deployed docker container running your app's web process will bind to either the internal docker network interface (i.e. `docker inspect --format '{{ .NetworkSettings.IPAddress }}' $CONTAINER_ID`) or an external interface (i.e. 0.0.0.0) depending on dokku's VHOST configuration. Dokku will attempt to bind to the internal docker network interface unless you specifically set NO_VHOST for the given app or your dokku installation is not setup to use VHOSTS (i.e. $DOKKU_ROOT/VHOST is missing or $DOKKU_ROOT/HOSTNAME is set to an IPv4 or IPv6 address)

```shell
# container bound to docker interface
root@dokku:~/dokku# docker ps
CONTAINER ID        IMAGE                      COMMAND                CREATED              STATUS              PORTS               NAMES
1b88d8aec3d1        dokku/node-js-app:latest   "/bin/bash -c '/star   About a minute ago   Up About a minute                       goofy_albattani

root@dokku:~/dokku# docker inspect --format '{{ .NetworkSettings.IPAddress }}' goofy_albattani
172.17.0.6

# container bound to all interfaces (previous default)
root@dokku:/home/dokku# docker ps
CONTAINER ID        IMAGE                      COMMAND                CREATED              STATUS              PORTS                     NAMES
d6499edb0edb        dokku/node-js-app:latest   "/bin/bash -c '/star   About a minute ago   Up About a minute   0.0.0.0:49153->5000/tcp   nostalgic_tesla
```

## Default site

By default, dokku will route any received request with an unknown HOST header value to the lexicographically first site in the nginx config stack. If this is not the desired behavior, you may want to add the following configuration to the global nginx configuration. This will catch all unknown HOST header values and return a `410 Gone` response. You can replace the `return 410;` with `return 444;` which will cause nginx to not respond to requests that do not match known domains (connection refused).

```
server {
  listen 80 default_server;
  listen [::]:80 default_server;

  server_name _;
  return 410;
  log_not_found off;
}
```

You may also wish to use a separate vhost in your `/etc/nginx/sites-enabled` directory. To do so, create the vhost in that directory as `/etc/nginx/sites-enabled/00-default.conf`. You will also need to change two lines in the main `nginx.conf`:

```
# Swap both conf.d include line and the sites-enabled include line. From:
include /etc/nginx/conf.d/*.conf;
include /etc/nginx/sites-enabled/*;

# to the following

include /etc/nginx/sites-enabled/*;
include /etc/nginx/conf.d/*.conf;
```

Alternatively, you may push an app to your dokku host with a name like "00-default". As long as it lists first in `ls /home/dokku/*/nginx.conf | head`, it will be used as the default nginx vhost.

## Running behind a load balancer

See the [load balancer documentation](/dokku/deployment/ssl-configuration/#running-behind-a-load-balancer).

## HSTS Header

See the [HSTS documentation](/dokku/deployment/ssl-configuration/#hsts-header).

## SSL Configuration

See the [ssl documentation](/dokku/deployment/ssl-configuration/).
