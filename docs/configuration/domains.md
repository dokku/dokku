# Domain Configuration

> New as of 0.3.10

```
domains:add <app> DOMAIN                 # Add a domain to app
domains:add-global <domain>              # Add global domain name
domains [<app>]                          # List domains
domains:clear <app>                      # Clear all domains for app
domains:disable <app>                    # Disable VHOST support
domains:enable <app>                     # Enable VHOST support
domains:remove <app> DOMAIN              # Remove a domain from app
domains:remove-global <domain>           # Remove global domain name
```

> Adding a domain before deploying an application will result in port mappings being set. This may cause issues for applications that use non-standard ports, as those will not be automatically detected. Please refer to the [proxy documentation](/dokku/advanced-usage/proxy-management/) for information as to how to reconfigure the mappings.

## Customizing hostnames

Applications typically have the following structure for their hostname:

```
scheme://subdomain.domain.tld
```

The `subdomain` is inferred from the pushed application name, while the `domain.tld` is set during initial configuration and stored in the `$DOKKU_ROOT/VHOST` file. It can then be modified with `dokku domains:add-global` and `dokku domains:remove-global`. This value is used as a default TLD for all applications on a host.

If a FQDN such as `other.tld` is used as the application name, the default `$DOKKU_ROOT/VHOST` will be ignored and the resulting vhost URL for that application will be `other.tld`. The exception to this rule being that if the FQDN has the same ending as the default vhost (such as `subdomain.domain.tld`), then the entire FQDN will be treated as a subdomain. The application will therefore be deployed at `subdomain.domain.tld.domain.tld`.

You can optionally override this in a plugin by implementing the `nginx-hostname` plugin trigger. For example, you can reverse the subdomain with the following sample `nginx-hostname` plugin trigger:

```shell
#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"; SUBDOMAIN="$2"; VHOST="$3"

NEW_SUBDOMAIN=`echo $SUBDOMAIN | rev`
echo "$NEW_SUBDOMAIN.$VHOST"
```

If the `nginx-hostname` has no output, the normal hostname algorithm will be executed.

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

By default, Dokku will route any received request with an unknown HOST header value to the lexicographically first site in the nginx config stack. If this is not the desired behavior, you may want to add the following configuration to the global nginx configuration. This will catch all unknown HOST header values and return a `410 Gone` response. You can replace the `return 410;` with `return 444;` which will cause nginx to not respond to requests that do not match known domains (connection refused).

```nginx
server {
  listen 80 default_server;
  listen [::]:80 default_server;

  server_name _;
  return 410;
  log_not_found off;
}
```

You may also wish to use a separate vhost in your `/etc/nginx/sites-enabled` directory. To do so, create the vhost in that directory as `/etc/nginx/sites-enabled/00-default.conf`. You will also need to change two lines in the main `nginx.conf`:

```nginx
# Swap both conf.d include line and the sites-enabled include line. From:
include /etc/nginx/conf.d/*.conf;
include /etc/nginx/sites-enabled/*;

# to the following

include /etc/nginx/sites-enabled/*;
include /etc/nginx/conf.d/*.conf;
```

Alternatively, you may push an app to your Dokku host with a name like "00-default". As long as it lists first in `ls /home/dokku/*/nginx.conf | head`, it will be used as the default nginx vhost.
