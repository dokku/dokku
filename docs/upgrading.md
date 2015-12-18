# Upgrading

This document covers upgrades for the 0.3.0 series and up. If upgrading from older versions, we recommend [a fresh install](http://dokku.viewdocs.io/dokku/installation) on a new server.

> As of 0.3.18, dokku is installed by default via a debian package. Source-based installations are still available, though not recommended.

## Dokku

If dokku was installed via a debian package, you can upgrade dokku via the following command:

```shell
sudo apt-get install dokku
```

For unattended upgrades, you may run the following command:

```shell
sudo apt-get install -qq -y dokku
```

If you have installed dokku from source, you may run the following commands to upgrade:

```shell
cd ~/dokku
git pull --tags origin master

# continue to install from source
sudo DOKKU_BRANCH=master make install

# upgrade to debian package-based installation
sudo make install
```

All changes will take effect upon next application deployment. To trigger a rebuild of every application, simply run the following command:

```shell
dokku ps:rebuildall
```

## Herokuish image

If dokku was installed via a debian package, you can upgrade herokuish via the following command:

```shell
sudo apt-get install herokuish
```

For unattended upgrades, you may run the following command:

```shell
sudo apt-get install -qq -y herokuish
```

In some cases, it may be desirable to run a specific version of herokuish. To install/upgrade herokuish from source, run the following commands:

```shell
cd /tmp
git clone https://github.com/gliderlabs/herokuish.git
cd herokuish
git pull origin master
IMAGE_NAME=gliderlabs/herokuish BUILD_TAG=latest VERSION=master make -e build-in-docker
```
