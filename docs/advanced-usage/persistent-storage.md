# Persistent Storage

> New as of 0.5.0

The preferred method to mount external containers to a Dokku managed container, is to use the Dokku storage plugin.

```
storage:ensure-directory [--chown option] <directory>  # Creates a persistent storage directory in the recommended storage path
storage:list <app>                                     # List bind mounts for app's container(s) (host:container)
storage:mount <app> <host-dir:container-dir>           # Create a new bind mount
storage:report [<app>] [<flag>]                        # Displays a checks report for one or more apps
storage:unmount <app> <host-dir:container-dir>         # Remove an existing bind mount
```

> The storage plugin is compatible with storage mounts created with the docker-options. The storage plugin will only list mounts from the deploy/run phase.

The storage plugin supports the following mount points:

- explicit paths that exist on the host
- docker volumes

## Usage

### Creating storage directories

> New as of 0.25.5

A storage directory can be created with the `storage:ensure-directory` command. This command will create a subdirectory in the recommended `/var/lib/dokku/data/storage` path - created during Dokku installation - and prepare it for use with an app.

```shell
dokku storage:ensure-directory lollipop
```

```
-----> Ensuring /var/lib/dokku/data/storage/lollipop exists
       Setting directory ownership to 32767:32767
       Directory ready for mounting
```

By default, permissions are set for usage with Herokuish buildpacks. These permissions can be changed via the `--chown` option according to the following table:

- `--chown herokuish` (default): Use `32767:32767` as the folder permissions.
  - This is used for apps deployed with Buildpacks via Herokuish.
- `--chown heroku`: Use `1000:1000` as the folder permissions.
  - This is used for apps deployed with Cloud Native Buildpacks using the `heroku/buildpacks` builder.
- `--chown packeto`: Use `2000:2000` as the folder permissions.
  - This is used for apps deployed with Cloud Native Buildpacks using the `cloudfoundry/cnb` or `packeto` builders.
- `--chown false`: Skips the `chown` call.

Users deploying via Dockerfile will want to specify `--chown false` and manually `chown` the created directory if the user and/or group id  of the runnning process in the deployed container do not correspond to any of the above options.

> Warning: Failing to set the correct directory ownership may result in issues in persisting files written to the mounted storage directory.

### Mounting storage into apps

Dokku supports mounting both explicit host paths as well as docker volumes via the `storage:mount` command. This takes two arguments, an app name and a `host-path:container-path` or `docker-volume:container-path` combination.

```shell
# mount the directory into your container's /app/storage directory, relative to the container root (/)
# explicit host paths _must_ exist prior to usage.
dokku storage:mount node-js-app /var/lib/dokku/data/storage/node-js-app:/app/storage

# mount the docker volume into your container's /app/storage directory, relative to the container root (/)
# docker volumes _must_ exist prior to usage.
dokku storage:mount node-js-app some-docker-volume:/app/storage
```

In the first example, Dokku will then mount the shared contents of `/var/lib/dokku/data/storage/node-js-app` to `/app/storage` inside the container.  The mount point is *not* relative to your app's working directory, and is instead relative to the root (`/`) of the container. Mounts are only available for containers created via `run` and by the deploy process, and not during the build process. In addition, the host path is never auto-created by either Dokku or Docker, and should be an explicit path, not one relative to the current working directory.

> If the `/storage` path within the container had pre-existing content, the container files will be over-written. This may be an issue for users that create assets at build time but then mount a directory at the same place during runtime. Files are not merged.

Once persistent storage is mounted, the app requires a restart. See the [process scaling documentation](/docs/processes/process-management.md) for more information.

```shell
dokku ps:restart app-name
```

### Unmounting storage

If an app no longer requires a mounted volume or directory, the `storage:unmount` command can be called. This takes the same arguments as the `storage:mount` command, an app name and a `host-path:container-path` or `docker-volume:container-path` combination.

```shell
# unmount the directory from your container's /app/storage directory, relative to the container root (/)
dokku storage:unmount node-js-app /var/lib/dokku/data/storage/node-js-app:/app/storage

# unmount the docker volume from your container's /app/storage directory, relative to the container root (/)
dokku storage:unmount node-js-app some-docker-volume:/app/storage
```

Once persistent storage is unmounted, the app requires a restart. See the [process scaling documentation](/docs/processes/process-management.md) for more information.

```shell
dokku ps:restart app-name
```

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

### Sharing storage across deploys

Dokku is powered by Docker containers, which recommends in their [best practices](https://docs.docker.com/engine/userguide/eng-image/dockerfile_best-practices/#containers-should-be-ephemeral) that containers be treated as ephemeral. In order to manage persistent storage for web apps, like user uploads or large binary assets like images, a directory outside the container should be mounted.

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

## App User and Persistent Storage file ownership (buildpack apps only)

> New as of 0.7.1

By default, Dokku will execute your buildpack app processes as the `herokuishuser` user. You may override this by setting the `DOKKU_APP_USER` config variable.

> NOTE: this user must exist in your herokuish image.

Additionally, the default `docker-local` scheduler that comes with Dokku will ensure your storage mounts are owned by either `herokuishuser` or the overridden value you have set in `DOKKU_APP_USER`. See the [docker-local scheduler documentation](/docs/deployment/schedulers/docker-local.md#disabling-chown-of-persistent-storage) docs for more information.
