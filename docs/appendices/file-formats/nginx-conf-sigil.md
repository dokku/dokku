# nginx.conf.sigil

The `nginx.conf.sigil` file is used to configure the nginx server for an application. The default template can be found [here](https://github.com/dokku/dokku/blob/master/plugins/nginx-vhosts/templates/nginx.conf.sigil). Dokku uses a tool named [sigil](https://github.com/gliderlabs/sigil) to generate the nginx configuration based on the template provided.

## Validation

A custom `nginx.conf.sigil` is pre-validated at the start of every deploy, immediately after it is extracted from the source tree and before the build phase runs. Pre-validation renders the template via sigil with the same parameters used at deploy time (using empty `DOKKU_APP_WEB_LISTENERS` for first deploys), wraps the rendered config in a minimal `events`/`http` scaffold, and runs `nginx -t` against the result. The deploy is aborted if either the sigil render or the `nginx -t` check fails, so build work is not wasted on a syntactically invalid template.

Pre-validation is skipped when the proxy type is not `nginx` or when `disable-custom-config` is set to `true` for the app.
