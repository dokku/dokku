# nginx.conf.sigil

The `nginx.conf.sigil` file is used to configure the nginx server for an application. The default template can be found [here](https://github.com/dokku/dokku/blob/master/plugins/nginx-vhosts/templates/nginx.conf.sigil). Dokku uses a tool named [sigil](https://github.com/gliderlabs/sigil) to generate the nginx configuration based on the template provided.

## Validation

A custom `nginx.conf.sigil` is pre-validated at the start of every deploy, immediately after it is extracted from the source tree and before the build phase runs. Pre-validation renders the template via sigil with the same parameters used at deploy time, wraps the rendered config in a minimal `events`/`http` scaffold, and runs `nginx -t` against the result. The deploy is aborted if either the sigil render or the `nginx -t` check fails, so build work is not wasted on a syntactically invalid template. When no app listeners exist yet (typical for first deploys), pre-validation injects a placeholder `127.0.0.1:5000` listener for `DOKKU_APP_WEB_LISTENERS` so that the rendered upstream block has a static server entry and `nginx -t` does not bail out on "host not found in upstream".

Pre-validation is skipped when the proxy type is not `nginx` or when `disable-custom-config` is set to `true` for the app.
