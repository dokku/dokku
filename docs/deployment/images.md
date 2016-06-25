# Image Tagging

> New as of 0.4.0

The dokku tags plugin allows you to add docker image tags to the currently deployed app image for versioning and subsequent deployment.

```
tags <app>                                       List all app image tags
tags:create <app> <tag>                          Add tag to latest running app image
tags:deploy <app> <tag>                          Deploy tagged app image
tags:destroy <app> <tag>                         Remove app image tag
```

Example:

```
$ dokku tags node-js-app
=====> Image tags for dokku/node-js-app
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
dokku/node-js-app   latest              936a42f25901        About a minute ago   1.025 GB

$ dokku tags:create node-js-app v0.9.0
=====> Added v0.9.0 tag to dokku/node-js-app

$ dokku tags node-js-app
=====> Image tags for dokku/node-js-app
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
dokku/node-js-app   latest              936a42f25901        About a minute ago   1.025 GB
dokku/node-js-app   v0.9.0              936a42f25901        About a minute ago   1.025 GB

$ dokku tags:deploy node-js-app v0.9.0
-----> Releasing node-js-app (dokku/node-js-app:v0.9.0)...
-----> Deploying node-js-app (dokku/node-js-app:v0.9.0)...
-----> Running pre-flight checks
       For more efficient zero downtime deployments, create a file CHECKS.
       See http://dokku.viewdocs.io/dokku/checks-examples.md for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
=====> node-js-app container output:
       Detected 512 MB available memory, 512 MB limit per process (WEB_MEMORY)
       Recommending WEB_CONCURRENCY=1
       > node-js-app@0.1.0 start /app
       > node index.js
       Node app is running at localhost:5000
=====> end node-js-app container output
-----> Running post-deploy
-----> Configuring node-js-app.dokku.me...
-----> Creating http nginx.conf
-----> Running nginx-pre-reload
       Reloading nginx
-----> Shutting down old containers in 60 seconds
=====> 025eec3fa3b442fded90933d58d8ed8422901f0449f5ea0c23d00515af5d3137
=====> Application deployed:
       http://node-js-app.dokku.me

```

## Deploying from a Docker Registry

You can alternatively add image pulled from a docker Registry and deploy app from it by using tagging feature. In this example, we are deploying from Docker Hub.

1. Create dokku app as usual

    ```shell
    dokku apps:create test-app
    ```

2. Pull image from Docker Hub

    ```shell
    docker pull demo-repo/some-image:v12
    ```

3. Retag the image to match the created app

    ```shell
    docker tag demo-repo/some-image:v12 dokku/test-app:v12
    ```

4. Deploy tag

    ```shell
    dokku tags:deploy test-app v12
    ```

> Note: When deploying an image, we will use `docker inspect` to extract the `ExposedPorts` configuration and if defined, use that to populate `DOKKU_DOCKERFILE_PORTS`. If this behavior is not desired, you can override that configuration variable with the `config:set` command.
> Example: `dokku config:set test-app DOKKU_DOCKERFILE_PORTS="5984/tcp 80/tcp"`

## Deploying an image from CI

To ensure your builds are always reproducible, it's considered bad practice to store build
artifacts in your repository. For some projects however, building artifacts during deployment
to dokku may affect the performance of running applications.

One solution is to build a finished Docker image on a CI service (or even locally) and deploy
it directly to the host running dokku.

1. Build image on CI (or locally)

    ```shell
    docker build -t dokku/test-app:v12
    # Note: The image must be tagged `dokku/<app-name>:<version>`
    ```

2. Deploy image to dokku host

    ```shell
    docker save dokku/test-app:v12 | ssh my.dokku.host "docker load | dokku tags:deploy test-app v12"
    ```

> Note: You can also use a Docker Registry to push and pull
> the image rather than uploading it directly.

Here's a more complete example using the above method:

```shell
# build the image
docker build -t dokku/test-app:v12
# copy the image to the dokku host
docker save dokku/test-app:v12 | bzip2 | ssh my.dokku.host "bunzip2 | docker load"
# tag and deploy the image
ssh my.dokku.host "dokku tags:create test-app previous; dokku tags:deploy test-app v12 && dokku tags:create test-app latest"
```
