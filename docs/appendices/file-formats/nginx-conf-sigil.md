# nginx.conf.sigil

The `nginx.conf.sigil` file is used to configure the nginx server for an application. The [default template](https://github.com/dokku/dokku/blob/master/plugins/nginx-vhosts/templates/nginx.conf.sigil) is in the nginx-vhosts plugin. Dokku uses a tool named [sigil](https://github.com/gliderlabs/sigil) to generate the nginx configuration based on the template provided.

## Validation

A custom `nginx.conf.sigil` is pre-validated at the start of every deploy, immediately after it is extracted from the source tree and before the build phase runs. Pre-validation renders the template via sigil with the same parameters used at deploy time, wraps the rendered config in a minimal `events`/`http` scaffold, and runs `nginx -t` against the result. The deploy is aborted if either the sigil render or the `nginx -t` check fails, so build work is not wasted on a syntactically invalid template. When no app listeners exist yet (typical for first deploys), pre-validation injects placeholders so the rendered config parses cleanly under `nginx -t`:

- `DOKKU_APP_WEB_LISTENERS` is set to `127.0.0.1:5000` so the legacy `upstream` block has a static server entry.
- `DOKKU_APP_WEB_LISTENER_HOST` is set to `127.0.0.1` so the resolver-based `set $dokku_upstream "...";` line renders to a syntactically valid value.

Pre-validation is skipped when the proxy type is not `nginx` or when `disable-custom-config` is set to `true` for the app.

The wrapper declares the same `user` directive as the host's main nginx config, exposed to the wrapper template as `$.NGINX_USER`. This matters because `nginx -t` runs as `root` and chowns every directory the config registers - in particular any `proxy_cache_path` - to whichever user the config resolves to, whether or not that directory already exists. Without a matching `user`, validating a template that declares a `proxy_cache_path` would hand the running server's live cache directory to nginx's compile-time default user for the length of the build, and every request served from or stored in that cache would fail with a permission error until the deploy's own nginx reload restored the ownership.

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
{{ if $.NGINX_USER }}user {{ $.NGINX_USER }};{{ end }}
events { worker_connections 768; }
http {
  access_log off;
  error_log /dev/null;
  include {{ $.NGINX_CONF }};
}
```

A custom wrapper that omits the `user` line still has one injected before `nginx -t` runs, so wrappers written against older Dokku releases keep working, but declaring it explicitly documents the behavior and keeps the template self-contained.

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

## Upstream Rendering: IP-based vs DNS-based

> [!IMPORTANT]
> New as of 0.39.0

The bundled `nginx.conf.sigil` supports two rendering paths for the upstream, and both sets of sigil variables are populated on every render. Vendored templates can use either path independently of the global `dns-resolver` setting.

### IP-based upstream (legacy path)

The classic form uses `DOKKU_APP_<PROC>_LISTENERS`: a space-separated `<ip>:<port>` list, one entry per scaled instance of the process type. The IPs come from `/home/dokku/<APP>/IP.<proc>.<index>` and the ports from `/home/dokku/<APP>/PORT.<proc>.<index>`. A typical use is an `upstream` block with one `server <ip>:<port>;` per instance:

```nginx
upstream {{ $.APP }}-{{ $upstream_port }} {
{{ range $listeners := $.DOKKU_APP_WEB_LISTENERS | split " " }}
{{ $listener_list := $listeners | split ":" }}
{{ $listener_ip := index $listener_list 0 }}
  server {{ $listener_ip }}:{{ $upstream_port }};{{ end }}
}
```

This path renders the IP that Dokku saw when the config was last built. After a host reboot, an `apt upgrade docker` without live-restore, or a single container restart, those IPs can become stale until `dokku proxy:build-config` runs again. The DNS-based path below addresses that staleness.

### DNS-based upstream (resolver path)

The new variables (`DOKKU_APP_<PROC>_LISTENER_HOST` + `NGINX_DNS_RESOLVER` + `NGINX_DNS_RESOLVER_TIMEOUT` + `NGINX_DNS_ZONE`) describe a single hostname per process type of the form `<app>.<proctype>.<dns-zone>` (for example, `myapp.web.docker`). The hostname is served by [coredns-docker](https://github.com/dokku/coredns-docker) (>= `0.8.0`), which watches Docker events and collapses every running instance of the process type onto one multi-A record set. Combined with an nginx `resolver` directive and a variable-based `proxy_pass`, the upstream re-resolves at request time honoring DNS TTL - which gives runtime IP refresh on container restarts and DNS round-robin across scaled instances.

The bundled template emits this snippet:

```nginx
{{ if $.NGINX_DNS_RESOLVER }}
resolver {{ $.NGINX_DNS_RESOLVER }} valid=10s ipv6=off;
resolver_timeout {{ $.NGINX_DNS_RESOLVER_TIMEOUT }};
set $dokku_upstream "{{ $.DOKKU_APP_WEB_LISTENER_HOST }}:{{ $upstream_port }}";
proxy_pass http://$dokku_upstream;
{{ else }}
proxy_pass http://{{ $.APP }}-{{ $upstream_port }};
{{ end }}
```

The `$dokku_upstream` variable name is intentionally namespaced (rather than `$upstream`) so vendored templates that already define `$upstream` will not collide if they adopt the resolver branch.

`NGINX_DNS_RESOLVER` is empty by default, `NGINX_DNS_RESOLVER_TIMEOUT` defaults to `5s`, and `NGINX_DNS_ZONE` defaults to `docker`; all three are configurable via `dokku nginx:set --global dns-resolver ...`, `dokku nginx:set --global dns-resolver-timeout ...`, and `dokku nginx:set --global dns-zone ...`. The bundled template only renders the resolver branch when `NGINX_DNS_RESOLVER` is set, for example to `127.0.0.1:1053`. Unsetting `dns-resolver` switches the bundled template back to the IP-based upstream. The short default `dns-resolver-timeout` keeps clients from waiting on nginx's built-in 30s timeout when coredns-docker is unreachable; nginx returns 502 after the configured interval instead.

### Choosing between the two

The IP-based path is the default for the bundled template, and the DNS-based path is used once `dns-resolver` is set. Vendored templates that have not been updated continue to use `DOKKU_APP_<PROC>_LISTENERS` and are unaffected by the `dns-resolver` setting. The `DOKKU_APP_<PROC>_LISTENER_HOST` and `NGINX_DNS_ZONE` variables are populated regardless of whether `dns-resolver` is set, so a vendored template can adopt the resolver path on its own schedule by switching its rendering snippet, even if other apps on the host are still using IP-based upstreams.
