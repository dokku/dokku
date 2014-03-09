# Dokku core

This plugin provides core functionality for Dokku, in the form of the base
Dokku installation steps, as well as the implementation of most core commands
and hooks that add some of their base functionality.

# Dokku core installation

- Sets the dokku HOSTNAME file from the server's hostname, if missing.
- Creates an Upstart rule to restart all apps after a reboot (redeploying them
  to refresh their port mappings for virtual hosting).

# Dokku core commands

These commands can be accessed directly as `core:<command>`. The `commands`
script in this plugin makes it so that each of these subcommands is responsive
unprefixed.

## build

This command is used internally to build an app (usually after receiving an
app with the `receive` command).

It runs any pre-build hooks, builds a container for the app using
[buildstep][], then runs any post-build hooks. Further release and deployment,
if any, is handled by the external context.

[buildstep]: https://github.com/progrium/buildstep

## cleanup

This command is used internally to clean up any unused images and/or containers
when receiving an app with the `receive` command.

## delete

This command runs the stops the container for a running app and deletes its
Docker image. Additionally, the pre-delete hook deletes the app's build cache,
and the post-delete hook deletes the app's repository directory.

## receive

This command is used internally to build, release, and deploy an app when
called, for instance, by a receive hook as managed by the core `git` Dokku
plugin.

## release

This command is used internally to cut a new release of an app with a new
configuration environment (usually after the `receive` command has built an
app, or when setting a new configuration via the `config` command).

It runs any pre-release hooks, creates a new container with the new environment
included as its profile, then runs any post-release hooks. Further deployment,
if any, is handled by the external context.

## run

This command runs a command in the environment of an app's container.

## logs

This command shows logs from a running app's container.

## url

This command shows the URL at which a running app may be accessed.
