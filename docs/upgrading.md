# Upgrading

If your version of dokku is pre 0.3.0 (check with `dokku version`), we recommend [a fresh install](/docs/installation) on a new server.

## Migration Guides

Before upgrading, check the migration guides to get comfortable with new features and prepare your deployment to be upgraded.

- [0.5 Migration Guide](/docs/appendices/0.5.0-migration-guide/)

## Upgrade Instructions

If dokku was installed via `apt-get install dokku` or `bootstrap.sh` (most common), upgrade with:

```shell
sudo apt-get update
dokku apps
dokku ps:stop <app> # repeat to shut down each running app
sudo apt-get install -qq -y dokku herokuish
dokku ps:rebuildall # restart all applications
```

### Upgrade From Source

If you installed dokku from source (less common), upgrade with:

```shell
dokku apps
dokku ps:stop <app> # repeat to shut down each running app
cd ~/dokku
git pull --tags origin master

# continue to install from source
sudo DOKKU_BRANCH=master make install

# upgrade to debian package-based installation
sudo make install
dokku ps:rebuildall # restart all applications
```

To upgrade herokuish from source, upgrade with:

```shell
cd /tmp
git clone https://github.com/gliderlabs/herokuish.git
cd herokuish
git pull origin master
IMAGE_NAME=gliderlabs/herokuish BUILD_TAG=latest VERSION=master make -e build-in-docker
```
