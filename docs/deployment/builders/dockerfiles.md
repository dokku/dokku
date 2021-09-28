# Dockerfile Deployment

> New as of 0.3.15

While Dokku normally defaults to using [Heroku buildpacks](https://devcenter.heroku.com/articles/buildpacks) for deployment, you can also use Docker's native `Dockerfile` system to define a container.

> Dockerfile support is considered a *power user* feature. By using Dockerfile-based deployment, you agree that you will not have the same comfort as that enjoyed by buildpack users, and Dokku features may work differently. Differences between the two systems will be documented here.

## Usage

### Detection

This builder will be auto-detected in the following case:

- A `Dockerfile` exists in the root of the app repository.

Dokku will only select the `dockerfile` builder if both the `herokuish` and `pack` builders are not detected and a Dockerfile exists. For more information on how those are detected, see the following links:

- [Cloud Native Buildpacks documentation](/docs/deployment/builders/cloud-native-buildpacks.md#detection)
- [Herokuish documentation](/docs/deployment/builders/herokuish-buildpacks.md#detection)

### Switching from buildpack deployments

If an application was previously deployed via buildpacks, the following commands should be run before a Dockerfile deploy will succeed:

```shell
dokku config:unset --no-restart node-js-app DOKKU_PROXY_PORT_MAP 
```

### Changing the `Dockerfile` location

> The previous method to perform this - via `docker-options:add` - should be removed in favor of the `builder-dockerfile:set` command outlined here.

When deploying a monorepo, it may be desirable to specify the specific path of the `Dockerfile` file to use for a given app. This can be done via the `builder-dockerfile:set` command. If a value is specified and that file does not exist in the app's build directory, then the build will fail.

```shell
dokku builder-dockerfile:set node-js-app dockerfile-path Dockerfile2
```

The default value may be set by passing an empty value for the option:

```shell
dokku builder-dockerfile:set node-js-app dockerfile-path
```

The `dockerfile-path` property can also be set globally. The global default is `Dockerfile`, and the global value is used when no app-specific value is set.

```shell
dokku builder-dockerfile:set --global dockerfile-path Dockerfile2
```

The default value may be set by passing an empty value for the option.

```shell
dokku builder-dockerfile:set --global dockerfile-path
```

### Displaying builder-dockerfile reports for an app

> New as of 0.25.0

You can get a report about the app's storage status using the `builder-dockerfile:report` command:

```shell
dokku app-json:report
```

```
=====> node-js-app builder-dockerfile information
       Builder-dockerfile computed dockerfile path: Dockerfile2
       Builder-dockerfile global dockerfile path:   Dockerfile
       Builder-dockerfile dockerfile path:          Dockerfile2
=====> python-sample builder-dockerfile information
       Builder-dockerfile computed dockerfile path: Dockerfile
       Builder-dockerfile global dockerfile path:   Dockerfile
       Builder-dockerfile dockerfile path:
=====> ruby-sample builder-dockerfile information
       Builder-dockerfile computed dockerfile path: Dockerfile
       Builder-dockerfile global dockerfile path:   Dockerfile
       Builder-dockerfile dockerfile path:
```

You can run the command for a specific app also.

```shell
dokku builder-dockerfile:report node-js-app
```

```
=====> node-js-app builder-dockerfile information
       Builder-dockerfile computed dockerfile path: Dockerfile2
       Builder-dockerfile global dockerfile path:   Dockerfile
       Builder-dockerfile dockerfile path:          Dockerfile2
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku builder-dockerfile:report node-js-app --builder-dockerfile-dockerfile-path
```

```
Dockerfile2
```

### Build-time configuration variables

For security reasons - and as per [Docker recommendations](https://github.com/docker/docker/issues/13490) - Dockerfile-based deploys have variables available only during runtime.

For users that require customization in the `build` phase, you may use build arguments via the [docker-options plugin](docs/advanced-usage/docker-options.md):

```shell
dokku docker-options:add node-js-app build '--build-arg NODE_ENV=production'
```

Once set, the Dockerfile usage would be as follows:

```Dockerfile
FROM ubuntu:18.04

# set the argument default
ARG NODE_ENV=production

# use the argument
RUN echo $NODE_ENV
```

You may also set the argument as an environment variable

```Dockerfile
FROM ubuntu:18.04

# set the argument default
ARG NODE_ENV=production

# assign it to an environment variable
# we can wrap the variable in brackets
ENV NODE_ENV ${NODE_ENV}

# or omit them completely

# use the argument
RUN echo $NODE_ENV
```

### Building images with Docker Buildkit

If your Dockerfile is using Docker engine's [buildkit](https://docs.docker.com/develop/develop-images/build_enhancements/) (not to be confused with buildpacks), then the `DOCKER_BUILDKIT=1` environment variable needs to be set. Additionally, complete build log output can be forced via `BUILDKIT_PROGRESS=plain`. Both of these environment variables can be set as follows:

```shell
echo "export DOCKER_BUILDKIT=1" | sudo tee -a /etc/default/dokku
echo "export BUILDKIT_PROGRESS=plain" | sudo tee -a /etc/default/dokku
```

#### Buildkit directory caching

Buildkit implements the `RUN --mount` option, enabling mount directory caches for `RUN` directives. The following is an example that mounts debian packaging related directories, which can speed up fetching of remote package data.

```Dockerfile
FROM debian:latest
RUN --mount=target=/var/lib/apt/lists,type=cache \
    --mount=target=/var/cache/apt,type=cache \
    apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      git
```

Mount cache targets may vary depending on the tool in use, and users are encouraged to investigate the directories that apply for their language and framework.

You would adjust the cache directory for whatever application cache you have, e.g. `/root/.pnpm-store/v3` for pnpm, `$HOME/.m2` for maven, or `/root/.cache` for golang.

### Customizing the run command

By default no arguments are passed to `docker run` when deploying the container and the `CMD` or `ENTRYPOINT` defined in the `Dockerfile` are executed. You can take advantage of docker ability of overriding the `CMD` or passing parameters to your `ENTRYPOINT` setting `$DOKKU_DOCKERFILE_START_CMD`. Let's say for example you are deploying a base Node.js image, with the following `ENTRYPOINT`:

```Dockerfile
ENTRYPOINT ["node"]
```

You can do:

```shell
dokku config:set node-js-app DOKKU_DOCKERFILE_START_CMD="--harmony server.js"
```

To tell Docker what to run.

Setting `$DOKKU_DOCKERFILE_CACHE_BUILD` to `true` or `false` will enable or disable Docker's image layer cache. Lastly, for more granular build control, you may also pass any `docker build` option to `docker`, by setting `$DOKKU_DOCKER_BUILD_OPTS`.

### Procfiles and multiple processes

> New as of 0.5.0

You can also customize the run command using a `Procfile`, much like you would on Heroku or
with a buildpack deployed app. The `Procfile` should contain one or more lines defining [process types and associated commands](https://devcenter.heroku.com/articles/procfile#declaring-process-types).
When you deploy your app, a Docker image will be built. The `Procfile` will be extracted from the image
(it must be in the folder defined in your `Dockerfile` as `WORKDIR` or `/app`) and the commands
in it will be passed to `docker run` to start your process(es). Here's an example `Procfile`:

```Procfile
web: bin/run-prod.sh
worker: bin/run-worker.sh
```

And `Dockerfile`:

```Dockerfile
FROM ubuntu:18.04
WORKDIR /app
COPY . ./
CMD ["bin/run-dev.sh"]
```

When you deploy this app the `web` process will automatically be scaled to 1 and your Docker container
will be started basically using the command `docker run bin/run-prod.sh`. If you want to also run
a worker container for this app, you can run `dokku ps:scale worker=1` and a new container will be
started by running `docker run bin/run-worker.sh` (the actual `docker run` commands are a bit more
complex, but this is the basic idea). If you use an `ENTRYPOINT` in your `Dockerfile`, the lines
in your `Procfile` will be passed as arguments to the `ENTRYPOINT` script instead of being executed.

### Exposed ports

See the [port management documentation](/docs/networking/port-management.md) for more information on how Dokku exposes ports for applications and how you can configure these for your app.
