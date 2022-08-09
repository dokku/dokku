# Builder Management

> New as of 0.24.0

```
builder:report [<app>] [<flag>]   # Displays a builder report for one or more apps
builder:set <app> <key> (<value>) # Set or clear a builder property for an app
```

Builders are a way of customizing how an app is built from a source, allowing users flexibility in how artifacts are created for later scheduling.

## Usage

### Builder selection

Dokku supports the following built-in builders:

- `builder-dockerfile`: Builds apps using a `Dockerfile` via `docker build`. See the [dockerfile builder documentation](/docs/deployment/builders/dockerfiles.md) for more information on how this builder functions.
- `builder-herokuish`: Builds apps with Heroku's v2a Buildpack specification via `gliderlabs/herokuish`. See the [herokuish builder documentation](/docs/deployment/builders/herokuish-buildpacks.md) for more information on how this builder functions.
- `builder-lambda`: Builds AWS Lambda functions in an environment simulating AWS Lambda runtimes via [lambda-builder](https://github.com/dokku/lambda-builder). See the [lambda builder documentation](/docs/deployment/builders/lambda.md) for more information on how this builder functions.
- `builder-null`: Does nothing during the build phase. See the [null builder documentation](/docs/deployment/builders/null.md) for more information on how this builder functions.
- `builder-pack`: Builds apps with Cloud Native Buildpacks via the `pack-cli`  tool. See the [cloud native buildpacks builder documentation](/docs/deployment/builders/cloud-native-buildpacks.md) for more information on how this builder functions.

Builders run a detection script against a source code repository, and the first detected builder will be used to build the app artifact. The exception to this is when a `Dockerfile` is detected and the app is also able to use either `herokuish` or `pack-cli` for building, in which case one of the latter will be chosen.

### Overriding the auto-selected builder

If desired, the builder can be specified via the `builder:set` command by speifying a value for `selected`. The selected builder will always be used.

```shell
dokku builder:set node-js-app selected dockerfile
```

The default value may be set by passing an empty value for the option:

```shell
dokku builder:set node-js-app selected
```

The `selected` property can also be set globally. The global default is an empty string, and auto-detection will be performed when no value is set per-app or globally.

```shell
dokku builder:set --global selected herokuish
```

The default value may be set by passing an empty value for the option.

```shell
dokku builder:set --global selected
```

#### Changing the build directory

> Warning: Please keep in mind that setting a custom build directory will result in loss of any changes to the top-level directory, such as the `git.keep-git-dir` property.

When deploying a monorepo, it may be desirable to specify the specific build directory to use for a given app. This can be done via the `builder:set` command. If a value is specified and that directory does not exist within the repository, the build will fail.

```shell
dokku builder:set node-js-app build-dir app2
```

The default value may be set by passing an empty value for the option:

```shell
dokku builder:set node-js-app build-dir
```

The `build-dir` property can also be set globally. The global default is empty string, and the global value is used when no app-specific value is set.

```shell
dokku builder:set --global build-dir app2
```

The default value may be set by passing an empty value for the option.

```shell
dokku builder:set --global build-dir
```

### Displaying builder reports for an app

You can get a report about the app's builder status using the `builder:report` command:

```shell
dokku builder:report
```

```
=====> node-js-app builder information
       Builder build dir:          custom
       Builder computed build dir: custom
       Builder computed selected:  herokuish
       Builder global build dir:
       Builder global selected: herokuish
       Builder selected: herokuish
=====> python-sample builder information
       Builder build dir:
       Builder computed build dir:
       Builder computed selected: dockerfile
       Builder global build dir:
       Builder global selected: herokuish
       Builder selected: dockerfile
=====> ruby-sample builder information
       Builder build dir:
       Builder computed build dir:
       Builder computed selected: herokuish
       Builder global build dir:
       Builder global selected: herokuish
       Builder selected:
```

You can run the command for a specific app also.

```shell
dokku builder:report node-js-app
```

```
=====> node-js-app builder information
       Builder selected: herokuish
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku builder:report node-js-app --builder-selected
```

### Custom builders

To create a custom builder, the following triggers must be implemented:

- `builder-build`:
  - arguments: `BUILDER_TYPE` `APP` `SOURCECODE_WORK_DIR`
  - description: Creates a docker image named with the output of `common#get_app_image_name $APP`.
- `builder-detect`:
  - arguments: `APP` `SOURCECODE_WORK_DIR`
  - description: Outputs the name of the builder (without the `builder-` prefix) to use to build the app.
- `builder-release`:
  - arguments: `BUILDER_TYPE` `APP` `IMAGE_AG`
  - description: A post-build, pre-release trigger that can be used to post-process the image. Usually simply tags and labels the image appropriately.

Custom plugins names _must_ have the prefix `builder-` or builder overriding via `builder:set` may not function as expected.

Builders can use any tools available on the system to build the docker image, and may even be used to schedule building off-server. The only current requirement is that the image must exist on the server at the end of the `builder-build` command, though this requirement may be relaxed in a future release.

For a simple example of how to implement this trigger, see `builder-pack`, which utilizes a cli tool - `pack-cli` - to generate an OCI image that is compatible with Docker and can be scheduled by the official scheduling plugins.
