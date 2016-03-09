# Nginx Configuration

Dokku uses nginx as its server for routing requests to specific applications. By default, access and error logs are written for each app to `/var/log/nginx/${APP}-access.log` and `/var/log/nginx/${APP}-error.log` respectively

```
nginx:access-logs <app> [-t]                                                                 Show the nginx access logs for an application (-t follows)
nginx:build-config <app>                                                                     (Re)builds nginx config for given app
nginx:error-logs <app> [-t]                                                                  Show the nginx error logs for an application (-t follows)
```

## Customizing the nginx configuration

> New as of 0.5.0

Dokku uses a templating library by the name of [sigil](https://github.com/gliderlabs/sigil) to generate nginx configuration for each app. If you'd like to provide a custom template for your application, there are a couple options:

- Copy the following example template to a file named `nginx.conf.sigil` and either:
  - check it into the root of your app repo
  - `ADD` it to your dockerfile `WORKDIR`

### Example Custom Template
```
server {
  listen      [::]:{{ .NGINX_PORT }};
  listen      {{ .NGINX_PORT }};
  server_name {{ .NOSSL_SERVER_NAME }};
  access_log  /var/log/nginx/{{ .APP }}-access.log;
  error_log   /var/log/nginx/{{ .APP }}-error.log;

  # set a custom header for requests
  add_header X-Served-By www-ec2-01;

  location    / {
    proxy_pass  http://{{ .APP }};
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $http_host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Port $server_port;
    proxy_set_header X-Request-Start $msec;
  }
  include {{ .DOKKU_ROOT }}/{{ .APP }}/nginx.conf.d/*.conf;

  upstream {{ .APP }} {
  {{ range .DOKKU_APP_LISTENERS | split " " }}
    server {{ . }};
  {{ end }}
  }
}
```

The above is a sample http configuration that adds an `X-Served-By` header to requests.

### Available template variables
```
{{ .APP }}                          Application name
{{ .APP_SSL_PATH }}                 Path to SSL certificate and key
{{ .DOKKU_ROOT }}                   Global dokku root directory (ex: app dir would be `{{ .DOKKU_ROOT }}/{{ .APP }}`)
{{ .DOKKU_APP_LISTENERS }}          List of IP:PORT pairs of app containers
{{ .NGINX_PORT }}                   Non-SSL nginx listener port (same as `DOKKU_NGINX_PORT` config var)
{{ .NGINX_SSL_PORT }}               SSL nginx listener port (same as `DOKKU_NGINX_SSL_PORT` config var)
{{ .NOSSL_SERVER_NAME }}            List of non-SSL VHOSTS
{{ .RAW_TCP_PORTS }}                List of exposed tcp ports as defined by Dockerfile `EXPOSE` directive (**Dockerfile apps only**)
{{ .SSL_INUSE }}                    Boolean set when an app is SSL-enabled
{{ .SSL_SERVER_NAME }}              List of SSL VHOSTS
```

### Customizing via configuration files included by the default templates

The default nginx.conf template will include everything from your apps `nginx.conf.d/` subdirectory in the main `server {}` block (see above):

```
include {{ .DOKKU_ROOT }}/{{ .APP }}/nginx.conf.d/*.conf;
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

The `subdomain` is inferred from the pushed application name, while the `domain` is set during initial configuration in the `$DOKKU_ROOT/VHOST` file or via `dokku domains:set-global`.

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

### Domains plugin

> New as of 0.3.10

```shell
domains:add <app> DOMAIN                                                                     Add a domain to app
domains [<app>]                                                                              List domains
domains:clear <app>                                                                          Clear all domains for app
domains:disable <app>                                                                        Disable VHOST support
domains:enable <app>                                                                         Enable VHOST support
domains:remove <app> DOMAIN                                                                  Remove a domain from app
domains:set-global <domain>                                                                  Set global domain name
```

### Disabling VHOSTS

If desired, it is possible to disable vhosts with the domains plugin.

```shell
dokku domains:disable myapp
```

On subsequent deploys, the nginx virtualhost will be discarded. This is useful when deploying internal-facing services that should not be publicly routeable. As of 0.4.0, nginx will still be configured to proxy your app on some random high port. This allows internal services to maintain the same port between deployments. You may change this port by setting `DOKKU_NGINX_PORT` and/or `DOKKU_NGINX_SSL_PORT` (for services configured to use SSL.)


The domains plugin allows you to specify custom domains for applications. This plugin is aware of any ssl certificates that are imported via `certs:add`. Be aware that disabling domains (with `domains:disable`) will override any custom domains.

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

## Disabling Nginx

See the [proxy documentation](/dokku/deployment/proxy/).
