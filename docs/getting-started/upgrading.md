# Upgrading

If your version of Dokku is pre 0.3.0 (check with `dokku version`), we recommend [a fresh install](/docs/getting-started/installation.md) on a new server.


## Security Updates

For any security related updates, please follow our [Twitter account](https://twitter.com/dokku). As Dokku does not run any daemons, the security risk introduced by our software is minimal.

Your operating system may occasionally provide security updates. We recommend setting unattended upgrades for your operating system. Here are some helpful links:

- [Arch Linux System Maintenance](https://wiki.archlinux.org/index.php/System_maintenance)
- [Centos Automatic Security Updates](https://serversforhackers.com/c/automatic-security-updates-centos)
- [Debian Unattended Upgrades](https://wiki.debian.org/UnattendedUpgrades)
- [Ubuntu Unattended Upgrades](https://help.ubuntu.com/community/AutomaticSecurityUpdates)

Docker releases updates periodically to their engine. We recommend reading their release notes and upgrading accordingly. Please see the [Docker documentation](https://docs.docker.com/) for more details.


## Migration Guides

Before upgrading, check the migration guides to get comfortable with new features and prepare your deployment to be upgraded.

- [Upgrading to 0.20](/docs/appendices/0.20.0-migration-guide.md)
- [Upgrading to 0.10](/docs/appendices/0.10.0-migration-guide.md)
- [Upgrading to 0.9](/docs/appendices/0.9.0-migration-guide.md)
- [Upgrading to 0.8](/docs/appendices/0.8.0-migration-guide.md)
- [Upgrading to 0.7](/docs/appendices/0.7.0-migration-guide.md)
- [Upgrading to 0.6](/docs/appendices/0.6.0-migration-guide.md)
- [Upgrading to 0.5](/docs/appendices/0.5.0-migration-guide.md)


## Before upgrading

If you'll be updating docker simultaneously, it's recommended that you stop all
applications before upgrading:

```shell
# for 0.11.4 and newer versions, use:
dokku ps:stopall

# for versions between 0.8.1 and 0.11.3, use
dokku --quiet apps:list | xargs -L1 dokku ps:stop

# for versions versions older than 0.8.1, use
dokku --quiet apps | xargs -L1 dokku ps:stop
```

## After upgrading

After upgrading, you should rebuild the applications to take advantage of any
new buildpacks that were released:

```shell
dokku ps:rebuildall
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


## Upgrading using `dokku-update`

We provide a helpful binary called `dokku-update`. This is a recommended package that:

- Can be installed separately, so upgrading Dokku will not affect the running of this package.
- Automates many of the upgrade instructions for you.
- Provides a clean way for us to further enhance the upgrade process in the future.

When installing from source, this is available from `contrib/dokku-update`, and is also available on Debian and RPM-based systems from our package repositories under the name `dokku-update`.


## Upgrading using `apt`

If Dokku was installed in a Debian or Ubuntu system, via `apt install dokku` or `bootstrap.sh`, you can upgrade with `apt`:

```shell
# update your local apt cache
sudo apt update

# update dokku and its dependencies
sudo apt install -qq -y dokku herokuish sshcommand plugn gliderlabs-sigil

# or just upgrade every package:
sudo apt upgrade
```

## Upgrading from source

If you installed Dokku from source (less common), upgrade with:

```shell
cd ~/dokku
git pull --tags origin master

# continue to install from source
sudo DOKKU_BRANCH=master make install

# upgrade to debian package-based installation
sudo make install
```

To upgrade Herokuish from source, upgrade with:

```shell
cd /tmp
git clone https://github.com/gliderlabs/herokuish.git
cd herokuish
git pull origin master
IMAGE_NAME=gliderlabs/herokuish BUILD_TAG=latest VERSION=master make -e build-in-docker
```
