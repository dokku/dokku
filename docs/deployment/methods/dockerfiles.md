# Dockerfile Deployment

> New as of 0.3.15

While Dokku normally defaults to using [Heroku buildpacks](https://devcenter.heroku.com/articles/buildpacks) for deployment, you can also use Docker's native `Dockerfile` system to define a container.

> Dockerfile support is considered a *power user* feature. By using Dockerfile-based deployment, you agree that you will not have the same comfort as that enjoyed by buildpack users, and Dokku features may work differently. Differences between the two systems will be documented here.

To use a Dockerfile for deployment, commit a valid `Dockerfile` to the root of your repository and push the repository to your Dokku installation. If this file is detected, Dokku will default to using it to construct containers *except* in the following two cases:

- The application has a `BUILDPACK_URL` environment variable set via the `dokku config:set` command or in a committed `.env` file. In this case, Dokku will use your specified buildpack.
- The application has a `.buildpacks` file in the root of the repository. In this case, Dokku will use your specified buildpack(s).

## Switching from buildpack deployments

If an application was previously deployed via buildpacks, the following commands should be run before a Dockerfile deploy will succeed:

```shell
dokku config:unset --no-restart node-js-app DOKKU_PROXY_PORT_MAP 
```

## Build-time configuration variables

For security reasons - and as per [Docker recommendations](https://github.com/docker/docker/issues/13490) - Dockerfile-based deploys have variables available only during runtime.

For users that require customization in the `build` phase, you may use build arguments via the [docker-options plugin](docs/advanced-usage/docker-options.md):

```shell
dokku docker-options:add node-js-app build '--file Dockerfile.dokku' # e.g. to use alternate Dockerfile
dokku docker-options:add node-js-app build '--build-arg NODE_ENV=production'
```

Once set, the Dockerfile usage would be as follows:

```Dockerfile
FROM debian:jessie

# set the argument default
ARG NODE_ENV=production

# use the argument
RUN echo $NODE_ENV
```

You may also set the argument as an environment variable

```Dockerfile
FROM debian:jessie

# set the argument default
ARG NODE_ENV=production

# assign it to an environment variable
# we can wrap the variable in brackets
ENV NODE_ENV ${NODE_ENV}

# or omit them completely

# use the argument
RUN echo $NODE_ENV
```

## Customizing the run command

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
FROM debian:jessie
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

## Exposed ports

See the [port management documentation](/docs/networking/port-management.md).
