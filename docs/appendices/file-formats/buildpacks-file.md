# .buildpacks

The `.buildpacks` file is used to specify the buildpacks for an application. It is a list of buildpack URLs.

```shell
https://github.com/heroku/heroku-buildpack-nodejs.git
https://github.com/heroku/heroku-buildpack-ruby.git
```

When installed, a buildpack is checked to see if it is a git repository via `git ls-remote`. If it is not, the buildpack is checked to see if it ends in `.tgz`, `.tar.gz`, `.tbz`, `tar.bz`, or `.tar` and is downloaded/extracted before being executed.

If the buildpack is located on Github, a shorthand can be used in the form of `organization/buildpack-name`. This will be expanded to the `https://github.com/organization/heroku-buildpack-buildpack-name.git`. The `heroku-community` organization name is treated as `heroku`.

Comments are not allowed in the `.buildpacks` file.
