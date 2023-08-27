# Nixpacks

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

When deploying a monorepo, it may be desirable to specify the specific path of the `nixpacks.toml` file to use for a given app. This can be done via the `builder-nixpacks:set` command. If a value is specified and that file does not exist in the app's build directory, then the build will fail.

```shell
dokku builder-nixpacks:set node-js-app nixpackstoml-path nixpacks2.toml
```

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

Cache is enabled by default, but can be disabled by setting the `no-cache` property to `true`:

```shell
dokku builder-nixpacks:set node-js-app no-cache true
```

The default value may be set by passing an empty value for the option:

```shell
dokku builder-nixpacks:set node-js-app no-cache
```

The `no-cache` property can also be set globally. The global default is `false`, and the global value is used when no app-specific value is set.

```shell
dokku builder-nixpacks:set --global no-cache true
```

The default value may be set by passing an empty value for the option.

```shell
dokku builder-nixpacks:set --global no-cache
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
       Builder-nixpacks computed no cache:          true
       Builder-nixpacks global no cache:            false
       Builder-nixpacks no cache:                   true
=====> python-sample builder-nixpacks information
       Builder-nixpacks computed nixpackstoml path: nixpacks.toml
       Builder-nixpacks global nixpackstoml path:   nixpacks.toml
       Builder-nixpacks nixpackstoml path:
       Builder-nixpacks computed no cache:          false
       Builder-nixpacks global no cache:            false
       Builder-nixpacks no cache:
=====> ruby-sample builder-nixpacks information
       Builder-nixpacks computed nixpackstoml path: nixpacks.toml
       Builder-nixpacks global nixpackstoml path:   nixpacks.toml
       Builder-nixpacks nixpackstoml path:
       Builder-nixpacks computed no cache:          false
       Builder-nixpacks global no cache:            false
       Builder-nixpacks no cache:
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
       Builder-nixpacks computed no cache:          true
       Builder-nixpacks global no cache:            false
       Builder-nixpacks no cache:                   true
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku builder-nixpacks:report node-js-app --builder-nixpacks-no-cache
```

```
true
```
