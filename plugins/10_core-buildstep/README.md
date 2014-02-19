# Dokku core app build command

This plugin defines the `build` command used internally to build an app
(usually after receiving an app with the `receive` command as defined in the
`core-receive-cycle` plugin).

It runs any pre-build hooks, builds a container for the app using
[buildstep][], then runs any post-build hooks. Further release and deployment,
if any, is handled by the external context.

[buildstep]: https://github.com/progrium/buildstep
