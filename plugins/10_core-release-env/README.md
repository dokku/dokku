# Dokku core app release command

This plugin defines the `release` command used internally to cut a new release
of an app with a new configuration environment (usually after the `receive`
command has built an app, or when setting a new configuration via the `config`
command).

It runs any pre-release hooks, creates a new container with the new environment
included as its profile, then runs any post-release hooks. Further deployment,
if any, is handled by the external context.
