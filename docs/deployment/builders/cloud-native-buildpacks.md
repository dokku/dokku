# Cloud Native Buildpacks

> [!IMPORTANT]
> New as of 0.22.0

```
builder-pack:report [<app>] [<flag>]   # Displays a builder-pack report for one or more apps
builder-pack:set <app> <key> (<value>) # Set or clear a builder-pack property for an app
```

```
buildpacks:set-property [--global|<app>] <key> <value>  # Set or clear a buildpacks property for an app
```

Cloud Native Buildpacks are an evolution over the Buildpacks technology provided by the Herokuish builder. See the [herokuish buildpacks documentation](/docs/deployment/builders/herokuish-buildpacks.md) for more information on how to clear buildpack build cache for an application.

> [!WARNING]
> This functionality uses the `pack` cli from the [Cloud Native Buildpacks](https://buildpacks.io) project to build apps. As the integration is experimental in Dokku, it is likely to change over time.

## Usage

### Requirements

The `pack` cli tool is not included by default with Dokku or as a dependency. It must also be installed as shown on [this page](https://buildpacks.io/docs/tools/pack/).

Builds will proceed with the `pack` cli for the app from then on.

### Caveats

As this functionality is highly experimental, there are a number of caveats. Please note that not all issues are listed below.

- Specifying specific buildpacks is not currently possible.
    - A future release will add support for specifying buildpacks via the `buildpacks` plugin.
- There is currently no way to specify extra arguments for `pack` cli invocations.
    - A future release will add support for injecting extra arguments during the build process.
- The default process type is `web`.
- Build cache is stored in Docker volumes instead of on disk. As such, `repo:purge-cache` currently has no effect.
    - A future version will add integration with the `repo` plugin.
- `pack` is not currently included with Dokku, nor is it added as a package dependency.
    - A future version will include it as a package dependency.

### Detection

This builder will be auto-detected in either the following cases:

- The `DOKKU_CNB_EXPERIMENTAL` app environment variable is set to `1`.

    ```shell
    dokku config:set --no-restart node-js-app DOKKU_CNB_EXPERIMENTAL=1
    ```

- A `project.toml` file exists in the root of the app repository.
    - This file is consumed by `pack-cli` and used to describe how the app is built.

The builder can also be specified via the `builder:set` command:

```shell
dokku builder:set node-js-app selected pack
```

> Dokku will only select the `dockerfile` builder if both the `herokuish` and `pack` builders are not detected and a Dockerfile exists. See the [dockerfile builder documentation](/docs/deployment/builders/dockerfiles.md) for more information on how that builder functions.

### Changing the `project.toml` location

The `project.toml` is expected to be found in a specific directory, depending on the deploy approach:

- The `WORKDIR` of the Docker image for deploys resulting from `git:from-image` and `git:load-image` commands.
- The root of the source code tree for all other deploys (git push, `git:from-archive`, `git:sync`).

Sometimes it may be desirable to set a different path for a given app, e.g. when deploying from a monorepo. This can be done via the `projecttoml-path` property:

```shell
dokku builder-pack:set node-js-app projecttoml-path .dokku/project.toml
```

The value is the path to the desired file *relative* to the base search directory, and will never be treated as absolute paths in any context. If that file does not exist within the repository, Dokku will continue the build process as if the repository has no `project.toml`.

The default value may be set by passing an empty value for the option:

```shell
dokku builder-pack:set node-js-app projecttoml-path
```

The `projecttoml-path` property can also be set globally. The global default is `project.toml`, and the global value is used when no app-specific value is set.

```shell
dokku builder-pack:set --global projecttoml-path project2.toml
```

The default value may be set by passing an empty value for the option.

```shell
dokku builder-pack:set --global projecttoml-path
```

### Displaying builder-pack reports for an app

> [!IMPORTANT]
> New as of 0.25.0

You can get a report about the app's storage status using the `builder-pack:report` command:

```shell
dokku builder-pack:report
```

```
=====> node-js-app builder-pack information
       Builder pack computed projecttoml path: project2.toml
       Builder pack global projecttoml path:   project.toml
       Builder pack projecttoml path:          project2.toml
=====> python-sample builder-pack information
       Builder pack computed projecttoml path: project.toml
       Builder pack global projecttoml path:   project.toml
       Builder pack projecttoml path:
=====> ruby-sample builder-pack information
       Builder pack computed projecttoml path: project.toml
       Builder pack global projecttoml path:   project.json
       Builder pack projecttoml path:
```

You can run the command for a specific app also.

```shell
dokku builder-pack:report node-js-app
```

```
=====> node-js-app builder-pack information
       Builder pack computed projecttoml path: project2.toml
       Builder pack global projecttoml path:   project.toml
       Builder pack projecttoml path:          project2.toml
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku builder-pack:report node-js-app --builder-pack-projecttoml-path
```

```
project2.toml
```

### Customizing the Buildpack stack builder

> [!IMPORTANT]
> New as of 0.23.0

The default stack builder in use by CNB buildpacks in Dokku is based on `heroku/builder:22`. Users may desire to switch the stack builder to a custom version, either to update the operating system or to customize packages included with the stack builder. This can be performed via the `buildpacks:set-property` command.

```shell
dokku buildpacks:set-property node-js-app stack paketobuildpacks/build:base-cnb
```

The specified stack builder can also be unset by omitting the name of the stack builder when calling `buildpacks:set-property`.

```shell
dokku buildpacks:set-property node-js-app stack
```

A change in the stack builder value will execute the `post-stack-set` trigger.

Finally, stack builders can be set or unset globally as a fallback. This will take precedence over a globally set `DOKKU_CNB_BUILDER` environment variable (`heroku/builder:22` by default).

```shell
# set globally
dokku buildpacks:set-property --global stack paketobuildpacks/build:base-cnb

# unset globally
dokku buildpacks:set-property --global stack
```

### Specifying commands via Procfile

See the [Procfile documentation](/docs/processes/process-management.md#procfile) for more information on how to specify different processes for your app.
