# SSL Configuration

> New as of 0.4.0

Dokku supports SSL/TLS certificate inspection and CSR/Self-signed certificate generation via the `certs` plugin. Note that whenever SSL/TLS support is enabled SPDY is also enabled.

```
certs:add <app> CRT KEY                  # Add an ssl endpoint to an app. Can also import from a tarball on stdin.
certs:generate <app> DOMAIN              # Generate a key and certificate signing request (and self-signed certificate)
certs:remove <app>                       # Remove an SSL Endpoint from an app.
certs:report [<app>] [<flag>]            # Displays an ssl report for one or more apps
certs:update <app> CRT KEY               # Update an SSL Endpoint on an app. Can also import from a tarball on stdin
```

```shell
# for 0.3.x
dokku nginx:import-ssl <app> < certs.tar
```

> Adding an ssl certificate before deploying an application will result in port mappings being updated. This may cause issues for applications that use non-standard ports, as those may not be automatically detected. Please refer to the [proxy documentation](/docs/advanced-usage/proxy-management.md) for information as to how to reconfigure the mappings.

## Per-application certificate management

Dokku provides built-in support for managing SSL certificates on a per-application basis. SSL is managed via nginx outside of application containers, and as such can be updated on-the-fly without rebuilding containers. At this time, applications only support a single SSL certificate at a time. To support multiple domains for a single application, wildcard certificate usage is encouraged.

### Certificate setting

The `certs:add` command can be used to push a `tar` containing a certificate `.crt` and `.key` file to a single application. The command should correctly handle cases where the `.crt` and `.key` are not named properly or are nested in a subdirectory of said `tar` file. You can import it as follows:

```shell
tar cvf cert-key.tar server.crt server.key
dokku certs:add node-js-app < cert-key.tar
```

> Note: If your `.crt` file came alongside a `.ca-bundle`, you'll want to concatenate those into a single `.crt` file before adding it to the `.tar`.

```shell
cat yourdomain_com.crt yourdomain_com.ca-bundle > server.crt
```

#### SSL and Multiple Domains

When an SSL certificate is associated to an application, the certificate will be associated with *all* domains currently associated with said application. Your certificate _should_ be associated with all of those domains, otherwise accessing the application will result in SSL errors. If you wish to remove one of the domains from the application, refer to the [domain configuration documentation](/docs/configuration/domains.md).

Note that with the default nginx template, requests will be redirected to the `https` version of the domain. If this is not the desired state of request resolution, you may customize the nginx template in use. For more details, see the [nginx documentation](/docs/configuration/nginx.md).

### Certificate generation

> Note: Using this method will create a self-signed certificate, which is only recommended for development or staging use, not production environments.

The `certs:generate` command will walk you through the correct `openssl` commands to create a key, csr and a self-signed cert for a given app/domain. We automatically put the self-signed cert in place as well as add the specified domain to the application configuration.

If you decide to obtain a CA signed certificate, you can import that certificate using the aforementioned `dokku certs:add` command.

### Certificate removal

The `certs:remove` command only works on app-specific certificates. It will `rm` the app-specific tls directory, rebuild the nginx configuration, and reload nginx.

### Displaying certificate reports for an app

> New as of 0.8.1

You can get a report about the apps ssl status using the `certs:report` command:

```shell
dokku certs:report
```

```
=====> node-js-app
       Ssl dir:             /home/dokku/node-js-app/tls
       Ssl enabled:         true
       Ssl hostnames:       *.node-js-app.org node-js-app.org
       Ssl expires at:      Oct  5 23:59:59 2019 GMT
       Ssl issuer:          C=GB, ST=Greater Manchester, L=Salford, O=COMODO CA Limited, CN=COMODO RSA Domain Validation Secure Server CA
       Ssl starts at:       Oct  5 00:00:00 2016 GMT
       Ssl subject:         OU=Domain Control Validated; OU=PositiveSSL Wildcard; CN=*.node-js-app.org
       Ssl verified:        self signed.
=====> python-app
       Ssl dir:             /home/dokku/python-app/tls
       Ssl enabled:         false
       Ssl hostnames:
       Ssl expires at:
       Ssl issuer:
       Ssl starts at:
       Ssl subject:
       Ssl verified:
```

You can run the command for a specific app also.

```shell
dokku certs:report node-js-app
```

```
=====> node-js-app ssl information
       Ssl dir:             /home/dokku/node-js-app/tls
       Ssl enabled:         true
       Ssl hostnames:       *.dokku.org dokku.org
       Ssl expires at:      Oct  5 23:59:59 2019 GMT
       Ssl issuer:          C=GB, ST=Greater Manchester, L=Salford, O=COMODO CA Limited, CN=COMODO RSA Domain Validation Secure Server CA
       Ssl starts at:       Oct  5 00:00:00 2016 GMT
       Ssl subject:         OU=Domain Control Validated; OU=PositiveSSL Wildcard; CN=*.dokku.org
       Ssl verified:        self signed.
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku certs:report node-js-app --ssl-enabled
```

## HSTS Header

The [HSTS header](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) is an HTTP header that can inform browsers that all requests to a given site should be made via HTTPS. Dokku does not enables this header by default

See the [NGINX HSTS documentation](/docs/configuration/nginx.md#hsts-header) for more information.

## HTTP/2 support

Certain versions of nginx have bugs that prevent [HTTP/2](https://nginx.org/en/docs/http/ngx_http_v2_module.html) from properly responding to all clients, thus causing applications to be unavailable. For HTTP/2 to be enabled in your applications' nginx configs, you need to have installed nginx 1.11.5 or higher. See [issue 2435](https://github.com/dokku/dokku/issues/2435) for more details.

## Running behind a load balancer

Your application has access to the HTTP headers `X-Forwarded-Proto`, `X-Forwarded-Port` and `X-Forwarded-For`. These headers indicate the protocol of the original request (HTTP or HTTPS), the port number, and the IP address of the client making the request, respectively. The default configuration is for Nginx to set these headers.

If your server runs behind an HTTP(S) load balancer, then Nginx will see all requests as coming from the load balancer. If your load balancer sets the `X-Forwarded-` headers, you can tell Nginx to pass these headers from load balancer to your application by using a [custom nginx template](/docs/configuration/nginx.md#customizing-the-nginx-configuration). The following is a simple example of how to do so.

```go
server {
  listen      [::]:{{ .PROXY_PORT }};
  listen      {{ .PROXY_PORT }};
  server_name {{ .NOSSL_SERVER_NAME }};
  access_log  /var/log/nginx/{{ .APP }}-access.log;
  error_log   /var/log/nginx/{{ .APP }}-error.log;

  location    / {
    proxy_pass  http://{{ .APP }};
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $http_connection;
    proxy_set_header Host $http_host;
    proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
    proxy_set_header X-Forwarded-For $http_x_forwarded_for;
    proxy_set_header X-Forwarded-Port $http_x_forwarded_port;
    proxy_set_header X-Request-Start $msec;
  }
  include {{ .DOKKU_ROOT }}/{{ .APP }}/nginx.conf.d/*.conf;
}

upstream {{ .APP }} {
{{ range .DOKKU_APP_WEB_LISTENERS | split " " }}
  server {{ . }};
{{ end }}
}
```

Only use this option if:
1. All requests are terminated at the load balancer, and forwarded to Nginx
2. The load balancer is configured to send the `X-Forwarded-` headers (this may be off by default)

If it's possible to make HTTP(S) requests directly to Nginx, bypassing the load balancer, or if the load balancer is not configured to set these headers, then it becomes possible for a client to set these headers to arbitrary values.

This could result in security issue, for example, if your application looks at the value of the `X-Forwarded-Proto` to determine if the request was made over HTTPS.

### SSL Port Exposure

When your app is served from port `80` then the `/home/dokku/APP/nginx.conf` file will automatically be updated to instruct nginx to respond to ssl on port 443 as a new cert is added.  If your app uses a non-standard port (perhaps you have a dockerfile deploy exposing port `99999`) you may need to manually expose an ssl port via `dokku proxy:ports-add <APP> https:443:99999`.
