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
  - if your dockerfile has no `WORKDIR`, `ADD` it to the `/app` folder

> When using a custom `nginx.conf.sigil` file, depending upon your application configuration, you *may* be exposing the file externally. As this file is extracted before the container is run, you can, safely delete it in a custom `entrypoint.sh` configured in a Dockerfile `ENTRYPOINT`.

### Example Custom Template

Use case: add an `X-Served-By` header to requests
```
upstream {{ .APP }} {
{{ range .DOKKU_APP_LISTENERS | split " " }}
  server {{ . }};
{{ end }}
}
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
}
```

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

> NOTE: Application config variables are available for use in custom templates. To do so, use the form of `{{ var "FOO" }}` to access a variable named `FOO`.


### Example HTTP to HTTPS Custom Template
Use case: a simple dockerfile app that includes `EXPOSE 80`
```
upstream {{ .APP }} {
{{ range .DOKKU_APP_LISTENERS | split " " }}
  server {{ . }};
{{ end }}
}
server {
  listen      [::]:80;
  listen      80;
  server_name {{ .NOSSL_SERVER_NAME }};

  access_log  /var/log/nginx/{{ .APP }}-access.log;
  error_log   /var/log/nginx/{{ .APP }}-error.log;

  return 301 https://$host:443$request_uri;
}
server {
  listen      [::]:443 ssl spdy;
  listen      443 ssl spdy;
  {{ if .SSL_SERVER_NAME }}server_name {{ .SSL_SERVER_NAME }}; {{ end }}

  access_log  /var/log/nginx/{{ .APP }}-access.log;
  error_log   /var/log/nginx/{{ .APP }}-error.log;

  ssl_certificate     {{ .APP_SSL_PATH }}/server.crt;
  ssl_certificate_key {{ .APP_SSL_PATH }}/server.key;

  keepalive_timeout   70;
  add_header          Alternate-Protocol  443:npn-spdy/2;
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
}
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

The example above uses additional configuration files directly on the dokku host. Unlike the `nginx.conf.sigil` file, these additional files will not be copied over from your application repo, and thus need to be placed in the `/home/dokku/myapp/nginx.conf.d/` directory manually.

## Domains plugin

See the [domain-configuration documentation](/dokku/deployment/domain-configuration/).

## Customizing hostnames

See the [customizing hostnames documentation](/dokku/deployment/domain-configuration/#customizing-hostnames).

## Disabling VHOSTS

See the [disabling vhosts documentation](/dokku/deployment/domain-configuration/#disabling-vhosts).

## Default site

See the [default site documentation](/dokku/deployment/domain-configuration/#default-site).

## Running behind a load balancer

See the [load balancer documentation](/dokku/deployment/ssl-configuration/#running-behind-a-load-balancer).

## HSTS Header

See the [HSTS documentation](/dokku/deployment/ssl-configuration/#hsts-header).

## SSL Configuration

See the [ssl documentation](/dokku/deployment/ssl-configuration/).

## Disabling Nginx

See the [proxy documentation](/dokku/proxy/).
