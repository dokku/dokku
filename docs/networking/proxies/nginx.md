# Nginx Proxy

Dokku uses nginx as its server for routing requests to specific applications. By default, access and error logs are written for each app to `/var/log/nginx/${APP}-access.log` and `/var/log/nginx/${APP}-error.log` respectively

```
nginx:access-logs <app> [-t]             # Show the nginx access logs for an application (-t follows)
nginx:error-logs <app> [-t]              # Show the nginx error logs for an application (-t follows)
nginx:report [<app>] [<flag>]            # Displays a nginx report for one or more apps
nginx:set <app> <property> (<value>)     # Set or clear an nginx property for an app
nginx:show-config <app>                  # Display app nginx config
nginx:start                              # Starts the nginx server
nginx:stop                               # Stops the nginx server
nginx:validate-config [<app>] [--clean]  # Validates and optionally cleans up invalid nginx configurations
```

## Usage

> [!WARNING]
> As using multiple proxy plugins on a single Dokku installation can lead to issues routing requests to apps, doing so should be avoided.

### Request Proxying

By default, the `web` process is the only process proxied by the nginx proxy implementation. Proxying to other process types may be handled by a custom `nginx.conf.sigil` file, as generally described [below](/docs/networking/proxies/nginx.md#customizing-the-nginx-configuration)

Nginx will proxy the requests in a [round-robin balancing fashion](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream) to the different deployed (scaled) containers running the `web` proctype. This way, the host's resources can be fully leveraged for single-threaded applications (e.g. `dokku ps:scale node-js-app web=4` on a 4-core machine).

> [!NOTE]
> Due to how the plugin is implemented, if an app successfully starts up `web` containers but fails to deploy some other containers, nginx may eventually stop routing requests. Users should revert their code in these cases, or manually trigger `dokku proxy:build-config $APP` in order to ensure requests route to the new web containers.

### Starting nginx

> [!IMPORTANT]
> New as of 0.28.0

The nginx server can be started via `nginx:start`.

```shell
dokku nginx:start
```

### Stopping nginx

> [!IMPORTANT]
> New as of 0.28.0

The nginx server can be stopped via `nginx:stop`.

```shell
dokku nginx:stop
```

### Checking access logs

> [!NOTE]
> Changing this value globally or on a per-app basis will require rebuilding the nginx config via the `proxy:build-config` command.

You may check nginx access logs via the `nginx:access-logs` command. This assumes that app access logs are being stored in `/var/log/nginx/$APP-access.log`, as is the default in the generated `nginx.conf`.

```shell
dokku nginx:access-logs node-js-app
```

You may also follow the logs by specifying the `-t` flag.

```shell
dokku nginx:access-logs node-js-app -t
```

### Checking error logs

You may check nginx error logs via the `nginx:error-logs` command. This assumes that app error logs are being stored in `/var/log/nginx/$APP-error.log`, as is the default in the generated `nginx.conf`.

```shell
dokku nginx:error-logs node-js-app
```

You may also follow the logs by specifying the `-t` flag.

```shell
dokku nginx:error-logs node-js-app -t
```

### Showing the nginx config

For debugging purposes, it may be useful to show the nginx config. This can be achieved via the `nginx:show-config` command.

```shell
dokku nginx:show-config node-js-app
```

### Validating nginx configs

It may be desired to validate an nginx config outside of the deployment process. To do so, run the `nginx:validate-config` command. With no arguments, this will validate all app nginx configs, one at a time. A minimal wrapper nginx config is generated for each app's nginx config, upon which `nginx -t` will be run.

```shell
dokku nginx:validate-config
```

As app nginx configs are actually executed within a shared context, it is possible for an individual config to be invalid when being validated standalone but _also_ be valid within the global server context. As such, the exit code for the `nginx:validate-config` command is the exit code of `nginx -t` against the server's real nginx config.

The `nginx:validate-config` command also takes an optional `--clean` flag. If specified, invalid nginx configs will be removed.

> [!WARNING]
> Invalid app nginx config's will be removed _even if_ the config is valid in the global server context.

```shell
dokku nginx:validate-config --clean
```

The `--clean` flag may also be specified for a given app:

```shell
dokku nginx:validate-config node-js-app --clean
```

### Custom Error Pages

By default, Dokku provides custom error pages for the following three categories of errors:

- 4xx: For all non-404 errors with a 4xx response code.
- 404: For "404 Not Found" errors.
- 5xx: For all 5xx error responses

These are provided as an alternative to the generic Nginx error page, are shared for _all_ applications, and their contents are located on disk at `/var/lib/dokku/data/nginx-vhosts/dokku-errors`. To customize them for a specific app, create a custom `nginx.conf.sigil` as described above and change the paths to point elsewhere.

### Default site

By default, Dokku will route any received request with an unknown HOST header value to the lexicographically first site in the nginx config stack. This means that accessing the dokku server via its IP address or a bogus domain name may return a seemingly random website.

> [!WARNING]
> Some versions of Nginx may create a default site when installed. This site is simply a static page which says "Welcome to Nginx", and if this default site is enabled, Nginx will not route any requests with an unknown HOST header to Dokku. If you want Dokku to receive all requests, run the following commands:
>
> ```
> rm /etc/nginx/sites-enabled/default
> dokku nginx:stop
> dokku nginx:start
> ```

If services should only be accessed via their domain name, you may want to disable the default site by adding the following configuration to the global nginx configuration.

Create the file at `/etc/nginx/conf.d/00-default-vhost.conf`:

```nginx
server {
    listen 80 default_server;
    listen [::]:80 default_server;

    # If services hosted by dokku are available via HTTPS, it is recommended
    # to also uncomment the following section.
    #
    # Please note that in order to let this work, you need an SSL certificate. However
    # it does not need to be valid. Users of Debian-based distributions can install the
    # `ssl-cert` package with `sudo apt install ssl-cert` to automatically generate
    # a self-signed certificate that is stored at `/etc/ssl/certs/ssl-cert-snakeoil.pem`.
    #
    #listen 443 ssl;
    #listen [::]:443 ssl;
    #ssl_certificate /etc/ssl/certs/ssl-cert-snakeoil.pem;
    #ssl_certificate_key /etc/ssl/private/ssl-cert-snakeoil.key;

    server_name _;
    access_log off;
    return 444;
}
```

Make sure to reload nginx after creating this file by running `systemctl reload nginx.service`.

This will catch all unknown HOST header values and close the connection without responding. You can replace the `return 444;` with `return 410;` which will cause nginx to respond with an error page.

The configuration file must be loaded before `/etc/nginx/conf.d/dokku.conf`, so it can not be arranged as a vhost in `/etc/nginx/sites-enabled` that is only processed afterwards.

Alternatively, you may push an app to your Dokku host with a name like "00-default". As long as it lists first in `ls /home/dokku/*/nginx.conf | head`, it will be used as the default nginx vhost.

### Customizing the nginx configuration

> [!IMPORTANT]
> New as of 0.5.0

Dokku uses a templating library by the name of [sigil](https://github.com/gliderlabs/sigil) to generate nginx configuration for each app. This may be overridden by committing the [default configuration template](https://github.com/dokku/dokku/blob/master/plugins/nginx-vhosts/templates/nginx.conf.sigil) to a file named `nginx.conf.sigil`.

The `nginx.conf.sigil` is expected to be found in a specific directory, depending on the deploy approach:

- The `WORKDIR` of the Docker image for deploys resulting from `git:from-image` and `git:load-image` commands.
- The root of the source code tree for all other deploys (git push, `git:from-archive`, `git:sync`).

Sometimes it may be desirable to set a different path for a given app, e.g. when deploying from a monorepo. This can be done via the `nginx-conf-sigil-path` property:

```shell
dokku nginx:set node-js-app nginx-conf-sigil-path .dokku/nginx.conf.sigil
```

The value is the path to the desired file _relative_ to the base search directory, and will never be treated as absolute paths in any context. If that file does not exist within the repository, Dokku will continue the build process as if the repository has no `nginx.conf.sigil`.

The default value may be set by passing an empty value for the option:

```shell
dokku nginx:set node-js-app nginx-conf-sigil-path
```

The `nginx-conf-sigil-path` property can also be set globally. The global default is `nginx.conf.sigil`, and the global value is used when no app-specific value is set.

```shell
dokku nginx:set --global nginx-conf-sigil-path nginx.conf.sigil
```

The default value may be set by passing an empty value for the option.

```shell
dokku nginx:set --global nginx-conf-sigil-path
```

> The [default template](https://github.com/dokku/dokku/blob/master/plugins/nginx-vhosts/templates/nginx.conf.sigil) may change with new releases of Dokku. Please refer to the appropriate template file version for your Dokku version, and make sure to look out for changes when you upgrade.

#### Disabling custom nginx config

> [!NOTE]
> Changing this value globally or on a per-app basis will require rebuilding the nginx config via the `proxy:build-config` command.

While enabled by default, using a custom nginx config can be disabled via `nginx:set`. This may be useful in cases where you do not want to allow users to override any higher-level customization of app nginx config.

```shell
# enable fetching custom config (default)
dokku nginx:set node-js-app disable-custom-config false

# disable fetching custom config
dokku nginx:set node-js-app disable-custom-config true
```

Unsetting this value is the same as enabling custom nginx config usage.

#### Available template variables

```
{{ .APP }}                          Application name
{{ .APP_SSL_PATH }}                 Path to SSL certificate and key
{{ .DOKKU_ROOT }}                   Global Dokku root directory (ex: app dir would be `{{ .DOKKU_ROOT }}/{{ .APP }}`)
{{ .PROXY_PORT }}                   Non-SSL nginx listener port (same as `DOKKU_PROXY_PORT` config var)
{{ .PROXY_SSL_PORT }}               SSL nginx listener port (same as `DOKKU_PROXY_SSL_PORT` config var)
{{ .NOSSL_SERVER_NAME }}            List of non-SSL VHOSTS
{{ .PROXY_PORT_MAP }}               List of port mappings (same as the `map` ports property)
{{ .PROXY_UPSTREAM_PORTS }}         List of configured upstream ports (derived from the `map` ports property)
{{ .SSL_INUSE }}                    Boolean set when an app is SSL-enabled
{{ .SSL_SERVER_NAME }}              List of SSL VHOSTS
```

Finally, each process type has it's network listeners - a list of IP:PORT pairs for the respective app containers - exposed via an `.DOKKU_APP_${PROCESS_TYPE}_LISTENERS` variable - the `PROCESS_TYPE` will be upper-cased with hyphens transformed into underscores. Users can use the new variables to expose non-web processes via the nginx proxy.

> [!NOTE]
> Application environment variables are available for use in custom templates. To do so, use the form of `{{ var "FOO" }}` to access a variable named `FOO`.

#### Customizing via configuration files included by the default templates

The default nginx.conf template will include everything from your apps `nginx.conf.d/` subdirectory in the main `server {}` block (see above):

```
include {{ .DOKKU_ROOT }}/{{ .APP }}/nginx.conf.d/*.conf;
```

That means you can put additional configuration in separate files. To increase the client request header timeout, the following can be performed:

```shell
mkdir /home/dokku/node-js-app/nginx.conf.d/
echo 'client_header_timeout 50s;' > /home/dokku/node-js-app/nginx.conf.d/timeout.conf
chown dokku:dokku /home/dokku/node-js-app/nginx.conf.d/upload.conf
service nginx reload
```

The example above uses additional configuration files directly on the Dokku host. Unlike the `nginx.conf.sigil` file, these additional files will not be copied over from your application repo, and thus need to be placed in the `/home/dokku/node-js-app/nginx.conf.d/` directory manually.

For PHP Buildpack users, you will also need to provide a `Procfile` and an accompanying `nginx.conf` file to customize the nginx config _within_ the container. The following are example contents for your `Procfile`

```
web: vendor/bin/heroku-php-nginx -C nginx.conf -i php.ini php/
```

Your `nginx.conf` file - not to be confused with Dokku's `nginx.conf.sigil` - would also need to be configured as shown in this example:

```
client_header_timeout 50s;
location / {
    index index.php;
    try_files $uri $uri/ /index.php$is_args$args;
}
```

Please adjust the `Procfile` and `nginx.conf` file as appropriate.

### Setting Properties for the nginx config

The nginx plugin exposes a variety of properties configurable via the `nginx:set` command. The properties are used to configure the generated `nginx.conf` file from the `nginx.conf.sigil` template. The value precedence is app-specific, then global, and finally the Dokku default.

The nginx:set command takes an app name or the `--global` flag.

```shell
# set a property for the node-js-app
dokku nginx:set node-js-app property-name some-value

# set a property globally
dokku nginx:set --global property-name some-value
```

Additionally, setting an empty value will result in reverting the value back to it's default. For app-specific values, this means that Dokku will revert to the globally specified (or global default) value.

```shell
# default the value back to the global value for node-js-app
dokku nginx:set node-js-app property-name

# use the dokku default as the global value
dokku nginx:set --global property-name
```

Changing these value globally or on a per-app basis will require rebuilding the nginx config via the `proxy:build-config` command.

> [!WARNING]
> Validation is not performed against the values, and they are used as is within Dokku.

| Property                  | Default                               | Type    | Explanation                                                                         |
| --------------------------|---------------------------------------|---------|-------------------------------------------------------------------------------------|
| access-log-format         | empty string                          | string  | Name of custom log format to use (log format must be specified elsewhere)           |
| access-log-path           | `${NGINX_LOG_ROOT}/${APP}-access.log` | string  | Log path for nginx access logs (set to `off` to disable)                            |
| bind-address-ipv4         | `0.0.0.0`                             | string  | Default IPv4 address to bind to                                                     |
| bind-address-ipv6         | `[::]`                                | string  | Default IPv6 address to bind to                                                     |
| client-max-body-size      | `1m`                                  | string  | Size (with units) of client request body (usually modified for file uploads)        |
| client-body-timeout       | `60s`                                 | string  | Timeout (with units) for reading the client request body                            |
| client-header-timeout     | `60s`                                 | string  | Timeout (with units) for reading the client request headers                         |
| error-log-path            | `${NGINX_LOG_ROOT}/${APP}-error.log`  | string  | Log path for nginx error logs (set to `off` to disable)                             |
| hsts                      | `true`                                | boolean | Enables or disables HSTS for your application                                       |
| hsts-include-subdomains   | `true`                                | boolean | Forces the browser to apply the HSTS policy to all app subdomains                   |
| hsts-max-age              | `15724800`                            | integer | Time in seconds to cache HSTS configuration                                         |
| hsts-preload              | `false`                               | boolean | Tells the browser to include the domain in their HSTS preload lists                 |
| keepalive-timeout         | `75s`                                 | string  | Timeout (with units) during which a keep-alive client connection will stay open on the server side |
| lingering-timeout         | `5s`                                  | string  | Timeout (with units) is the maximum waiting time for more client data to arrive     |
| nginx-conf-sigil-path     | `nginx.conf.sigil`                    | string  | Path in the repository to the `nginx.conf.sigil` file                               |
| proxy-buffer-size         | `8k` (# is os pagesize)               | string  | Size of the buffer used for reading the first part of proxied server response       |
| proxy-buffering           | `on`                                  | string  | Enables or disables buffering of responses from the proxied server                  |
| proxy-buffers             | `8 8k`                                | string  | Number and size of the buffers used for reading the proxied server response, for a single connection |
| proxy-busy-buffers-size   | `16k`                                 | string  | Limits the total size of buffers that can be busy sending a response to the client while the response is not yet fully read. |
| proxy-connect-timeout     | `60s`                                 | string  | Timeout (with units) for establishing a connection to your backend server           |
| proxy-read-timeout        | `60s`                                 | string  | Timeout (with units) for reading response from your backend server                  |
| proxy-send-timeout        | `60s`                                 | string  | Timeout (with units) for transmitting a request to your backend server              |
| send-timeout              | `60s`                                 | string  | Timeout (with units) for transmitting a response to your the client                 |
| underscore-in-headers     | `off`                                 | string  | Enables or disables the use of underscores in client request header fields.         |
| x-forwarded-for-value     | `$remote_addr`                        | string  | Used for specifying the header value to set for the `X-Forwarded-For` header        |
| x-forwarded-port-value    | `$server_port`                        | string  | Used for specifying the header value to set for the `X-Forwarded-Port` header       |
| x-forwarded-proto-value   | `$scheme`                             | string  | Used for specifying the header value to set for the `X-Forwarded-Proto` header      |
| x-forwarded-ssl           | empty string                          | string  | Less commonly used alternative to `X-Forwarded-Proto` (valid values: `on` or `off`) |

#### Binding to specific addresses

> [!NOTE]
> Users with apps that contain a custom `nginx.conf.sigil` file will need to modify the files to respect the new `NGINX_BIND_ADDRESS_IPV4` and `NGINX_BIND_ADDRESS_IPV6` variables.

Properties:

- `bind-address-ipv4`
- `bind-address-ipv6`

This is useful in cases where the proxying should be internal to a network or if there are multiple network interfaces that should respond with different content.

#### HSTS Header

> [!WARNING]
> if you enable the header and a subsequent deploy of your application results in an HTTP deploy (for whatever reason), the way the header works means that a browser will not attempt to request the HTTP version of your site if the HTTPS version fails until the max-age is reached.

Properties:

- `hsts`
- `hsts-include-subdomains`
- `hsts-max-age`
- `hsts-preload`

If SSL certificates are present, HSTS will be automatically enabled.

#### Running behind another proxy — configuring `X-Forwarded-*` headers

> [!WARNING]
> These values should only be modified if there is an intermediate Load balancer or CDN between the user and the Dokku server hosting your application.

Properties:

- `x-forwarded-for-value`
- `x-forwarded-port-value`
- `x-forwarded-proto-value`
- `x-forwarded-ssl`

Dokku's default Nginx configuration passes the de-facto standard HTTP headers [`X-Forwarded-For`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For), [`X-Forwarded-Proto`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Proto), and `X-Forwarded-Port` to your application.
These headers indicate the IP address of the original client making the request, the protocol of the original request (HTTP or HTTPS), and the port number of the original request, respectively.

If you have another HTTP proxy sitting in between the end user and your server (for example, a load balancer, or a CDN), then the values of these headers will contain information about (e.g. the IP address of) the closest proxy, and not the end user.

To fix this, assuming that the other proxy also passes `X-Forwarded-*` headers, which in turn contain information about the end user, you can tell Nginx include those values in the `X-Forwarded-*` headers that it sends to your application. You can do this via `nginx:set`, like so:

```shell
dokku nginx:set node-js-app x-forwarded-for-value '$http_x_forwarded_for'
dokku nginx:set node-js-app x-forwarded-port-value '$http_x_forwarded_port'
dokku nginx:set node-js-app x-forwarded-proto-value '$http_x_forwarded_proto'
```

However, note that you should only do this if:

1. Requests to your website always go through a trusted proxy.
2. That proxy is configured to send the aforementioned `X-Forwarded-*` headers.

Otherwise, if it's possible for clients to make HTTP requests directly against your server, bypassing the other proxy, or if the other proxy is not configured to set these headers, then a client can basically pass any arbitrary values for these headers (which your app then presumably reads) and thereby fake an IP address, for example.

There's also the `X-Forwarded-Ssl` header which a less common alternative to `X-Forwarded-Proto` — and because of that, isn't included in Dokku's default Nginx configuration. It can be turned `on` if need be:

```shell
# force-setting value to `on`
dokku nginx:set node-js-app x-forwarded-ssl on
```

#### Changing log path

> [!WARNING]
> The defaults should not be changed without verifying that the paths will be writeable by nginx.

Properties:

- `access-log-path`
- `error-log-path`

These setting can be useful for enabling or disabling logging by setting the values to `off`.

```shell
dokku nginx:set node-js-app access-log-path off
dokku nginx:set node-js-app error-log-path off
```

#### Changing log format

Properties:

- `acccess-log-format`

Prior to changing the log-format, log formats should be specified at a file such as `/etc/nginx/conf.d/00-log-formats.conf`. This will ensure they are available within your app's nginx context. For instance, the following may be added to the above file. It only needs to be specified once to be used for all apps.

```nginx
# /etc/nginx/conf.d/00-log-formats.conf
# escape=json was added in nginx 1.11.8
log_format json_combined escape=json
  '{'
    '"time_local":"$time_local",'
    '"remote_addr":"$remote_addr",'
    '"remote_user":"$remote_user",'
    '"request":"$request",'
    '"status":"$status",'
    '"body_bytes_sent":"$body_bytes_sent",'
    '"request_time":"$request_time",'
    '"http_referrer":"$http_referer",'
    '"http_user_agent":"$http_user_agent"'
  '}';
```

#### Specifying a read timeout

> [!NOTE]
> All numeric values _must_ have a trailing time value specified (`s` for seconds, `m` for minutes).

Properties:

- `proxy-read-timeout`

#### Specifying a custom client_max_body_size

> [!NOTE]
> All numerical values _must_ have a trailing size unit specified (`k` for kilobytes, `m` for megabytes).

Properties:

- `client-max-body-size`

This property is commonly used to increase the max file upload size.

Changing this value when using the PHP buildpack (or any other buildpack that uses an intermediary server) will require changing the value in the server config shipped with that buildpack. Consult your buildpack documentation for further details.

## Other

### Domains plugin

See the [domain configuration documentation](/docs/configuration/domains.md) for more information on how to configure domains for your app.

### Customizing hostnames

See the [customizing hostnames documentation](/docs/configuration/domains.md#customizing-hostnames) for more information on how to configure domains for your app.

### Disabling VHOSTS

See the [disabling vhosts documentation](/docs/configuration/domains.md#disabling-vhosts) for more information on how to disable domain usage for your app.

### SSL Configuration

See the [ssl documentation](/docs/configuration/ssl.md) for more information on how to configure SSL certificates for your application.

### Disabling Nginx

See the [proxy documentation](/docs/networking/proxy-management.md) for more information on how to disable nginx as the proxy implementation for your app.

### Managing Proxy Port mappings

See the [ports documentation](/docs/networking/port-management.md) for more information on how to manage ports proxied for your app.

### Regenerating nginx config

See the [proxy documentation](/docs/networking/proxy-management.md#regenerating-proxy-config) for more information on how to rebuild the nginx proxy configuration for your app.
