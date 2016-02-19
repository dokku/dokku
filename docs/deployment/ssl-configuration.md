# SSL Configuration

> New as of 0.4.0

Dokku supports SSL/TLS certificate inspection and CSR/Self-signed certificate generation via the `certs` plugin. Note that whenever SSL/TLS support is enabled SPDY is also enabled.

```
certs:add <app> CRT KEY                          Add an ssl endpoint to an app. Can also import from a tarball on stdin.
certs:generate <app> DOMAIN                      Generate a key and certificate signing request (and self-signed certificate)
certs:info <app>                                 Show certificate information for an ssl endpoint.
certs:remove <app>                               Remove an SSL Endpoint from an app.
certs:update <app> CRT KEY                       Update an SSL Endpoint on an app. Can also import from a tarball on stdin

# for 0.3.x
dokku nginx:import-ssl <app> < certs.tar
```

## Per-application certificate management

Dokku provides built-in support for managing SSL certificates on a per-application basis. SSL is managed via nginx outside of application containers, and as such can be updated on-the-fly without rebuilding containers. At this time, applications only support a single SSL certificate at a time. To support multiple domains for a single application, wildcard certificate usage is encouraged.

### Certificate setting

The `certs:add` command can be used to push a `tar` containing a certificate `.crt` and `.key` file to a single application. The command should correctly handle cases where the `.crt` and `.key` are not named properly or are nested in a subdirectory of said `tar` file. You can import it as follows:

```shell
tar cvf cert-key.tar server.crt server.key
# replace APP with the name of your application
dokku certs:add <app> < cert-key.tar
```

> Note: If your `.crt` file came alongside a `.ca-bundle`, you'll want to concatenate those into a single `.crt` file before adding it to the `.tar`.

```shell
cat yourdomain_com.crt yourdomain_com.ca-bundle > server.crt
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

## Running behind a load balancer

Your application has access to the HTTP headers `X-Forwarded-Proto`, `X-Forwarded-For` and `X-Forwarded-Port`. These headers indicate the protocol of the original request (HTTP or HTTPS), the port number, and the IP address of the client making the request, respectively. The default configuration is for Nginx to set these headers.

If your server runs behind an HTTP/S load balancer, then Nginx will see all requests as coming from the load balancer. If your load balancer sets the `X-Forwarded-` headers, you can tell Nginx to pass these headers from load balancer to your application by using the following [nginx custom template](/dokku/nginx/#customizing-the-nginx-configuration)

```shell
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
    proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
    proxy_set_header X-Forwarded-For $http_x_forwarded_for;
    proxy_set_header X-Forwarded-Port $http_x_forwarded_port;
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

Only use this option if:
1. All requests are terminated at the load balancer, and forwarded to Nginx
2. The load balancer is configured to send the `X-Forwarded-` headers (this may be off by default)

If it's possible to make HTTP/S requests directly to Nginx, bypassing the load balancer, or if the load balancer is not configured to set these headers, then it becomes possible for a client to set these headers to arbitrary values.

This could result in security issue, for example, if your application looks at the value of the `X-Forwarded-Proto` to determine if the request was made over HTTPS.
