# Lambda Builder

> [!IMPORTANT]
> New as of 0.28.0

```
builder-lambda:report [<app>] [<flag>]   # Displays a builder-lambda report for one or more apps
builder-lambda:set <app> <key> (<value>) # Set or clear a builder-lambda property for an app
```

The `lambda` builder builds AWS Lambda functions in an environment simulating AWS Lambda runtimes via [lambda-builder](https://github.com/dokku/lambda-builder). Apps built via this builder can run natively in Dokku and may also have their artifacts scheduled to Lambda via an appropriate scheduler.

## Usage

### Detection

This builder will be auto-detected in the following case:

- A `lambda.yml` exists in the root of the app repository.

The builder may also be selected via the `builder:set` command

```shell
dokku builder:set node-js-app selected lambda
```

### Supported languages

The `lambda` builder plugin supports the following AWS runtime languages on Amazon Linux 2:

- dotnet
- go (builder is based on AL1)
- nodejs
- python
- ruby

See the [lambda-builder](https://github.com/dokku/lambda-builder#how-does-it-work) documentation for more information on how specific languages are detected.

### Customizing the build environment

The `lambda` builder plugin delegates all build logic to [lambda-builder](https://github.com/dokku/lambda-builder), including language detection and build/runtime environment specification. The lambda-builder tool supports a `lambda.yml` file format for customizing how apps are built. Please see the readme for lambda-builder for more information on different options.

### Changing the `lambda.yml` location

The `lambda.yml` is expected to be found in a specific directory, depending on the deploy approach:

- The `WORKDIR` of the Docker image for deploys resulting from `git:from-image` and `git:load-image` commands.
- The root of the source code tree for all other deploys (git push, `git:from-archive`, `git:sync`).

Sometimes it may be desirable to set a different path for a given app, e.g. when deploying from a monorepo. This can be done via the `lambdayml-path` property:

```shell
dokku builder-lambda:set node-js-app lambdayml-path .dokku/lambda.yml
```

The value is the path to the desired file *relative* to the base search directory, and will never be treated as absolute paths in any context. If that file does not exist within the repository, the build will fail.

The default value may be set by passing an empty value for the option:

```shell
dokku builder-lambda:set node-js-app lambdayml-path
```

The `lambdayml-path` property can also be set globally. The global default is `lambda.yml`, and the global value is used when no app-specific value is set.

```shell
dokku builder-lambda:set --global lambdayml-path lambda2.yml
```

The default value may be set by passing an empty value for the option.

```shell
dokku builder-lambda:set --global lambdayml-path
```

### Displaying builder-lambda reports for an app

You can get a report about the app's storage status using the `builder-lambda:report` command:

```shell
dokku builder-lambda:report
```

```
=====> node-js-app builder-lambda information
       Builder lambda computed lambdayml path: lambda2.yml
       Builder lambda global lambdayml path:   lambda.yml
       Builder lambda lambdayml path:          lambda2.yml
=====> python-sample builder-lambda information
       Builder lambda computed lambdayml path: lambda.yml
       Builder lambda global lambdayml path:   lambda.yml
       Builder lambda lambdayml path:
=====> ruby-sample builder-lambda information
       Builder lambda computed lambdayml path: lambda.yml
       Builder lambda global lambdayml path:   lambda.yml
       Builder lambda lambdayml path:
```

You can run the command for a specific app also.

```shell
dokku builder-lambda:report node-js-app
```

```
=====> node-js-app builder-lambda information
       Builder lambda computed lambdayml path: lambda2.yml
       Builder lambda global lambdayml path:   lambda.yml
       Builder lambda lambdayml path:          lambda2.yml
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku builder-lambda:report node-js-app --builder-lambda-lambdayml-path
```

```
lambda2.yml
```
