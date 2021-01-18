# Herokuish Buildpack Deployment

> Subcommands new as of 0.15.0

```
buildpacks:add [--index 1] <app> <buildpack>  # Add new app buildpack while inserting into list of buildpacks if necessary
buildpacks:clear <app>                        # Clear all buildpacks set on the app
buildpacks:list <app>                         # List all buildpacks for an app
buildpacks:remove <app> <buildpack>           # Remove a buildpack set on the app
buildpacks:report [<app>] [<flag>]            # Displays a buildpack report for one or more apps
buildpacks:set [--index 1] <app> <buildpack>  # Set new app buildpack at a given position defaulting to the first buildpack if no index is specified
buildpacks:stacks-set <app> <stack>           # Sets the stack of an app
```

> Warning: If using the `buildpacks` plugin, be sure to unset any `BUILDPACK_URL` and remove any such entries from a committed `.env` file. A specified `BUILDPACK_URL` will always override a `.buildpacks` file or the buildpacks plugin.

Dokku normally defaults to using [Heroku buildpacks](https://devcenter.heroku.com/articles/buildpacks) for deployment, though this may be overridden by committing a valid `Dockerfile` to the root of your repository and pushing the repository to your Dokku installation. To avoid this automatic `Dockerfile` deployment detection, you may do one of the following:

- Set a `BUILDPACK_URL` environment variable
  - This can be done via `dokku config:set` or via a committed `.env` file in the root of the repository. See the [environment variable documentation](/docs/configuration/environment-variables.md) for more details.
- Create a `.buildpacks` file in the root of your repository.
  - This can be via a committed `.buildpacks` file or managed via the `buildpacks` plugin commands.

This page will cover usage of the `buildpacks` plugin.

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

### Customizing the Buildpack stack

> New as of 0.23.0

The default stack in use by Herokuish buildpacks in Dokku is based on `gliderlabs/herokuish`. Typically, this is installed via an OS package which pulls the requisite Docker image. Users may desire to switch the stack to a custom version, either to update the stack operating system or to customize packages included with the stack. This can be performed via teh `buildpacks:stack-set` command.

```shell
dokku buildpacks:stack-set node-js-app gliderlabs/herokuish:latest
```

The specified stack can also be unset by omitting the name of the stack when calling `buildpacks:stack-set`.

```shell
dokku buildpacks:stack-set node-js-app
```

Finally, stacks can be set or unset globally as a fallback. This will take precedence over a globally set `DOKKU_IMAGE` environment variable (`gliderlabs/herokuish:latest` by default).

```shell
# set globally
dokku buildpacks:stack-set --global gliderlabs/herokuish:latest

# unset globally
dokku buildpacks:stack-set --global
```

### Displaying buildpack reports for an app

You can get a report about the app's buildpacks status using the `buildpacks:report` command:

```shell
dokku buildpacks:report
```

```
=====> node-js-app buildpacks information
       Buildpacks computed stack:  gliderlabs/herokuish:v0.5.23-20
       Buildpacks global stack:    gliderlabs/herokuish:latest
       Buildpacks list:            https://github.com/heroku/heroku-buildpack-nodejs.git
       Buildpacks stack:           gliderlabs/herokuish:v0.5.23-20
=====> python-sample buildpacks information
       Buildpacks computed stack:  gliderlabs/herokuish:v0.5.23-20
       Buildpacks global stack:    gliderlabs/herokuish:latest
       Buildpacks list:            https://github.com/heroku/heroku-buildpack-nodejs.git,https://github.com/heroku/heroku-buildpack-python.git
       Buildpacks stack:
=====> ruby-sample buildpacks information
       Buildpacks computed stack:  gliderlabs/herokuish:v0.5.23-20
       Buildpacks global stack:    gliderlabs/herokuish:latest
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

## Errata

### Switching from Dockerfile deployments

If an application was previously deployed via Dockerfile, the following commands should be run before a buildpack deploy will succeed:

```shell
dokku config:unset --no-restart node-js-app DOKKU_PROXY_PORT_MAP
```

### Using a specific buildpack version

> Always remember to pin your buildpack versions when using the multi-buildpacks method, or you may find deploys changing your deployed environment.

By default, Dokku uses the [gliderlabs/herokuish](https://github.com/gliderlabs/herokuish/) project, which pins all of it's vendored buildpacks. There may be occasions where the pinned version results in a broken deploy, or does not have a particular feature that is required to build your project. To use a more recent version of a given buildpack, the buildpack may be specified *without* a Git commit SHA like so:

```shell
# using the latest nodejs buildpack
dokku buildpacks:set node-js-app https://github.com/heroku/heroku-buildpack-nodejs
```

This will use the latest commit on the `master` branch of the specified buildpack. To pin to a newer version of a buildpack, a sha may also be specified by using the form `REPOSITORY_URL#COMMIT_SHA`, where `COMMIT_SHA` is any tree-ish git object - usually a git tag.

```shell
# using v87 of the nodejs buildpack
dokku buildpacks:set node-js-app https://github.com/heroku/heroku-buildpack-nodejs#v87
```

### Specifying commands via Procfile

While many buildpacks have a default command that is run when a detected repository is pushed, it is possible to override this command via a Procfile. A Procfile can also be used to specify multiple commands, each of which is subject to process scaling. See the [process scaling documentation](/docs/deployment/process-management.md) for more details around scaling individual processes.

A Procfile is a file named `Procfile`. It should be named `Procfile` exactly, and not anything else. For example, `Procfile.txt` is not valid. The file should be a simple text file.

The file must be placed in the root directory of your application. It will not function if placed in a subdirectory.

If the file exists, it should not be empty, as doing so may result in a failed deploy.

The syntax for declaring a `Procfile` is as follows. Note that the format is one process type per line, with no duplicate process types.

```
<process type>: <command>
```

If, for example, you have multiple queue workers and wish to scale them separately, the following would be a valid way to work around the requirement of not duplicating process types:

```Procfile
worker:           env QUEUE=* bundle exec rake resque:work
importantworker:  env QUEUE=important bundle exec rake resque:work
```

The `web` process type holds some significance in that it is the only process type that is automatically scaled to `1` on the initial application deploy. See the [process scaling documentation](/docs/deployment/process-management.md) for more details around scaling individual processes.


### `curl` build timeouts

Certain buildpacks may time out in retrieving dependencies via `curl`. This can happen when your network connection is poor or if there is significant network congestion. You may see a message similar to `gzip: stdin: unexpected end of file` after a `curl` command.

If you see output similar this when deploying , you may need to override the `curl` timeouts to increase the length of time allotted to those tasks. You can do so via the `config` plugin:

```shell
dokku config:set --global CURL_TIMEOUT=1200
dokku config:set --global CURL_CONNECT_TIMEOUT=180
```

### Clearing buildpack cache

See the [repository management documentation](/docs/advanced-usage/repository-management.md#clearing-app-cache) for more information on how to clear buildpack build cache for an application.
