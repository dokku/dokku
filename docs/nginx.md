# Nginx

Dokku uses nginx as it's server for routing requests to specific applications. By default, access and error logs are written for each app to `/var/log/nginx/${APP}-access.log` and `/var/log/nginx/${APP}-error.log` respectively

## TLS/SPDY support

Dokku provides easy TLS/SPDY support out of the box. This can be done app-by-app or for all subdomains at once. Note that whenever TLS support is enabled SPDY is also enabled.

### SSL Configuration

In 0.4.0, SSL Configuration has been replaced by the [`certs` plugin](http://progrium.viewdocs.io/dokku/deployment/ssl-configuration)). For users of dokku 0.3.x, please refer to the following sections.

### Per App

To enable TLS connections to to one of your applications, do the following:

* Create a key file and a cert file.
 * You can find detailed steps for generating a self-signed certificate at https://devcenter.heroku.com/articles/ssl-certificate-self
 * If you are not paranoid and need it just for a DEV or STAGING app, you can use http://www.selfsignedcertificate.com/ to generate your 2 files more easily.
* Rename your files to server.key and server.crt
* tar these 2 files together, *without* subdirectories. Example: tar cvf cert-key.tar server.crt server.key
* Install the pair for your app, like this: ssh dokku@ip-of-your-dokku-server nginx:import-ssl < cert-key.tar

You will need to repeat the steps above for each domain used to serve your app. You can't simply create a single tar with all key/cert files in it (see https://github.com/progrium/dokku/issues/1195).


### All Subdomains

To enable TLS connections for all your applications at once you will need a wildcard TLS certificate.

To enable TLS across all apps, copy or symlink the `.crt` and `.key` files into the  `/home/dokku/tls` folder (create this folder if it doesn't exist) as `server.crt` and `server.key` respectively. Then, enable the certificates by editing `/etc/nginx/conf.d/dokku.conf` and uncommenting these two lines (remove the #):

```
ssl_certificate /home/dokku/tls/server.crt;
ssl_certificate_key /home/dokku/tls/server.key;
```

The nginx configuration will need to be reloaded in order for the updated TLS configuration to be applied. This can be done either via the init system or by re-deploying the application. Once TLS is enabled, the application will be accessible by `https://` (redirection from `http://` is applied as well).

**Note**: TLS will not be enabled unless the application's VHOST matches the certificate's name. (i.e. if you have a cert for *.example.com TLS won't be enabled for something.example.org or example.net)

### HSTS Header

The [HSTS header](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) is an HTTP header that can inform browsers that all requests to a given site should be made via HTTPS. dokku does not, by default, enable this header. It is thus left up to you, the user, to enable it for your site.

Beware that if you enable the header and a subsequent deploy of your application results in an HTTP deploy (for whatever reason), the way the header works means that a browser will not attempt to request the HTTP version of your site if the HTTPS version fails.

### Importing ssl certificates

You can import ssl certificates via tarball using the following command:

``` bash
dokku nginx:import-ssl myapp < archive-of-certs.tar
```

This archive is expanded via `tar xvf`. It should contain `server.crt` and `server.key`.

## Customizing the nginx configuration

> New as of 0.3.17.

Dokku currently templates out an nginx configuration that is included in the `nginx-vhosts` plugin. If you'd like to provide a custom template for your application, you should copy the existing template - ssl or non-ssl - into your application repository's root directory as the file `nginx.conf.template`. The next time you deploy, Nginx will use your template instead of the default.

> New as of 0.3.10.

You may also place this file on disk at the path `/home/dokku/myapp/nginx.conf.template`. If placed on disk on the dokku server, the template file **must** be owned by the user `dokku:dokku`.

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

    include $DOKKU_ROOT/$APP/nginx.conf.d/*.conf;

. That means you can put additional configuration in separate files, for example to limit the uploaded body size to 50 megabytes, do

    mkdir /home/dokku/myapp/nginx.conf.d/
    echo 'client_max_body_size 50M;' > /home/dokku/myapp/nginx.conf.d/upload.conf
    chown dokku:dokku /home/dokku/myapp/nginx.conf.d/upload.conf
    service nginx reload

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

If desired, it is possible to disable vhosts by setting the `NO_VHOST` environment variable:

```shell
dokku config:set myapp NO_VHOST=1
```

On subsequent deploys, the nginx virtualhost will be discarded. This is useful when deploying internal-facing services that should not be publicly routeable.

### Domains plugin

> New as of 0.3.10

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

### Container network interface binding

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

# Default site

By default, dokku will route any received request with an unknown HOST header value to the lexicographically first site in the nginx config stack. If this is not the desired behavior, you may want to add the following configuration to nginx. This will catch all unknown HOST header values and return a `410 Gone` response. You can replace the `return 410;` with `return 444;` which will cause nginx to not respond to requests that do not match known domains (connection refused).

```
server {
  listen 80 default_server;
  listen [::]:80 default_server;

  server_name _;
  return 410;
  log_not_found off;
}
```
