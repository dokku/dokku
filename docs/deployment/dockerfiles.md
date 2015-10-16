# Dockerfile Deployment

> New as of 0.3.15

While Dokku normally defaults to using [heroku buildpacks](https://devcenter.heroku.com/articles/buildpacks) for deployment, you can also use docker's native `Dockerfile` system to define a container.

To use a dockerfiles for deployment, commit a valid `Dockerfile` to the root of your repository and push the repository to your Dokku installation. If this file is detected, Dokku will default to using it to construct containers **except** in the following two cases:

- The application has a `BUILDPACK_URL` environment variable set via the `dokku config:set` command or in a committed `.env` file. In this case, Dokku will use your specified buildpack.
- The application has a `.buildpacks` file in the root of the repository. In this case, Dokku will use your specified buildpack(s).

## Exposed ports

By default, Dokku will extract the first `EXPOSE` tcp port and use said port with nginx to proxy your app to that port. For applications that have multiple ports exposed, you may override this port via the following command:

```shell
# replace APP with the name of your application
dokku config:set APP DOKKU_DOCKERFILE_PORT=8000
```

Dokku will not expose other ports on your application without a [custom docker-option](/dokku/configuration/docker-options/).

## Customizing the run command

By default no arguments are passed to `docker run` when deploying the container and the `CMD` or `ENTRYPOINT` defined in the `Dockerfile` are executed. You can take advantage of docker ability of overriding the `CMD` or passing parameters to your `ENTRYPOINT` setting `$DOKKU_DOCKERFILE_START_CMD`. Let's say for example you are deploying a base nodejs image, with the following `ENTRYPOINT`:

```
ENTRYPOINT ["node"]
```

You can do:

```
dokku config:set APP DOKKU_DOCKERFILE_START_CMD="--harmony server.js"
```

To tell docker what to run.

Setting `$DOKKU_DOCKERFILE_CACHE_BUILD` to `true` or `false` will enable or disable docker's image layer cache. Lastly, for more granular build control, you may also pass any `docker build` option to `docker`, by setting `$DOKKU_DOCKER_BUILD_OPTS`.
