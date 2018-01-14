# Upgrading

If your version of Dokku is pre 0.3.0 (check with `dokku version`), we recommend [a fresh install](/docs/getting-started/installation.md) on a new server.

## Security Updates

For any security related updates, please follow our [twitter account](https://twitter.com/savant). As Dokku does not run any daemons, the security risk introduced by our software is minimal.

Your operating system may occasionally provide security updates. We recommend setting unattended upgrades for your operating system. Here are some helpful links:

- [Arch Linux System Maintenance](https://wiki.archlinux.org/index.php/System_maintenance)
- [Centos Automatic Security Updates](https://serversforhackers.com/c/automatic-security-updates-centos)
- [Debian Unattended Upgrades](https://wiki.debian.org/UnattendedUpgrades)
- [Ubuntu Unattended Upgrades](https://help.ubuntu.com/community/AutomaticSecurityUpdates)

Finally, Docker releases updates periodically to their engine. We recommend reading their release notes and upgrading accordingly. Please see the [Docker documentation](https://docs.docker.com/) for more details.

## Migration Guides

Before upgrading, check the migration guides to get comfortable with new features and prepare your deployment to be upgraded.

### 0.5 Migration Guide

- [0.5 Migration Guide](/docs/appendices/0.5.0-migration-guide.md)

### 0.6 Migration Guide

- [0.6 Migration Guide](/docs/appendices/0.6.0-migration-guide.md)

### 0.7 Migration Guide

- [0.7 Migration Guide](/docs/appendices/0.7.0-migration-guide.md)

### 0.8 Migration Guide

- [0.8 Migration Guide](/docs/appendices/0.8.0-migration-guide.md)

### 0.9 Migration Guide

- [0.9 Migration Guide](/docs/appendices/0.9.0-migration-guide.md)

### 0.10 Migration Guide

- [0.10 Migration Guide](/docs/appendices/0.10.0-migration-guide.md)

## Upgrade Instructions

If Dokku was installed via `apt-get install dokku` or `bootstrap.sh` (most common), upgrade with:

```shell
# update your local apt cache
sudo apt-get update

# stop each running app
# for 0.11.4 and newer versions, use
dokku ps:stopall
# for versions between 0.8.1 and 0.11.3, use
dokku --quiet apps:list | xargs -L1 dokku ps:stop
# for versions versions older than 0.8.1, use
dokku --quiet apps | xargs -L1 dokku ps:stop

# update dokku and it's dependencies
sudo apt-get install -qq -y dokku herokuish sshcommand plugn

# rebuild all of your applications
dokku ps:rebuildall # rebuilds all applications
```

> If you have any applications deployed via the `tags` or `tar` commands, do not run the `ps:rebuildall` command,
> and instead trigger `ps:rebuild` manually for each `git`-deployed application:
>
> ```
> dokku ps:rebuild APP
> ```
>
> Please see the [images documentation](/docs/deployment/methods/images.md) and [tar documentation](/docs/deployment/methods/tar.md)
> for instructions on rebuilding applications deployed by those plugins.

### Upgrade From Source

If you installed Dokku from source (less common), upgrade with:

```shell
dokku --quiet apps | xargs -L1 dokku ps:stop # stops each running app
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
