# Herokuish Buildpacks

> [!IMPORTANT]
> Subcommands new as of 0.15.0

```
buildpacks:set-property [--global|<app>] <key> <value>  # Set or clear a buildpacks property for an app
```

```
builder-herokuish:report [<app>] [<flag>]   # Displays a builder-herokuish report for one or more apps
builder-herokuish:set <app> <key> (<value>) # Set or clear a builder-herokuish property for an app
```

> [!WARNING]
> If using the `buildpacks` plugin, be sure to unset any `BUILDPACK_URL` and remove any such entries from a committed `.env` file. A specified `BUILDPACK_URL` will always override a `.buildpacks` file or the buildpacks plugin.

Dokku normally defaults to using [Heroku buildpacks](https://devcenter.heroku.com/articles/buildpacks) for deployment, though this may be overridden by committing a valid `Dockerfile` to the root of your repository and pushing the repository to your Dokku installation. To avoid this automatic `Dockerfile` deployment detection, you may do one of the following:

- Set a `BUILDPACK_URL` environment variable
    - This can be done via `dokku config:set` or via a committed `.env` file in the root of the repository. See the [environment variable documentation](/docs/configuration/environment-variables.md) for more details.
- Create a `.buildpacks` file in the root of your repository.
    - This can be via a committed `.buildpacks` file or managed via the `buildpacks` plugin commands.

This page will cover usage of the `buildpacks` plugin.

## Usage

### Detection

This builder will be auto-detected in either the following cases:

- The `BUILDPACK_URL` app environment variable is set.
    - This can be done via `dokku config:set` or via a committed `.env` file in the root of the repository. See the [environment variable documentation](/docs/configuration/environment-variables.md) for more details.
- A `.buildpacks` file exists in the root of the app repository.
    - This can be via a committed `.buildpacks` file or managed via the `buildpacks` plugin commands.

The builder can also be specified via the `builder:set` command:

```shell
dokku builder:set node-js-app selected herokuish
```

> Dokku will only select the `dockerfile` builder if both the `herokuish` and `pack` builders are not detected and a Dockerfile exists. See the [dockerfile builder documentation](/docs/deployment/builders/dockerfiles.md) for more information on how that builder functions.

### Customizing the Buildpack stack builder

> [!IMPORTANT]
> New as of 0.23.0

The default stack builder in use by Herokuish buildpacks in Dokku is based on `gliderlabs/herokuish:latest`. Typically, this is installed via an OS package which pulls the requisite Docker image. Users may desire to switch the stack builder to a custom version, either to update the operating system or to customize packages included with the stack builder. This can be performed via the `buildpacks:set-property` command.

```shell
dokku buildpacks:set-property node-js-app stack gliderlabs/herokuish:latest
```

The specified stack builder can also be unset by omitting the name of the stack builder when calling `buildpacks:set-property`.

```shell
dokku buildpacks:set-property node-js-app stack
```

A change in the stack builder value will execute the `post-stack-set` trigger.

Finally, stack builders can be set or unset globally as a fallback. This will take precedence over a globally set `DOKKU_IMAGE` environment variable (`gliderlabs/herokuish:latest-24` by default).

```shell
# set globally
dokku buildpacks:set-property --global stack gliderlabs/herokuish:latest

# unset globally
dokku buildpacks:set-property --global stack
```

### Allowing herokuish for non-amd64 platforms

> [!IMPORTANT]
> New as of 0.29.0

By default, the builder-herokuish plugin is not enabled for non-amd64 platforms, and attempting to use it is blocked. This is because the majority of buildpacks are not cross-platform compatible, and thus building apps will either be considerably slower - due to emulating the amd64 platform - or won't work - due to building amd64 packages on arm64 platforms.

To force-enable herokuish on non-amd64 platforms, the `allowed` property can be set via `builder-herokuish:set`. The default value depends on the host platform architecture (`true` on amd64, `false` otherwise).

```shell
dokku builder-herokuish:set node-js-app allowed true
```

The default value may be set by passing an empty value for the option:

```shell
dokku builder-herokuish:set node-js-app allowed
```

The `allowed` property can also be set globally. The global default is platform-dependent, and the global value is used when no app-specific value is set.

```shell
dokku builder-herokuish:set --global allowed true
```

The default value may be set by passing an empty value for the option.

```shell
dokku builder-herokuish:set --global allowed
```

### Displaying builder-herokuish reports for an app

> [!IMPORTANT]
> New as of 0.29.0

You can get a report about the app's storage status using the `builder-herokuish:report` command:

```shell
dokku builder-herokuish:report
```

```
=====> node-js-app builder-herokuish information
       Builder herokuish computed allowed: false
       Builder herokuish global allowed:   true
       Builder herokuish allowed:          false
=====> python-sample builder-herokuish information
       Builder herokuish computed allowed: true
       Builder herokuish global allowed:   true
       Builder herokuish allowed:
=====> ruby-sample builder-herokuish information
       Builder herokuish computed allowed: true
       Builder herokuish global allowed:   true
       Builder herokuish allowed:
```

You can run the command for a specific app also.

```shell
dokku builder-herokuish:report node-js-app
```

```
=====> node-js-app builder-herokuish information
       Builder herokuish computed allowed: false
       Builder herokuish global allowed:   true
       Builder herokuish allowed:          false
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku builder-herokuish:report node-js-app --builder-herokuish-allowed
```

```
false
```

## Errata

### Switching from Dockerfile deployments

If an application was previously deployed via Dockerfile, the following commands should be run before a buildpack deploy will succeed:

```shell
dokku ports:clear node-js-app
```

### `curl` build timeouts

Certain buildpacks may time out in retrieving dependencies via `curl`. This can happen when your network connection is poor or if there is significant network congestion. You may see a message similar to `gzip: stdin: unexpected end of file` after a `curl` command.

If you see output similar this when deploying , you may need to override the `curl` timeouts to increase the length of time allotted to those tasks. You can do so via the `config` plugin:

```shell
dokku config:set --global CURL_TIMEOUT=1200
dokku config:set --global CURL_CONNECT_TIMEOUT=180
```

### Clearing buildpack cache

See the [repository management documentation](/docs/advanced-usage/repository-management.md#clearing-app-cache) for more information on how to clear buildpack build cache for an application.

### Specifying commands via Procfile

See the [Procfile documentation](/docs/processes/process-management.md#procfile) for more information on how to specify different processes for your app.

### Listing Buildpacks in Use

See the [buildpack management documentation](/docs/processes/process-management.md#listing-buildpacks-in-use) for more information on how to list buildpacks in use.

### Adding custom buildpacks

See the [buildpack management documentation](/docs/processes/process-management.md#adding-custom-buildpacks) for more information on how to add custom buildpacks.

### Overwriting a buildpack position

See the [buildpack management documentation](/docs/processes/process-management.md#overwriting-a-buildpack-position) for more information on how to overwrite a buildpack position.

### Removing a buildpack

See the [buildpack management documentation](/docs/processes/process-management.md#removing-a-buildpack) for more information on how to remove a buildpack.

### Clearing all buildpacks

See the [buildpack management documentation](/docs/processes/process-management.md#clearing-all-buildpacks) for more information on how to clear all buildpacks.

### Using a specific buildpack version

See the [buildpack management documentation](/docs/processes/process-management.md#using-a-specific-buildpack-version) for more information on how to using a specific buildpack version


### Displaying buildpack reports for an app

See the [buildpack management documentation](/docs/processes/process-management.md#displaying-buildpack-reports-for-an-app) for more information on how to display buildpack reports for an app.