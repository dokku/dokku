# Upgrading

If your version of Dokku is pre 0.3.0 (check with `dokku version`), we recommend [a fresh install](/docs/getting-started/installation.md) on a new server.

## Security Updates

For any security related updates, please follow our [Twitter account](https://twitter.com/dokku). As Dokku does not run any daemons, the security risk introduced by our software is minimal.

Your operating system may occasionally provide security updates. We recommend setting unattended upgrades for your operating system. Here are some helpful links:

- [Arch Linux System Maintenance](https://wiki.archlinux.org/index.php/System_maintenance)
- [CentOS Automatic Security Updates](https://serversforhackers.com/c/automatic-security-updates-centos)
- [Debian Unattended Upgrades](https://wiki.debian.org/UnattendedUpgrades)
- [Ubuntu Unattended Upgrades](https://help.ubuntu.com/community/AutomaticSecurityUpdates)

Docker releases updates periodically to their engine. We recommend reading their release notes and upgrading accordingly. Please see the [Docker documentation](https://docs.docker.com/) for more details.

## Migration Guides

Before upgrading, check the migration guides to get comfortable with new features and prepare your deployment to be upgraded.

- [Upgrading to 0.27](/docs/appendices/0.27.0-migration-guide.md)
- [Upgrading to 0.26](/docs/appendices/0.26.0-migration-guide.md)
- [Upgrading to 0.25](/docs/appendices/0.25.0-migration-guide.md)
- [Upgrading to 0.24](/docs/appendices/0.24.0-migration-guide.md)
- [Upgrading to 0.23](/docs/appendices/0.23.0-migration-guide.md)
- [Upgrading to 0.22](/docs/appendices/0.22.0-migration-guide.md)
- [Upgrading to 0.21](/docs/appendices/0.21.0-migration-guide.md)
- [Upgrading to 0.20](/docs/appendices/0.20.0-migration-guide.md)
- [Upgrading to 0.10](/docs/appendices/0.10.0-migration-guide.md)
- [Upgrading to 0.9](/docs/appendices/0.9.0-migration-guide.md)
- [Upgrading to 0.8](/docs/appendices/0.8.0-migration-guide.md)
- [Upgrading to 0.7](/docs/appendices/0.7.0-migration-guide.md)
- [Upgrading to 0.6](/docs/appendices/0.6.0-migration-guide.md)
- [Upgrading to 0.5](/docs/appendices/0.5.0-migration-guide.md)

## Before upgrading

If you'll be updating docker or the herokuish package simultaneously, it's recommended
that you stop all applications before upgrading and rebuild afterwards. This is not
required if the upgrade only impacts the `dokku` package.

Why do we recommend stopping all apps?

- `docker`: Containers may be randomly reset during the upgrade process, resulting in
  requests being sent to the wrong containers. Acknowledging and scheduling downtime
  thus becomes much more important.
- `herokuish`: While not required, it may be useful to take advantage of the latest
  base image. Herokuish changes do not cause issues unless the base OS changes, which
  may happen in minor or major releases.

```shell
# for 0.22.0 and newer versions, use
dokku ps:stop --all

# for versions between 0.11.4 and 0.21.4, use
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
dokku ps:rebuild --all
```

## Upgrading using `dokku-update`

We provide a helpful binary called `dokku-update`. This is a recommended package that:

- Can be installed separately, so upgrading Dokku will not affect the running of this package.
- Automates many of the upgrade instructions for you.
- Provides a clean way for us to further enhance the upgrade process in the future.

This binary is available on Debian and RPM-based systems from our package repositories under the name `dokku-update`. When installing from source,
this is available from a separate Github repository at [dokku/dokku-update](https://github.com/dokku/dokku-update).

## Upgrading using `apt`

If Dokku was installed in a Debian or Ubuntu system, via `apt-get install dokku` or `bootstrap.sh`, you can upgrade with `apt-get`:

```shell
# update your local apt cache
sudo apt-get update -qq

# update dokku and its dependencies
sudo apt-get -qq -y --no-install-recommends install dokku herokuish sshcommand plugn gliderlabs-sigil dokku-update dokku-event-listener

# or just upgrade every package:
sudo apt-get upgrade
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
CIRCLECI=true IMAGE_NAME=gliderlabs/herokuish BUILD_TAG=latest make build/docker
```
