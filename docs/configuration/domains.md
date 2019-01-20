# Domain Configuration

> New as of 0.3.10

```
domains:add <app> <domain> [<domain> ...]      # Add domains to app
domains:add-global <domain> [<domain> ...]     # Add global domain names
domains:clear <app>                            # Clear all domains for app
domains:disable <app>                          # Disable VHOST support
domains:enable <app>                           # Enable VHOST support
domains:remove <app> <domain> [<domain> ...]   # Remove domains from app
domains:remove-global <domain> [<domain> ...]  # Remove global domain names
domains:report [<app>] [<flag>]                # Displays a domains report for one or more apps
domains:set <app> <domain> [<domain> ...]      # Set domains for app
domains:set-global <domain> [<domain> ...]     # Set global domain names
```

> Adding a domain before deploying an application will result in port mappings being set. This may cause issues for applications that use non-standard ports, as those will not be automatically detected. Please refer to the [proxy documentation](/docs/networking/proxy-management.md) for information as to how to reconfigure the mappings.

## Customizing hostnames

Applications typically have the following structure for their hostname:

```
scheme://subdomain.domain.tld
```

The `subdomain` is inferred from the pushed application name, while the `domain.tld` is set during initial dokku configuration. It can then be modified with `dokku domains:add-global` and `dokku domains:remove-global`. This value is used as a default TLD for all applications on a host.

If a FQDN such as `other.tld` is used as the application name, the global virtualhost will be ignored and the resulting vhost URL for that application will be `other.tld`. The exception to this rule being that if the FQDN has the same ending as the default vhost (such as `subdomain.domain.tld`), then the entire FQDN will be treated as a subdomain. The application will therefore be deployed at `subdomain.domain.tld.domain.tld`.

You can optionally override this in a plugin by implementing the `nginx-hostname` plugin trigger. For example, you can reverse the subdomain with the following sample `nginx-hostname` plugin trigger:

```shell
#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"; SUBDOMAIN="$2"; VHOST="$3"

NEW_SUBDOMAIN=`echo $SUBDOMAIN | rev`
echo "$NEW_SUBDOMAIN.$VHOST"
```

If the `nginx-hostname` plugin has no output, the normal hostname algorithm will be executed.

## Disabling VHOSTS

If desired, it is possible to disable vhosts with the domains plugin.

```shell
dokku domains:disable node-js-app
```

On subsequent deploys, the nginx virtualhost will be discarded. This is useful when deploying internal-facing services that should not be publicly routeable. As of 0.4.0, nginx will still be configured to proxy your app on some random high port. This allows internal services to maintain the same port between deployments. You may change this port by setting `DOKKU_PROXY_PORT` and/or `DOKKU_PROXY_SSL_PORT` (for services configured to use SSL.)


The domains plugin allows you to specify custom domains for applications. This plugin is aware of any ssl certificates that are imported via `certs:add`. Be aware that disabling domains (with `domains:disable`) will override any custom domains.

```shell
# where `node-js-app` is the name of your app

# add a domain to an app
dokku domains:add node-js-app dokku.me

# list custom domains for app
dokku domains node-js-app

# clear all custom domains for app
dokku domains:clear node-js-app

# remove a custom domain from app
dokku domains:remove node-js-app dokku.me

# set all custom domains for app
dokku domains:set node-js-app dokku.me dokku.org
```

## Displaying domains reports about an app

> New as of 0.8.1

You can get a report about the app's domains status using the `domains:report` command:

```shell
dokku domains:report
```

```
=====> node-js-app domains information
       Domains app enabled: true
       Domains app vhosts:  node-js-sample.dokku.org
       Domains global enabled: true
       Domains global vhosts: dokku.org
=====> python-sample domains information
       Domains app enabled: true
       Domains app vhosts:  python-sample.dokku.org
       Domains global enabled: true
       Domains global vhosts: dokku.org
=====> ruby-sample domains information
       Domains app enabled: true
       Domains app vhosts:  ruby-sample.dokku.org
       Domains global enabled: true
       Domains global vhosts: dokku.org
```

You can run the command for a specific app also.

```shell
dokku domains:report node-js-app
```

```
=====> node-js-app domains information
       Domains app enabled: true
       Domains app vhosts:  node-js-app.dokku.org
       Domains global enabled: true
       Domains global vhosts: dokku.org
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku domains:report node-js-app --domains-app-enabled
```

## Default site

By default, Dokku will route any received request with an unknown HOST header value to the lexicographically first site in the nginx config stack. If this is not the desired behavior, you may want to add the following configuration to the global nginx configuration.

Create the file at `/etc/nginx/conf.d/00-default-vhost.conf`:

```nginx
server {
    listen 80 default_server;
    server_name _;
    access_log off;
    return 410;
}

# To handle HTTPS requests, you can uncomment the following section.
#
# Please note that in order to let this work as expected, you need a valid
# SSL certificate for any domains being served. Browsers will show SSL
# errors in all other cases.
#
# Note that the key and certificate files in the below example need to
# be copied into /etc/nginx/ssl/ folder.
#
# server {
#     listen 443 ssl;
#     server_name _;
#     ssl_certificate /etc/nginx/ssl/cert.crt;
#     ssl_certificate_key /etc/nginx/ssl/cert.key;
#     access_log off;
#     return 410;
# }
```

Make sure to reload nginx after creating this file by running `service nginx reload`.

This will catch all unknown HOST header values and return a `410 Gone` response. You can replace the `return 410;` with `return 444;` which will cause nginx to not respond to requests that do not match known domains (connection refused).

The configuration file must be loaded before `/etc/nginx/conf.d/dokku.conf`, so it can not be arranged as a vhost in `/etc/nginx/sites-enabled` that is only processed afterwards.

Alternatively, you may push an app to your Dokku host with a name like "00-default". As long as it lists first in `ls /home/dokku/*/nginx.conf | head`, it will be used as the default nginx vhost.
