# Uninstalling

While we hate to see you go, if you need to uninstall Dokku, the following may help you out:

## Arch Installation

```shell
# purge dokku from your system
yaourt -Rsn dokku
```

## CentOS Installation

```shell
# uninstall dokku
yum remove dokku herokuish
```

## Debian Installation

```shell
# purge dokku from your system
apt-get purge -y --allow dokku herokuish

# remove any dependencies that are no longer necessary
apt-get -y --allow autoremove
```

## Makefile Installation

This is a manual deletion process, and as it is not a recommended installation method, there is currently no automated uninstallation.

All service plugins should be unlinked from applications, stopped, and destroyed.

All applications should be stopped, and all docker containers and images deleted:

```shell
# stop all applications
dokku ps:stopall

# cleanup containers and images
dokku cleanup
```

The following user/group must be deleted:

- user: `dokku`
- group: `dokku`

The following directories must be deleted:

- `~dokku`
- `/var/lib/dokku`
- `/var/log/dokku`
