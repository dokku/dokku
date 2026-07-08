# nginx.conf.sigil

The `nginx.conf.sigil` file is used to configure the nginx server for an application. The default template can be found [here](https://github.com/dokku/dokku/blob/master/plugins/nginx-vhosts/templates/nginx.conf.sigil). Dokku uses a tool named [sigil](https://github.com/gliderlabs/sigil) to generate the nginx configuration based on the template provided.

## Validation

A custom `nginx.conf.sigil` is pre-validated at the start of every deploy, immediately after it is extracted from the source tree and before the build phase runs. Pre-validation renders the template via sigil with the same parameters used at deploy time, wraps the rendered config in a minimal `events`/`http` scaffold, and runs `nginx -t` against the result. The deploy is aborted if either the sigil render or the `nginx -t` check fails, so build work is not wasted on a syntactically invalid template. When no app listeners exist yet (typical for first deploys), pre-validation injects a placeholder `127.0.0.1:5000` listener for `DOKKU_APP_WEB_LISTENERS` so that the rendered upstream block has a static server entry and `nginx -t` does not bail out on "host not found in upstream".

Pre-validation is skipped when the proxy type is not `nginx` or when `disable-custom-config` is set to `true` for the app.

### Custom nginx modules

Pre-validation runs `nginx -t` against a minimal wrapper config that does _not_ include the top-level `load_module` directives from the global `/etc/nginx/nginx.conf`. A `nginx.conf.sigil` that uses a directive provided by a dynamically loaded module - such as `image_filter`, provided by the [ngx_http_image_filter_module](https://nginx.org/en/docs/http/ngx_http_image_filter_module.html) - therefore fails pre-validation with an `unknown directive` error, even though `nginx -t` succeeds against the real server config where the module is loaded.

The `load_module` directive cannot be added to the app's `nginx.conf.sigil` to work around this, as that file is included inside the `http { }` block while `load_module` is only valid in nginx's top-level main context.

To make pre-validation aware of a module, override the wrapper template used for validation. The [`nginx-app-template-source`](/docs/development/plugin-triggers.md#nginx-app-template-source) trigger returns the path to the `sigil` template used to generate a given nginx configuration file, and its `validate-config` template type controls the pre-validation wrapper. Create a [custom plugin](/docs/development/plugin-creation.md) that implements the trigger and returns a `validate.conf.sigil` that adds the required `load_module` line at the top of the wrapper.

The trigger file (named `nginx-app-template-source` and marked executable):

```shell
#!/usr/bin/env bash

set -eo pipefail
[[ $DOKKU_TRACE ]] && set -x

APP="$1"
TEMPLATE_TYPE="$2"
if [[ "$TEMPLATE_TYPE" == "validate-config" ]]; then
  echo "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/validate.conf.sigil"
fi
```

The custom `validate.conf.sigil`, which is the [default wrapper](https://github.com/dokku/dokku/blob/master/plugins/nginx-vhosts/templates/validate.conf.sigil) with the required `load_module` line added at the top. Use the same `load_module` line that the host's global `/etc/nginx/nginx.conf` uses:

```
load_module modules/ngx_http_image_filter_module.so;
events { worker_connections 768; }
http {
  access_log off;
  error_log /dev/null;
  include {{ $.NGINX_CONF }};
}
```

The same override also governs the standalone `dokku nginx:validate-config` command, which renders the `validate-config` template as well.

## HTTP/2

nginx 1.25.1 deprecated the `http2` parameter on the `listen` directive in favor of a standalone `http2 on;` directive. Custom `nginx.conf.sigil` templates that hardcode `listen ... ssl http2;` will produce `nginx: [warn] the "listen ... http2" directive is deprecated` warnings when run against nginx 1.25.1 or newer.

Dokku exposes the `HTTP2_DIRECTIVE_SUPPORTED` template variable, set to `"true"` when the host nginx is 1.25.1 or newer, so a single template can render the correct syntax against either version. The default template uses this pattern:

```
{{ if eq $.HTTP2_DIRECTIVE_SUPPORTED "true" }}
listen      [{{ $.NGINX_BIND_ADDRESS_IP6 }}]:{{ $listen_port }} ssl;
listen      {{ if $.NGINX_BIND_ADDRESS_IP4 }}{{ $.NGINX_BIND_ADDRESS_IP4 }}:{{end}}{{ $listen_port }} ssl;
http2 on;
{{ else }}
listen      [{{ $.NGINX_BIND_ADDRESS_IP6 }}]:{{ $listen_port }} ssl http2;
listen      {{ if $.NGINX_BIND_ADDRESS_IP4 }}{{ $.NGINX_BIND_ADDRESS_IP4 }}:{{end}}{{ $listen_port }} ssl http2;
{{ end }}
```

Dokku logs a deprecation warning during deploys when a custom template still uses the `listen ... http2` form, so the offending template can be located via the deploy output.
