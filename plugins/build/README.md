# Dokku core build command

This plugin defines the `build` command used internally to build an app
(usually after receiving an app with the internally-used `receive` command).

It runs any pre-build hooks, builds a container for the app using
[buildstep][], then runs any post-build hooks. Further release and deployment,
if any, is handled by the external context (again, usually the `receive`
command).

[buildstep]: https://github.com/progrium/buildstep
