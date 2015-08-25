# SSL Certificates

Dokku supports SSL/TLS certificate inspection and CSR/Self-signed certificate generation via the `certs` plugin

```
certs:generate <app> DOMAIN                                 Generate a key and certificate signing request (and self-signed certificate)
certs:info <app>                                            Show certificate information for an ssl endpoint.
certs:remove <app>                                          Remove an SSL Endpoint from an app.
```

## Certificate generation

The `certs:generate` command will walk you through the correct `openssl` commands to create a key, csr and a self-signed cert for a given app/domain. We automatically put the self-signed cert in place as well as add the specified domain to the application configuration. If you decide to obtain a CA signed certficate, simply place it in `$DOKKU_ROOT/$APP/tls/server.crt` and run `dokku nginx:build-config $APP`.

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

### `certs:remove` command

The `certs:remove` command only works on app-specific certificates. It will `rm` the app-specific tls directory, rebuild the nginx configuration, and reload nginx.
