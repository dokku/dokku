# Persistent Storage

> New as of 0.5.0

The preferred method to mount external containers to a Dokku managed container, is to use the Dokku storage plugin.


```
storage:list <app>                             # List bind mounts for app's container(s) (host:container)
storage:mount <app> <host-dir:container-dir>   # Create a new bind mount
storage:report [<app>] [<flag>]                # Displays a checks report for one or more apps
storage:unmount <app> <host-dir:container-dir> # Remove an existing bind mount
```

> The storage plugin is compatible with storage mounts created with the docker-options. The storage plugin will only list mounts from the deploy/run phase.

The storage plugin supports the following mount points:

- explicit paths that exist on the host
- docker volumes

There are a few caveats to using the persistent storage plugin:

- Using implicit paths that do not exist are no longer supported, and actually deprecated in [Docker 1.9.0](https://github.com/docker/docker/releases/tag/v1.9.0). When you specify a persistent storage mount, the host directory is **not** autocreated by either Dokku or Docker.
- We recommend using the directory `/var/lib/dokku/data/storage` directory as the root host path for mounts, and we create this on Dokku installation.
- Mounts are only available at run and deploy times, and you **must** redeploy (restart) an app to mount or unmount to an existing app's container.
- When a directory is mounted, any existing files within the container will be overwritten. If you are writing assets during the build process and then replace the directory with a mount, those files will no longer exist. This is a Docker limitation.
- Paths are mounted within the container at the root of the disk - `/` - and are **not** relative to `/app` (for buildpacks deploys) or the `WORKDIR` (for Dockerfile/Docker images).
- For applications using buildpack deploys, the host directory should be owned by the user and group id `32767`. This is due to how permissions within Herokuish - which builds the Docker images - works. For Dockerfile or Docker image deployments, please use the user and group id which corresponds to the one running the process within the container.

## Usage

This example demonstrates how to mount the recommended directory to `/storage` inside an application called `node-js-app`:

```shell
# we use a subdirectory inside of the host directory to scope it to just the app
dokku storage:mount node-js-app /var/lib/dokku/data/storage/node-js-app:/storage
```

Dokku will then mount the shared contents of `/var/lib/dokku/data/storage` to `/storage` inside the container.

Once you have mounted persistent storage, you will also need to restart the application. See the
[process scaling documentation](/docs/deployment/process-management.md) for more information.

```shell
dokku ps:restart app-name
```

A more complete workflow may require making a custom directory for your application and mounting it within your `/app/storage` directory instead. The mount point is *not* relative to your application's working directory, and is instead relative to the root of the container.

```shell
# creating storage for the app 'node-js-app'
mkdir -p  /var/lib/dokku/data/storage/node-js-app

# ensure the proper user has access to this directory
chown -R dokku:dokku /var/lib/dokku/data/storage/node-js-app

# as of 0.7.x, you should chown using the `32767` user and group id for buildpack deploys
# For dockerfile deploys, substitute the user and group id in use within the image
chown -R 32767:32767 /var/lib/dokku/data/storage/node-js-app

# mount the directory into your container's /app/storage directory, relative to root
dokku storage:mount app-name /var/lib/dokku/data/storage/node-js-app:/app/storage
```

You can mount one or more directories as desired by following the above pattern.

### Displaying storage reports for an app

> New as of 0.8.1

You can get a report about the app's storage status using the `storage:report` command:

```shell
dokku storage:report
```

```
=====> node-js-app storage information
       Storage build mounts:
       Storage deploy mounts: -v /var/lib/dokku/data/storage/node-js-app:/app/storage
       Storage run mounts:  -v /var/lib/dokku/data/storage/node-js-app:/app/storage
=====> python-sample storage information
       Storage build mounts:
       Storage deploy mounts:
       Storage run mounts:
=====> ruby-sample storage information
       Storage build mounts:
       Storage deploy mounts:
       Storage run mounts:
```

You can run the command for a specific app also.

```shell
dokku storage:report node-js-app
```

```
=====> node-js-app storage information
       Storage build mounts:
       Storage deploy mounts: -v /var/lib/dokku/data/storage/node-js-app:/app/storage
       Storage run mounts:  -v /var/lib/dokku/data/storage/node-js-app:/app/storage
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku storage:report node-js-app --storage-deploy-mounts
```

## Use Cases

### Persistent storage

Dokku is powered by Docker containers, which recommends in their [best practices](https://docs.docker.com/engine/userguide/eng-image/dockerfile_best-practices/#containers-should-be-ephemeral) that containers be treated as ephemeral. In order to manage persistent storage for web applications, like user uploads or large binary assets like images, a directory outside the container should be mounted.

### Shared storage between containers

When scaling your app, you may require a common location to access shared assets between containers, a storage mount can be used in this situation.

### Shared storage across environments

Your app may be used in a cluster that requires containers or resources not running on the same host access your data. Mounting a shared file service (like S3FS or EFS) inside your container will give you great flexibility.

### Backing up

Your app may have services that are running in memory and need to be backed up locally (like a key store). Mount a non ephemeral storage mount will allow backups that are not lost when the app is shut down.

### Build phase

By default, Dokku will only bind storage mounts during the deploy and run phases. Under certain conditions, one might want to bind a storage mount during the build phase. This can be accomplished by using the `docker-options` plugin directly.

```shell
dokku docker-options:add node-js-app build "-v /tmp/python-test:/opt/test"
```

You cannot use mounted volumes during the build phase of a Dockerfile deploy. This is because Docker does not support volumes when executing `docker build`.

> Note: **This can cause data loss** if you bind a mount under `/app` in buildpack apps as herokuish will attempt to remove the original app path during the build phase.

## Application User and Persistent Storage file ownership (buildpack apps only)

> New as of 0.7.1

By default, Dokku will execute your buildpack application processes as the `herokuishuser` user. You may override this by setting the `DOKKU_APP_USER` config variable.

> NOTE: this user must exist in your herokuish image.

Additionally, the default `docker-local` scheduler that comes with Dokku will ensure your storage mounts are owned by either `herokuishuser` or the overridden value you have set in `DOKKU_APP_USER`. See the [docker-local scheduler documentation](/docs/advanced-usage/schedulers/docker-local.md#disabling-chown-of-persistent-storage) docs for more information.
