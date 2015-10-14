# SSL Configuration

> New as of 0.4.0

Dokku supports SSL/TLS certificate inspection and CSR/Self-signed certificate generation via the `certs` plugin. Note that whenever SSL/TLS support is enabled SPDY is also enabled.

```
certs:add <app> CRT KEY                          Add an ssl endpoint to an app. Can also import from a tarball on stdin.
certs:generate <app> DOMAIN                      Generate a key and certificate signing request (and self-signed certificate)
certs:info <app>                                 Show certificate information for an ssl endpoint.
certs:remove <app>                               Remove an SSL Endpoint from an app.
certs:update <app> CRT KEY                       Update an SSL Endpoint on an app. Can also import from a tarball on stdin
```

## Per-application certificate management

Dokku provides built-in support for managing SSL certificates on a per-application basis. SSL is managed via nginx outside of application containers, and as such can be updated on-the-fly without rebuilding containers. At this time, applications only support a single SSL certificate at a time. To support multiple domains for a single application, wildcard certificate usage is encouraged.

### Certificate setting

The `certs:add` command can be used to push a `tar` containing a certificate `.crt` and `.key` file to a single application. The command should correctly handle cases where the `.crt` and `.key` are not named properly or are nested in a subdirectory of said `tar` file. You can import it as follows:

```shell
tar cvf cert-key.tar server.crt server.key
# replace APP with the name of your application
dokku certs:add < cert-key.tar
```

### Certificate generation

> Note: Using this method will create a self-signed certificate, which is only recommended for development or staging use, not production environments.

The `certs:generate` command will walk you through the correct `openssl` commands to create a key, csr and a self-signed cert for a given app/domain. We automatically put the self-signed cert in place as well as add the specified domain to the application configuration.

If you decide to obtain a CA signed certficate, you can import that certificate using the aformentioned `dokku certs:add` command.

### Certificate information

The `certs:info` command will simply inspect the install SSL cert and print out details. NOTE: The server-wide certificate will be inspect if installed and no app-specific certificate exists.

```
root@dokku:~/dokku# dokku certs:info node-js-app
-----> Fetching SSL Endpoint info for node-js-app...
-----> Certificate details:
=====> Common Name(s):
=====>    test.dokku.me
=====> Expires At: Aug 24 23:32:59 2016 GMT
=====> Issuer: C=US, ST=California, L=San Francisco, O=dokku.me, CN=test.dokku.me
=====> Starts At: Aug 25 23:32:59 2015 GMT
=====> Subject: C=US; ST=California; L=San Francisco; O=dokku.me; CN=test.dokku.me
=====> SSL certificate is self signed.
```

### Certificate removal

The `certs:remove` command only works on app-specific certificates. It will `rm` the app-specific tls directory, rebuild the nginx configuration, and reload nginx.

## Global Certification

Global certificate management is a manual process. To enable TLS connections for all your applications at once you will need a wildcard TLS certificate.

To enable TLS across all apps, you can run the following commands:

```shell
mkdir -p /home/dokku/tls
cp server.crt /home/dokku/tls/server.crt
cp server.key /home/dokku/tls/server.key
```

Next, you will want to enable the certificates by editing `/etc/nginx/conf.d/dokku.conf` and uncommenting these two lines (remove the `#`):

```
ssl_certificate /home/dokku/tls/server.crt;
ssl_certificate_key /home/dokku/tls/server.key;
```

The settings will take affect at the next deploy. If you would like to propagate the change to all apps immediately, you can also run the following command:

```shell
dokku ps:restartall
```

Once TLS is enabled, the application will be accessible by `https://` (redirection from `http://` is applied as well).

> Note: TLS will not be enabled unless the application's VHOST matches the certificate's name. (i.e. if you have a cert for `*.example.com` TLS won't be enabled for `something.example.org` or `example.net`)

## HSTS Header

The [HSTS header](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) is an HTTP header that can inform browsers that all requests to a given site should be made via HTTPS. dokku does not, by default, enable this header. It is thus left up to you, the user, to enable it for your site.

Beware that if you enable the header and a subsequent deploy of your application results in an HTTP deploy (for whatever reason), the way the header works means that a browser will not attempt to request the HTTP version of your site if the HTTPS version fails.
