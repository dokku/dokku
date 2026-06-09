# nginx.conf.sigil

The `nginx.conf.sigil` file is used to configure the nginx server for an application. The default template can be found [here](https://github.com/dokku/dokku/blob/master/plugins/nginx-vhosts/templates/nginx.conf.sigil). Dokku uses a tool named [sigil](https://github.com/gliderlabs/sigil) to generate the nginx configuration based on the template provided.

## Validation

A custom `nginx.conf.sigil` is pre-validated at the start of every deploy, immediately after it is extracted from the source tree and before the build phase runs. Pre-validation renders the template via sigil with the same parameters used at deploy time, wraps the rendered config in a minimal `events`/`http` scaffold, and runs `nginx -t` against the result. The deploy is aborted if either the sigil render or the `nginx -t` check fails, so build work is not wasted on a syntactically invalid template. When no app listeners exist yet (typical for first deploys), pre-validation injects placeholders so the rendered config parses cleanly under `nginx -t`:

- `DOKKU_APP_WEB_LISTENERS` is set to `127.0.0.1:5000` so the legacy `upstream` block has a static server entry.
- `DOKKU_APP_WEB_LISTENER_HOST` is set to `127.0.0.1` so the resolver-based `set $dokku_upstream "...";` line renders to a syntactically valid value.

Pre-validation is skipped when the proxy type is not `nginx` or when `disable-custom-config` is set to `true` for the app.

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

The new variables (`DOKKU_APP_<PROC>_LISTENER_HOST` + `NGINX_DNS_RESOLVER` + `NGINX_DNS_RESOLVER_TIMEOUT` + `NGINX_DNS_ZONE`) describe a single hostname per process type of the form `<app>.<proctype>.<dns-zone>` (for example, `myapp.web.docker`). The hostname is served by [coredns-docker](https://github.com/dokku/coredns-docker) (>= `0.6.0`), which watches Docker events and collapses every running instance of the process type onto one multi-A record set. Combined with an nginx `resolver` directive and a variable-based `proxy_pass`, the upstream re-resolves at request time honoring DNS TTL - which gives runtime IP refresh on container restarts and DNS round-robin across scaled instances.

The bundled template emits this snippet:

```nginx
{{ if ne $.NGINX_DNS_RESOLVER "off" }}
resolver {{ $.NGINX_DNS_RESOLVER }} valid=10s ipv6=off;
resolver_timeout {{ $.NGINX_DNS_RESOLVER_TIMEOUT }};
set $dokku_upstream "{{ $.DOKKU_APP_WEB_LISTENER_HOST }}:{{ $upstream_port }}";
proxy_pass http://$dokku_upstream;
{{ else }}
proxy_pass http://{{ $.APP }}-{{ $upstream_port }};
{{ end }}
```

The `$dokku_upstream` variable name is intentionally namespaced (rather than `$upstream`) so vendored templates that already define `$upstream` will not collide if they adopt the resolver branch.

`NGINX_DNS_RESOLVER` defaults to `127.0.0.1:1053`, `NGINX_DNS_RESOLVER_TIMEOUT` defaults to `5s`, and `NGINX_DNS_ZONE` defaults to `docker`; all three are configurable via `dokku nginx:set --global dns-resolver ...`, `dokku nginx:set --global dns-resolver-timeout ...`, and `dokku nginx:set --global dns-zone ...`. The literal string `off` is the sentinel for disabling the resolver branch in the bundled template - empty strings are treated as "use the default" because the property is stored on disk and an empty value deletes the property file. The short default `dns-resolver-timeout` keeps clients from waiting on nginx's built-in 30s timeout when coredns-docker is unreachable; nginx returns 502 after the configured interval instead.

### Choosing between the two

The DNS-based path is the default for the bundled template. Vendored templates that have not been updated continue to use `DOKKU_APP_<PROC>_LISTENERS` and are unaffected by the global default. The DNS-based variables are populated regardless of whether `dns-resolver` is `off` or not, so a vendored template can adopt the resolver path on its own schedule by switching its rendering snippet, even if other apps on the host are still using IP-based upstreams.
