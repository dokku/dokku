# Nixpacks

> [!IMPORTANT]
> New as of 0.32.0

The `nixpacks` builder builds apps via [Nixpacks](https://nixpacks.com/), a buildpack alternative.

## Usage

### Requirements

The `nixpacks` cli tool is not included by default with Dokku or as a dependency. It must also be installed as shown on [this page](https://nixpacks.com/docs/install#debian-(and-derivatives-like-ubuntu)).

Builds will proceed with the `nixpacks` cli for the app from then on.

### Detection

This builder will be auto-detected in the following case:

- A `nixpacks.toml` exists in the root of the app repository.

The builder may also be selected via the `builder:set` command

```shell
dokku builder:set node-js-app selected nixpacks
```

### Supported languages

See the [upstream nixpacks documentation](https://nixpacks.com/docs) for further information on what languages and frameworks are supported.

### Build-time configuration variables

For security reasons - and as per [Docker recommendations](https://github.com/docker/docker/issues/13490) - nixpacks-based deploys have variables available only during runtime.

For users that require customization in the `build` phase, you may use build arguments via the [docker-options plugin](/docs/advanced-usage/docker-options.md). All environment variables set by the `config` plugin are automatically exported within the nixpacks build environment, and thus `--env` only requires setting a key without a value.

```shell
dokku docker-options:add node-js-app build '--env NODE_ENV'
```

Alternatively, a full value may be provided in the form of `--env KEY=VALUE`:

```shell
dokku docker-options:add node-js-app build '--env NODE_ENV=production'
```

### Changing the `nixpacks.toml` location

The `nixpacks.toml` is expected to be found in a specific directory, depending on the deploy approach:

- The `WORKDIR` of the Docker image for deploys resulting from `git:from-image` and `git:load-image` commands.
- The root of the source code tree for all other deploys (git push, `git:from-archive`, `git:sync`).

Sometimes it may be desirable to set a different path for a given app, e.g. when deploying from a monorepo. This can be done via the `nixpackstoml-path` property:

```shell
dokku builder-nixpacks:set node-js-app nixpackstoml-path .dokku/nixpacks.toml
```

The value is the path to the desired file *relative* to the base search directory, and will never be treated as absolute paths in any context. If that file does not exist within the repository, the build will fail.

The default value may be set by passing an empty value for the option:

```shell
dokku builder-nixpacks:set node-js-app nixpackstoml-path
```

The `nixpackstoml-path` property can also be set globally. The global default is `nixpacks.toml`, and the global value is used when no app-specific value is set.

```shell
dokku builder-nixpacks:set --global nixpackstoml-path nixpacks2.toml
```

The default value may be set by passing an empty value for the option.

```shell
dokku builder-nixpacks:set --global nixpackstoml-path
```

### Disabling cache

Disable cache using the [docker-options] plugin:

```shell
dokku docker-options:add node-js-app build "--no-cache"
```

### Displaying builder-nixpacks reports for an app

You can get a report about the app's storage status using the `builder-nixpacks:report` command:

```shell
dokku builder-nixpacks:report
```

```
=====> node-js-app builder-nixpacks information
       Builder-nixpacks computed nixpackstoml path: nixpacks2.toml
       Builder-nixpacks global nixpackstoml path:   nixpacks.toml
       Builder-nixpacks nixpackstoml path:          nixpacks2.toml
=====> python-sample builder-nixpacks information
       Builder-nixpacks computed nixpackstoml path: nixpacks.toml
       Builder-nixpacks global nixpackstoml path:   nixpacks.toml
       Builder-nixpacks nixpackstoml path:
=====> ruby-sample builder-nixpacks information
       Builder-nixpacks computed nixpackstoml path: nixpacks.toml
       Builder-nixpacks global nixpackstoml path:   nixpacks.toml
       Builder-nixpacks nixpackstoml path:
```

You can run the command for a specific app also.

```shell
dokku builder-nixpacks:report node-js-app
```

```
=====> node-js-app builder-nixpacks information
       Builder-nixpacks computed nixpackstoml path: nixpacks2.toml
       Builder-nixpacks global nixpackstoml path:   nixpacks.toml
       Builder-nixpacks nixpackstoml path:          nixpacks2.toml
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku builder-nixpacks:report node-js-app --builder-nixpacks-nixpackstoml-path
```

```
nixpacks2.toml
```
