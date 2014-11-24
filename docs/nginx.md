# Nginx

Dokku uses nginx as it's server for routing requests to specific applications.

## TLS/SPDY support

Dokku provides easy TLS/SPDY support out of the box. This can be done app-by-app or for all subdomains at once. Note that whenever TLS support is enabled SPDY is also enabled.

### Per App

To enable TLS connection to to one of your applications, copy or symlink the `.crt`/`.pem` and `.key` files into the application's `/home/dokku/:app/tls` folder (create this folder if it doesn't exist) as `server.crt` and `server.key` respectively.

Redeployment of the application will be needed to apply TLS configuration. Once it is redeployed, the application will be accessible by `https://` (redirection from `http://` is applied as well).

### All Subdomains

To enable TLS connections for all your applications at once you will need a wildcard TLS certificate.

To enable TLS across all apps, copy or symlink the `.crt`/`.pem` and `.key` files into the  `/home/dokku/tls` folder (create this folder if it doesn't exist) as `server.crt` and `server.key` respectively. Then, enable the certificates by editing `/etc/nginx/conf.d/dokku.conf` and uncommenting these two lines (remove the #):

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

This archive should is expanded via `tar xvf`. It should contain `server.crt` and `server.key`.

## Disabling VHOSTS

If desired, it is possible to disable vhosts by setting the `NO_VHOST` environment variable:

```bash
dokku config:set myapp NO_VHOST=1
```

On subsequent deploys, the nginx virtualhost will be discarded. This is useful when deploying internal-facing services that should not be publicly routeable.

## Customizing hostnames

Applications typically have the following structure for their hostname:

```
scheme://subdomain.domain.tld
```

The `subdomain` is inferred from the pushed application name, while the `domain` is set during initial configuration in the `$DOKKU_ROOT/VHOST` file.

You can optionally override this in a plugin by implementing the `nginx-hostname` pluginhook. For example, you can reverse the subdomain with the following sample `nginx-hostname` pluginhook:

```bash
#!/usr/bin/env bash
set -eo pipefail; [[ $DOKKU_TRACE ]] && set -x

APP="$1"; SUBDOMAIN="$2"; VHOST="$3"

NEW_SUBDOMAIN=`echo $SUBDOMAIN | rev`
echo "$NEW_SUBDOMAIN.$VHOST"
```

If the `nginx-hostname` has no output, the normal hostname algorithm will be executed.
