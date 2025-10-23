# Railpack

> [!IMPORTANT]
> New as of 0.32.0

The `railpack` builder builds apps via [Railpack](https://railpack.com/), a buildpack alternative.

## Usage

### Requirements

Before using Railpacks, the following steps must be taken:

- Install the `railpack` cli: The `railpack` cli tool is not included by default with Dokku or as a dependency. It must also be installed as shown on [this page](https://railpack.com/installation).
- Create a `buildkit` builder: Railpack uses buildkit.
       ```shell
       docker run --rm --privileged -d --name buildkit moby/buildkit
       ```
- Set the buildkit builder: Update the `/etc/default/dokku` file to set `BUILDKIT_HOST`:
       ```shell
       touch /etc/default/dokku
       echo "export BUILDKIT_HOST='docker-container://buildkit'" >> /etc/default/dokku
       ````

### Detection

This builder will be auto-detected in the following case:

- A `railpack.json` exists in the root of the app repository.

The builder may also be selected via the `builder:set` command

```shell
dokku builder:set node-js-app selected railpack
```

### Supported languages

See the [upstream railpack documentation](https://railpack.com/) for further information on what languages and frameworks are supported.

### Build-time configuration variables

For security reasons - and as per [Docker recommendations](https://github.com/docker/docker/issues/13490) - railpack-based deploys have variables available only during runtime.

For users that require customization in the `build` phase, you may use build arguments via the [docker-options plugin](/docs/advanced-usage/docker-options.md). All environment variables set by the `config` plugin are automatically exported within the railpack build environment, and thus `--env` only requires setting a key without a value.

```shell
dokku docker-options:add node-js-app build '--env NODE_ENV'
```

Alternatively, a full value may be provided in the form of `--env KEY=VALUE`:

```shell
dokku docker-options:add node-js-app build '--env NODE_ENV=production'
```

### Changing the `railpack.json` location

The `railpack.json` is expected to be found in a specific directory, depending on the deploy approach:

- The `WORKDIR` of the Docker image for deploys resulting from `git:from-image` and `git:load-image` commands.
- The root of the source code tree for all other deploys (git push, `git:from-archive`, `git:sync`).

Sometimes it may be desirable to set a different path for a given app, e.g. when deploying from a monorepo. This can be done via the `railpackjson-path` property:

```shell
dokku builder-railpack:set node-js-app railpackjson-path .dokku/railpack.json
```

The value is the path to the desired file *relative* to the base search directory, and will never be treated as absolute paths in any context. If that file does not exist within the repository, the build will fail.

The default value may be set by passing an empty value for the option:

```shell
dokku builder-railpack:set node-js-app railpackjson-path
```

The `railpackjson-path` property can also be set globally. The global default is `railpack.json`, and the global value is used when no app-specific value is set.

```shell
dokku builder-railpack:set --global railpackjson-path railpack2.json
```

The default value may be set by passing an empty value for the option.

```shell
dokku builder-railpack:set --global railpackjson-path
```

### Disabling cache

Cache is enabled by default, but can be disabled by setting the `no-cache` property to `true`:

```shell
dokku builder-railpack:set node-js-app no-cache true
```

The default value may be set by passing an empty value for the option:

```shell
dokku builder-railpack:set node-js-app no-cache
```

The `no-cache` property can also be set globally. The global default is `false`, and the global value is used when no app-specific value is set.

```shell
dokku builder-railpack:set --global no-cache true
```

The default value may be set by passing an empty value for the option.

```shell
dokku builder-railpack:set --global no-cache
```

### Displaying builder-railpack reports for an app

You can get a report about the app's storage status using the `builder-railpack:report` command:

```shell
dokku builder-railpack:report
```

```
=====> node-js-app builder-railpack information
       Builder-railpack computed railpackjson path: railpack2.json
       Builder-railpack global railpackjson path:   railpack.json
       Builder-railpack railpackjson path:          railpack2.json
       Builder-railpack computed no cache:          true
       Builder-railpack global no cache:            false
       Builder-railpack no cache:                   true
=====> python-sample builder-railpack information
       Builder-railpack computed railpackjson path: railpack.json
       Builder-railpack global railpackjson path:   railpack.json
       Builder-railpack railpackjson path:
       Builder-railpack computed no cache:          false
       Builder-railpack global no cache:            false
       Builder-railpack no cache:
=====> ruby-sample builder-railpack information
       Builder-railpack computed railpackjson path: railpack.json
       Builder-railpack global railpackjson path:   railpack.json
       Builder-railpack railpackjson path:
       Builder-railpack computed no cache:          false
       Builder-railpack global no cache:            false
       Builder-railpack no cache:
```

You can run the command for a specific app also.

```shell
dokku builder-railpack:report node-js-app
```

```
=====> node-js-app builder-railpack information
       Builder-railpack computed railpackjson path: railpack2.json
       Builder-railpack global railpackjson path:   railpack.json
       Builder-railpack railpackjson path:          railpack2.json
       Builder-railpack computed no cache:          true
       Builder-railpack global no cache:            false
       Builder-railpack no cache:                   true
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku builder-railpack:report node-js-app --builder-railpack-no-cache
```

```
true
```
