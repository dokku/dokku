# Dockerfile Deployment

> New as of 0.4.3

As an extension of Dockerfile-based deployment, you can also deploy an application that uses `docker-compose`. This is useful as an alternative to running a process manager inside of a `Dockerfile`-based container, and should be considered the canonical way to deploy multiple processes when using a `Dockerfile` directly.

To deploy an application via `docker-compose`, your repository will need the following files:

- `Dockerfile`
- `docker-compose.yml`

To use `docker-compose` for deployment, commit a valid `Dockerfile` **and** `docker-compose.yml` to the root of your repository and push the repository to your Dokku installation. If these files are detected, Dokku will default to using it to construct containers **except** in the following two cases:

- The application has a `BUILDPACK_URL` environment variable set via the `dokku config:set` command or in a committed `.env` file. In this case, Dokku will use your specified buildpack.
- The application has a `.buildpacks` file in the root of the repository. In this case, Dokku will use your specified buildpack(s).

At this time, you cannot fallback to using `Dockerfile` only deployments.

## Web processes

At this time, you may only have **one** web process proxied by nginx. Having any other processes will require manual configuration of your nginx configuration.

The web process is noted as either the `app` or `web` stanza in your `docker-compose.yml` file, with the `web` stanza having precedence.

Your application *must* have either an `app` or `web` stanza or the `docker-compose` deployment will fail.

## Exposed ports

By default, Dokku will extract the first `EXPOSE` tcp port and use said port with nginx to proxy your app to that port. For applications that have multiple ports exposed, you may override this port via the following command:

```shell
# replace APP with the name of your application
dokku config:set APP DOKKU_DOCKERFILE_PORT=8000
```

Dokku will not expose other ports on your application without a [custom docker-option](/dokku/configuration/docker-options/).

If you do not have a port explicitly exposed, Dokku will automatically expose port `5000` for your application.

## Dependent images

Dokku will automatically use the images specified for each process-type specified in a `docker-compose.yml` file, and will fall back to the image created by either the `app` or `web` process-type. Non-primary process types will be scaled to 0 by default.

## Limitations

`docker-compose` support has a few known limitations. The below is not considered exhaustive, and may change over time.

- Container `links` do not work. When Dokku starts containers, the containers are named pseudo-randomly, and due to how zero-downtime deployments work, we cannot guarantee that any started links will start properly. Docker also cannot link already started containers, and thus this functionality is currently unused. For datastores, please use the [official datastore plugins](/dokku/plugins/#official-plugins-beta).
- Dokku uses a special `Procfile.compose` file to maintain a mapping of the process type to docker image. Please avoid using this file for other purposes.

## Process scaling

Process scaling works the same as it does for `buildpack` deployments. See the [process management documentation](/dokku/process-management/).
