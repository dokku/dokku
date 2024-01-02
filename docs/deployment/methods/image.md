# Docker Image Deployment

## Initializing an app repository from a docker image

> [!IMPORTANT]
> New as of 0.24.0

A Dokku app repository can be initialized or updated from a Docker image via the `git:from-image` command. This command will either initialize the app repository or update it to include the specified Docker image via a `FROM` stanza. This is an excellent way of tracking changes when deploying only a given docker image, especially if deploying an image from a remote CI/CD pipeline.

```shell
dokku git:from-image node-js-app my-registry/node-js-getting-started:latest
```

In the above example, Dokku will build the app as if the repository contained _only_ a `Dockerfile` with the following content:

```Dockerfile
FROM my-registry/node-js-getting-started:latest
```

If the specified image already exists on the Dokku host, it will not be pulled again, though this behavior may be changed using [build phase docker-options](/docs/advanced-usage/docker-options.md).

Triggering a build with the same arguments multiple times will result in Dokku exiting `0` early as there will be no changes detected. If the image tag is reused but the underlying image is different, it is recommended to use the image digest instead of the tag. This can be retrieved via the following command:

```shell
# for images pushed to a remote registry
docker inspect --format='{{index .RepoDigests 0}}' $IMAGE_NAME

# for images built locally and not pushed to a registry
# use when the previous command output is empty
docker images --no-trunc --quiet $IMAGE_NAME
```

The resulting `git:from-image` call would then be:

```shell
# where the image sha is: sha256:9d187c3025d03c033dcc71e3a284fee53be88cc4c0356a19242758bc80cab673
dokku git:from-image node-js-app my-registry/node-js-getting-started@sha256:9d187c3025d03c033dcc71e3a284fee53be88cc4c0356a19242758bc80cab673
```

The `git:from-image` command can optionally take a git `user.name` and `user.email` argument (in that order) to customize the author. If the arguments are left empty, they will fallback to `Dokku` and `automated@dokku.sh`, respectively.

```shell
dokku git:from-image node-js-app my-registry/node-js-getting-started:latest "Camila" "camila@example.com"
```

If the image is a private image that requires a docker login to access, the `registry:login` command should be used to log into the registry. See the [registry documentation](/docs/advanced-usage/registry-management.md#logging-into-a-registry) for more details on this process.

Finally, certain images may require a custom build context in order for `ONBUILD ADD` and `ONBUILD COPY` statements to succeed. A custom build context can be specified via the `--build-dir` flag. All files in the specified `build-dir` will be copied into the repository for use within the `docker build` process. The build context _must_ be specified on each deploy, and is not otherwise persisted between builds.

```shell
dokku git:from-image --build-dir path/to/build node-js-app domy-registrykku/node-js-getting-started:latest "Camila" "camila@example.com"
```

See the [dockerfile documentation](/docs/deployment/builders/dockerfiles.md) to learn about the different ways to configure Dockerfile-based deploys.

## Initializing an app repository from a remote image without a registry

> [!IMPORTANT]
> New as of 0.30.0

A Dokku app repository can be initialized or updated from the contents of an image archive tar file via the `git:load-image` command. This method can be used when a Docker Registry is unavailable to act as an intermediary for storing an image, such as when building an image in CI and deploying directly from that image.

```shell
docker image save my-registry/node-js-getting-started:latest | ssh dokku@dokku.me dokku git:load-image node-js-app my-registry/node-js-getting-started:latest
```

In the above example, we are saving the image to a tar file via `docker image save`, streaming that to the Dokku host, and then running `git:load-image` on the incoming stream. Dokku will build the app as if the repository contained _only_ a `Dockerfile` with the following content:

```Dockerfile
FROM my-registry/node-js-getting-started:latest
```

When deploying an app via `git:load-image`, it is highly recommended to use a unique image tag when building the image. Not doing so will result in Dokku exiting `0` early as there will be no changes detected. If the image tag is reused but the underlying image is different, it is recommended to use the image digest instead of the tag. This can be retrieved via the following command:

```shell

# for images pushed to a remote registry
docker inspect --format='{{index .RepoDigests 0}}' $IMAGE_NAME

# for images built locally and not pushed to a registry
# use when the previous command output is empty
docker images --no-trunc --quiet $IMAGE_NAME
```

The resulting `git:load-image` call would then be:

```shell
# where the image sha is: sha256:9d187c3025d03c033dcc71e3a284fee53be88cc4c0356a19242758bc80cab673
docker image save my-registry/node-js-getting-started:latest | ssh dokku@dokku.me dokku git:load-image node-js-app my-registry/node-js-getting-started@sha256:9d187c3025d03c033dcc71e3a284fee53be88cc4c0356a19242758bc80cab673
```

The `git:load-image` command can optionally take a git `user.name` and `user.email` argument (in that order) to customize the author. If the arguments are left empty, they will fallback to `Dokku` and `automated@dokku.sh`, respectively.

```shell
docker image save my-registry/node-js-getting-started:latest | ssh dokku@dokku.me dokku git:load-image node-js-app my-registry/node-js-getting-started:latest "Camila" "camila@example.com"
```

Finally, certain images may require a custom build context in order for `ONBUILD ADD` and `ONBUILD COPY` statements to succeed. A custom build context can be specified via the `--build-dir` flag. All files in the specified `build-dir` will be copied into the repository for use within the `docker build` process. The build context _must_ be specified on each deploy, and is not otherwise persisted between builds.

```shell
docker image save my-registry/node-js-getting-started:latest | ssh dokku@dokku.me dokku git:load-image --build-dir path/to/build node-js-app my-registry/node-js-getting-started:latest "Camila" "camila@example.com"
```

See the [dockerfile documentation](/docs/deployment/builders/dockerfiles.md) to learn about the different ways to configure Dockerfile-based deploys.
