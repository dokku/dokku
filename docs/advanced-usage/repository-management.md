# Repository Management

> New as of 0.6.0

```
repo:gc <app>                             # Runs 'git gc --aggressive' against the application's repo
repo:purge-cache <app>                    # Deletes the contents of the build cache stored in the repository
repo:report [<app>] [<flag>]              # Displays a repo report for one or more apps
repo:set <app> <property> (<value>)       # Set or clear a repo property for an app
```

The repository plugin is meant to allow users to perform management commands against a repository.

## Usage

### Clearing Application cache

Building containers with buildpacks currently results in a persistent `cache` directory between deploys. If you need to clear this cache directory for any reason, you may do so by running the following shell command:

```shell
dokku repo:purge-cache node-js-app
```

### Git Garbage Collection

This will run a git gc --aggressive against the applications repo. This is performed on the Dokku host, and not within an application container.

```shell
dokku repo:gc node-js-app
```

```
Counting objects: 396, done.
Delta compression using up to 2 threads.
Compressing objects: 100% (365/365), done.
Writing objects: 100% (396/396), done.
Total 396 (delta 79), reused 315 (delta 0)
```

### Copying files from a repo subdirectory to the repo root

> Warning: This has no effect on image-based deploys, and is disabled by default.

In some cases, you may wish to keep your repository root clear of dokku-specific files, such as `nginx.conf.sigil` or a production-specific `Dockerfile`. Dokku can be configured to set a specific subdirectory as the source of container-specific files via the repo `container-copy-folder` setting:

```shell
dokku repo:set node-js-app container-copy-folder config/dokku
```

This allows you to keep the repository root clean while "committing" the changed file structure to the built docker image.

### Copying files from a repo subdirectory to the app host root

> Warning: This has no effect on image-based deploys, and is disabled by default.

In some cases, you may wish to version files that Dokku would normally have on the Dokku host with your repository, such as `nginx.conf.d` files or a `DOKKU_SCALE` file. Dokku can be configured to set a specific subdirectory as the source of host-specific files via the repo `host-copy-folder` setting:

```shell
dokku repo:set node-js-app host-copy-folder config/dokku-host
```

This allows you to version all files related to deployment with your application, minimizing the need for configuration management on the host. As these files are copied to the "app" directory on the host. special care should be taken to avoid naming files that would conflict with the git directory structure.

### Displaying repo reports about an app

You can get a report about the app's repo settings using the `repo:report` command:

```shell
dokku repo:report
```

```
=====> node-js-app repo information
       Repo container copy folder: ".dokku"
       Repo host copy folder: ".dokku-host"
=====> ruby-sample repo information
       Repo container copy folder:
       Repo host copy folder:
```

You can run the command for a specific app also.

```shell
dokku repo:report node-js-app
```

```
=====> node-js-app repo information
       Repo container copy folder: ".dokku"
       Repo host copy folder: ".dokku-host"
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku repo:report node-js-app --repo-container-copy-folder
```
