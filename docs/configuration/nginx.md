# Nginx Configuration

Dokku uses nginx as its server for routing requests to specific applications. By default, access and error logs are written for each app to `/var/log/nginx/${APP}-access.log` and `/var/log/nginx/${APP}-error.log` respectively

```
nginx:access-logs <app> [-t]             # Show the nginx access logs for an application (-t follows)
nginx:build-config <app>                 # (Re)builds nginx config for given app
nginx:error-logs <app> [-t]              # Show the nginx error logs for an application (-t follows)
```

## Customizing the nginx configuration

> New as of 0.5.0

Dokku uses a templating library by the name of [sigil](https://github.com/gliderlabs/sigil) to generate nginx configuration for each app. You may also provide a custom template for your application as follows:

- Copy the following example template to a file named `nginx.conf.sigil` and either:
  - If using a buildpack application, you __must__ check it into the root of your app repo.
  - `ADD` it to your dockerfile `WORKDIR`
  - if your dockerfile has no `WORKDIR`, `ADD` it to the `/app` folder

> When using a custom `nginx.conf.sigil` file, depending upon your application configuration, you *may* be exposing the file externally. As this file is extracted before the container is run, you can, safely delete it in a custom `entrypoint.sh` configured in a Dockerfile `ENTRYPOINT`.

> The default template is available [here](https://github.com/dokku/dokku/blob/master/plugins/nginx-vhosts/templates/nginx.conf.sigil), and can be used as a guide for your own, custom `nginx.conf.sigil` file. Please refer to the appropriate template file version for your Dokku version.

### Available template variables

```
{{ .APP }}                          Application name
{{ .APP_SSL_PATH }}                 Path to SSL certificate and key
{{ .DOKKU_ROOT }}                   Global Dokku root directory (ex: app dir would be `{{ .DOKKU_ROOT }}/{{ .APP }}`)
{{ .DOKKU_APP_LISTENERS }}          List of IP:PORT pairs of app containers
{{ .PROXY_PORT }}                   Non-SSL nginx listener port (same as `DOKKU_PROXY_PORT` config var)
{{ .PROXY_SSL_PORT }}               SSL nginx listener port (same as `DOKKU_PROXY_SSL_PORT` config var)
{{ .NOSSL_SERVER_NAME }}            List of non-SSL VHOSTS
{{ .PROXY_PORT_MAP }}               List of port mappings (same as `DOKKU_PROXY_PORT_MAP` config var)
{{ .PROXY_UPSTREAM_PORTS }}         List of configured upstream ports (derived from `DOKKU_PROXY_PORT_MAP` config var)
{{ .RAW_TCP_PORTS }}                List of exposed tcp ports as defined by Dockerfile `EXPOSE` directive (**Dockerfile apps only**)
{{ .SSL_INUSE }}                    Boolean set when an app is SSL-enabled
{{ .SSL_SERVER_NAME }}              List of SSL VHOSTS
```

> Note: Application config variables are available for use in custom templates. To do so, use the form of `{{ var "FOO" }}` to access a variable named `FOO`.

### Customizing via configuration files included by the default templates

The default nginx.conf template will include everything from your apps `nginx.conf.d/` subdirectory in the main `server {}` block (see above):

```
include {{ .DOKKU_ROOT }}/{{ .APP }}/nginx.conf.d/*.conf;
```

That means you can put additional configuration in separate files, for example to limit the uploaded body size to 50 megabytes, do

```shell
mkdir /home/dokku/node-js-app/nginx.conf.d/
echo 'client_max_body_size 50m;' > /home/dokku/node-js-app/nginx.conf.d/upload.conf
chown dokku:dokku /home/dokku/node-js-app/nginx.conf.d/upload.conf
service nginx reload
```

The example above uses additional configuration files directly on the Dokku host. Unlike the `nginx.conf.sigil` file, these additional files will not be copied over from your application repo, and thus need to be placed in the `/home/dokku/node-js-app/nginx.conf.d/` directory manually.

For PHP Buildpack users, you will also need to provide a `Procfile` and an accompanying `nginx.conf` file to customize the nginx config *within* the container. The following are example contents for your `Procfile`

    web: vendor/bin/heroku-php-nginx -C nginx.conf -i php.ini php/
    
Your `nginx.conf` file - not to be confused with Dokku's `nginx.conf.sigil` - would also need to be configured as shown in this example:

    client_max_body_size 50m;
    location / {
        index index.php;
        try_files $uri $uri/ /index.php$is_args$args;
    }

Please adjust the `Procfile` and `nginx.conf` file as appropriate.

## Domains plugin

See the [domain configuration documentation](/docs/configuration/domains.md).

## Customizing hostnames

See the [customizing hostnames documentation](/docs/configuration/domains.md#customizing-hostnames).

## Disabling VHOSTS

See the [disabling vhosts documentation](/docs/configuration/domains.md#disabling-vhosts).

## Default site

See the [default site documentation](/docs/configuration/domains.md#default-site).

## Running behind a load balancer

See the [load balancer documentation](/docs/configuration/ssl.md#running-behind-a-load-balancer).

## HSTS Header

See the [HSTS documentation](/docs/configuration/ssl.md#hsts-header).

## SSL Configuration

See the [ssl documentation](/docs/configuration/ssl.md).

## Disabling Nginx

See the [proxy documentation](/docs/advanced-usage/proxy-management.md).

## Managing Proxy Port mappings

See the [proxy documentation](/docs/advanced-usage/proxy-management.md#proxy-port-mapping).
