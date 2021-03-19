# Docker Image Tag Deployment
----

!!! warning
    As of 0.24.0, this functionality is deprecated in favor of the [`git:from-image`](/deployment/methods/git#initializing-an-app-repository-from-a-docker-image) command. It will be removed in a future release, and is considered unmaintained. Users are highly encouraged to switch their workflows to `git:from-image`.

!!! tip "New as of 0.4.0"

```
tags <app>                                     # List all app image tags
tags:create <app> <tag>                        # Add tag to latest running app image
tags:deploy <app> <tag>                        # Deploy tagged app image
tags:destroy <app> <tag>                       # Remove app image tag
```

The Dokku tags plugin allows you to add Docker image tags to the currently deployed app image for versioning and subsequent deployment.

!!! info
    When triggering `dokku ps:rebuild APP` on an application deployed via the `tags` plugin, the following may occur:

    - Applications previously deployed via another method (`git`/`tar`): The application may revert to a state before the latest custom image tag was deployed.
    - Applications that were only ever deployed via the `tags` plugin: No action will be taken against your application.

    Please use the `tags:deploy` command when redeploying an application deployed via Docker image.

## Usage

### Exposed ports

See the [port management documentation](/networking/port-management) for more information on how Dokku exposes ports for applications and how you can configure these for your app.

### Listing tags for an application

For example, you can list all tags for a given application:

```shell
dokku tags node-js-app
```

```shell-session
=====> Image tags for dokku/node-js-app
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
dokku/node-js-app   latest              936a42f25901        About a minute ago   1.025 GB
```

### Creating a tag

You can also create new tags for that app using the `tags:create` function. Tags should conform to the Docker tagging specification for your Docker version. As of 1.10, that specification is available [here](https://github.com/docker/docker/blob/master/image/spec/v1.1), while users of older versions can check the documentation [here](https://github.com/docker/docker/blob/master/image/spec/v1).

```shell
dokku tags:create node-js-app v1
```

```shell-session
=====> Added v1 tag to dokku/node-js-app
```

Once the tag is created, you can see the output by running the `tags` command again.

```shell
dokku tags node-js-app
```

```shell-session
=====> Image tags for dokku/node-js-app
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
dokku/node-js-app   latest              936a42f25901        About a minute ago   1.025 GB
dokku/node-js-app   v1                  936a42f25901        About a minute ago   1.025 GB
```

### Deploying an image tag

Finally, you can also deploy a local image using the `tags:deploy` command. When specifying a tag that is not `latest`, the released image will be retagged as the `latest` image tag for the app.

!!! warning
    For images based on Herokuish, using the `tags:deploy` command will reset environment variables written into the image, causing a retag to occur. This will - on average - add two extra layers to your deployed image. Note that this does not affect Dockerfile-based images, which are the majority of images deployed via the `tags` command.

```shell
dokku tags:deploy node-js-app v1
```

```shell-session
-----> Releasing node-js-app (dokku/node-js-app:v1)...
-----> Deploying node-js-app (dokku/node-js-app:v1)...
-----> Running pre-flight checks
       For more efficient zero downtime deployments, create a file CHECKS.
       See https://dokku.com/docs/deployment/zero-downtime-deploys/ for examples
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
       Reloading nginx
-----> Shutting down old containers in 60 seconds
=====> 025eec3fa3b442fded90933d58d8ed8422901f0449f5ea0c23d00515af5d3137
=====> Application deployed:
       http://node-js-app.dokku.me
```

## Image workflows

### Deploying from a Docker registry

You can alternatively add image pulled from a Docker registry and deploy from it by using tagging feature. In this example, we are deploying from Docker Hub.

1. Create Dokku app as usual.

    ```shell
    dokku apps:create test-app
    ```

2. Pull image from Docker Hub.

    ```shell
    docker pull demo-repo/some-image:v12
    ```

3. Retag the image to match the created app.

    ```shell
    docker tag demo-repo/some-image:v12 dokku/test-app:v12
    ```

4. Deploy tag.

    ```shell
    dokku tags:deploy test-app v12
    ```

### Deploying an image from CI

To ensure your builds are always reproducible, it's considered bad practice to store build
artifacts in your repository. For some projects however, building artifacts during deployment
to Dokku may affect the performance of running applications.

One solution is to build a finished Docker image on a CI service (or even locally) and deploy
it directly to the host running Dokku.

1. Build image on CI (or locally).

    ```shell
    docker build -t dokku/test-app:v12 .
    # Note: The image must be tagged `dokku/<app-name>:<version>`
    ```

2. Deploy image to Dokku host.

    ```shell
    docker save dokku/test-app:v12 | ssh my.dokku.host "docker load | dokku tags:deploy test-app v12"
    ```

!!! note
    You can also use a Docker registry to push and pull
    the image rather than uploading it directly.

Here's a more complete example using the above method:

```shell
# build the image
docker build -t dokku/test-app:v12 .
# copy the image to the dokku host
docker save dokku/test-app:v12 | bzip2 | ssh my.dokku.host "bunzip2 | docker load"
# tag and deploy the image
ssh my.dokku.host "dokku tags:create test-app previous; dokku tags:deploy test-app v12"
```

## Related articles
- [Setting up persistent storage](/advanced-usage/persistent-storage)
- [Defining environment variables](/configuration/environment-variables)
- [Setting up the ports](/networking/proxy-management)
