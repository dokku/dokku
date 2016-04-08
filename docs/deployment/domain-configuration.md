# Domain Configuration

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

## Disabling VHOSTS

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
