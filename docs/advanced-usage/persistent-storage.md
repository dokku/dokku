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

## Usage

This example demonstrates how to mount the recommended directory to `/storage` inside an application called `node-js-app`. For simplicity, the Dokku project recommends using the directory `/var/lib/dokku/data/storage` directory as the root host path for mounts. This directory is created on Dokku installation.

```shell
# we use a subdirectory inside of the host directory to scope it to just the app
dokku storage:mount node-js-app /var/lib/dokku/data/storage/node-js-app:/storage
```

Dokku will then mount the shared contents of `/var/lib/dokku/data/storage/node-js-app` to `/storage` inside the container. Mounts are only available for containers crated via `run` and by the deploy process, and not during the build process. In addition, the host path is never auto-created by either Dokku or Docker, and should be an explicit path, not one relative to the current working directory.

> If the `/storage` path within the container had pre-existing content, the container files will be overrwritten. This may be an issue for users that create assets at build time but then mount a directory at the same place during runtime. Files are not merged.

Once you have mounted persistent storage, you will also need to restart the application. See the
[process scaling documentation](/docs/processes/process-management.md) for more information.

```shell
dokku ps:restart app-name
```

A more complete workflow may require making a custom directory for your application and mounting it within your `/app/storage` directory instead. The mount point is *not* relative to your application's working directory, and is instead relative to the root (`/`) of the container.

```shell
# creating storage for the app 'node-js-app'
mkdir -p  /var/lib/dokku/data/storage/node-js-app

# set the directory ownership. Below is an example for herokuish
# but see the `Directory Permissions` section for more details
chown -R 32767:32767 /var/lib/dokku/data/storage/node-js-app

# mount the directory into your container's /app/storage directory, relative to root
dokku storage:mount app-name /var/lib/dokku/data/storage/node-js-app:/app/storage
```

You can mount one or more directories as desired by following the above pattern.

### Directory Permissions

The host directory should always be owned by the container user and group id. If this is not the case, files may not persist when written to mounted storage.

- Buildpacks via Herokuish: Use `32767:32767` as the file permissions
- For Cloud Native Buildpacks: This will depend on the builder in question, but can be retrieved via the `CNB_USER_ID` and `CNB_GROUP_ID` environment variables on the builder image. Common builders are as follows:
  - heroku/buildpacks: `1000:1000`
  - cloudfoundry/cnb: `2000:2000`
  - packeto: `2000:2000`
- Dockerfile and Docker Image: Use the user and group id which corresponds to the one running the process within the container.

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
