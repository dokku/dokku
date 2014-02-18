# Dokku core deploy command

This plugin defines the `deploy` command used internally to start a released
app container (usually after the `receive` command has built and released an
app, or when setting a new configuration via the `config` command).

It runs any pre-deploy hooks, starts the app (saving its port and external
URL), then runs any post-deploy hooks.
