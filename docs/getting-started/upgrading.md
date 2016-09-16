# Upgrading

If your version of Dokku is pre 0.3.0 (check with `dokku version`), we recommend [a fresh install](/dokku/getting-started/installation/) on a new server.

## Migration Guides

Before upgrading, check the migration guides to get comfortable with new features and prepare your deployment to be upgraded.

### 0.5 Migration Guide

- [0.5 Migration Guide](/dokku/appendices/0.5.0-migration-guide/)

### 0.6 Migration Guide

- [0.6 Migration Guide](/dokku/appendices/0.6.0-migration-guide/)

### 0.7 Migration Guide

- [0.7 Migration Guide](/dokku/appendices/0.7.0-migration-guide/)

## Upgrade Instructions

If Dokku was installed via `apt-get install dokku` or `bootstrap.sh` (most common), upgrade with:

```shell
sudo apt-get update
dokku apps
dokku ps:stop <app> # repeat to shut down each running app
sudo apt-get install -qq -y dokku herokuish
dokku ps:rebuildall # rebuilds all applications
```

> If you have any applications deployed via the `tags` or `tar` commands, do not run the `ps:rebuildall` command,
> and instead trigger `ps:rebuild` manually for each `git`-deployed application:
>
> ```
> dokku ps:rebuild APP
> ```
>
> Please see the [images documentation](/dokku/deployment/methods/images/) and [tar documentation](/dokku/deployment/methods/tar/)
> for instructions on rebuilding applications deployed by those plugins.

### Upgrade From Source

If you installed Dokku from source (less common), upgrade with:

```shell
dokku apps
dokku ps:stop <app> # repeat to shut down each running app
cd ~/dokku
git pull --tags origin master

# continue to install from source
sudo DOKKU_BRANCH=master make install

# upgrade to debian package-based installation
sudo make install
dokku ps:rebuildall # rebuilds all applications
```

To upgrade herokuish from source, upgrade with:

```shell
cd /tmp
git clone https://github.com/gliderlabs/herokuish.git
cd herokuish
git pull origin master
IMAGE_NAME=gliderlabs/herokuish BUILD_TAG=latest VERSION=master make -e build-in-docker
```
