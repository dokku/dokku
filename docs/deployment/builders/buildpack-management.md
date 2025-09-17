# Buildpacks Management

> [!NOTE]
> Buildpacks commands apply to both herokuish and CNB-based builds

```
buildpacks:add [--index 1] <app> <buildpack>            # Add new app buildpack while inserting into list of buildpacks if necessary
buildpacks:clear <app>                                  # Clear all buildpacks set on the app
buildpacks:list <app>                                   # List all buildpacks for an app
buildpacks:remove <app> <buildpack>                     # Remove a buildpack set on the app
buildpacks:report [<app>] [<flag>]                      # Displays a buildpack report for one or more apps
buildpacks:set [--index 1] <app> <buildpack>            # Set new app buildpack at a given position defaulting to the first buildpack if no index is specified
```

## Usage

### Listing Buildpacks in Use

The `buildpacks:list` command can be used to show buildpacks that have been set for an app. This will omit any auto-detected buildpacks.

```shell
# running for an app with no buildpacks specified
dokku buildpacks:list node-js-app
```

```
-----> test buildpack urls
```

```shell
# running for an app with two buildpacks specified
dokku buildpacks:list node-js-app
```

```
-----> test buildpack urls
       https://github.com/heroku/heroku-buildpack-python.git
       https://github.com/heroku/heroku-buildpack-nodejs.git
```

### Adding custom buildpacks

> Please check the documentation for your particular buildpack as you may need to include configuration files (such as a Procfile) in your project root.

To add a custom buildpack, use the `buildpacks:add` command:

```shell
dokku buildpacks:add node-js-app https://github.com/heroku/heroku-buildpack-nodejs.git
```

When no buildpacks are currently specified, the specified buildpack will be the only one executed for detection and compilation.

Multiple buildpacks may be specified by using the `buildpacks:add` command multiple times.

```shell
dokku buildpacks:add node-js-app https://github.com/heroku/heroku-buildpack-ruby.git
dokku buildpacks:add node-js-app https://github.com/heroku/heroku-buildpack-nodejs.git
```

Buildpacks are executed in order, may be inserted at a specified index via the `--index` flag. This flag is specified starting at a 1-index value.

```shell
# will add the golang buildpack at the second position, bumping all proceeding ones by 1 position
dokku buildpacks:add --index 2 node-js-app https://github.com/heroku/heroku-buildpack-golang.git
```

### Overwriting a buildpack position

In some cases, it may be necessary to swap out a given buildpack. Rather than needing to re-specify each buildpack, the `buildpacks:set` command can be used to overwrite a buildpack at a given position.

```shell
dokku buildpacks:set node-js-app https://github.com/heroku/heroku-buildpack-ruby.git
```

By default, this will overwrite the _first_ buildpack specified. To specify an index, the `--index` flag may be used. This flag is specified starting at a 1-index value, and defaults to `1`.

```shell
# the following are equivalent commands
dokku buildpacks:set node-js-app https://github.com/heroku/heroku-buildpack-ruby.git
dokku buildpacks:set --index 1 node-js-app https://github.com/heroku/heroku-buildpack-ruby.git
```

If the index specified is larger than the number of buildpacks currently configured, the buildpack will be appended to the end of the list.

```shell
dokku buildpacks:set --index 99 node-js-app https://github.com/heroku/heroku-buildpack-ruby.git
```

### Removing a buildpack

> At least one of a buildpack or index must be specified

A single buildpack can be removed by name via the `buildpacks:remove` command.

```shell
dokku buildpacks:remove node-js-app https://github.com/heroku/heroku-buildpack-ruby.git
```

Buildpacks can also be removed by index via the `--index` flag. This flag is specified starting at a 1-index value.

```shell
dokku buildpacks:remove node-js-app --index 1
```

### Clearing all buildpacks

> This does not affect automatically detected buildpacks, nor does it impact any specified `BUILDPACK_URL` environment variable.

The `buildpacks:clear` command can be used to clear all configured buildpacks for a specified app.

```shell
dokku buildpacks:clear node-js-app
```

### Using a specific buildpack version

> Always remember to pin your buildpack versions when using the multi-buildpacks method, or you may find deploys changing your deployed environment.

By default, builders will pin their vendored buildpacks, resulting in a consistent build result for your application. There may be occasions where the pinned version results in a broken deploy, or does not have a particular feature that is required to build your project. To use a more recent version of a given buildpack, the buildpack may be specified _without_ a Git commit SHA like so:

```shell
# using the latest nodejs buildpack
dokku buildpacks:set node-js-app https://github.com/heroku/heroku-buildpack-nodejs
```

This will use the latest commit on the `master` branch of the specified buildpack. To pin to a newer version of a buildpack, a sha may also be specified by using the form `REPOSITORY_URL#COMMIT_SHA`, where `COMMIT_SHA` is any tree-ish git object - usually a git tag.

```shell
# using v87 of the nodejs buildpack
dokku buildpacks:set node-js-app https://github.com/heroku/heroku-buildpack-nodejs#v87
```

### Displaying buildpack reports for an app

You can get a report about the app's buildpacks status using the `buildpacks:report` command:

```shell
dokku buildpacks:report
```

```
=====> node-js-app buildpacks information
       Buildpacks computed stack:  gliderlabs/herokuish:v0.7.0-22
       Buildpacks global stack:    gliderlabs/herokuish:latest-24
       Buildpacks list:            https://github.com/heroku/heroku-buildpack-nodejs.git
       Buildpacks stack:           gliderlabs/herokuish:v0.7.0-20
=====> python-sample buildpacks information
       Buildpacks computed stack:  gliderlabs/herokuish:latest-24
       Buildpacks global stack:    gliderlabs/herokuish:latest-24
       Buildpacks list:            https://github.com/heroku/heroku-buildpack-nodejs.git,https://github.com/heroku/heroku-buildpack-python.git
       Buildpacks stack:
=====> ruby-sample buildpacks information
       Buildpacks computed stack:  gliderlabs/herokuish:latest-24
       Buildpacks global stack:    gliderlabs/herokuish:latest-24
       Buildpacks list:
       Buildpacks stack:
```

You can run the command for a specific app also.

```shell
dokku buildpacks:report node-js-app
```

```
=====> node-js-app buildpacks information
       Buildpacks list:               https://github.com/heroku/heroku-buildpack-nodejs.git
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku buildpacks:report node-js-app --buildpacks-list
```
