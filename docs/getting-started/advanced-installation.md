# Advanced installation

You can always install Dokku straight from the latest - potentially unstable - `master` branch via the following Bash command:

```shell
# using a branch results in installing from source
wget https://raw.githubusercontent.com/dokku/dokku/master/bootstrap.sh;
sudo DOKKU_BRANCH=master bash bootstrap.sh
```

## Development

If you plan on developing Dokku, the easiest way to install from your own repository is cloning the repository and calling the install script. Example:

```shell
git clone https://github.com/yourusername/dokku.git
cd dokku
sudo make install
```

The `Makefile` allows source URLs to be overridden to include customizations from your own repositories. The `DOCKER_URL`, `PLUGN_URL`, `SSHCOMMAND_URL` and `STACK_URL` environment variables may be set to override the defaults (see the `Makefile` for how these apply). Example:

```shell
sudo SSHCOMMAND_URL=https://raw.githubusercontent.com/yourusername/sshcommand/master/sshcommand make install
```

## Bootstrap a server from your own repository

The bootstrap script allows the Dokku repository URL to be overridden to bootstrap a host from your own clone of Dokku using the `DOKKU_REPO` environment variable. Example:

```shell
wget https://raw.githubusercontent.com/dokku/dokku/master/bootstrap.sh;
chmod +x bootstrap.sh
sudo DOKKU_REPO=https://github.com/yourusername/dokku.git DOKKU_BRANCH=master ./bootstrap.sh
```

## Custom Herokuish build

Dokku ships with a pre-built version of version of the [Herokuish](https://github.com/gliderlabs/herokuish) component by default. If you want to build your own version you can specify that with an environment variable.

```shell
git clone https://github.com/dokku/dokku.git
cd dokku
sudo BUILD_STACK=true STACK_URL=https://github.com/gliderlabs/herokuish.git make install
```

## Skipping Herokuish installation

The Herokuish package is recommended but not required if not using Heroku buildpacks for deployment. Debian-based OS users can run the bootstrap installer via `sudo DOKKU_NO_INSTALL_RECOMMENDS=true bash bootstrap.sh` to skip the dependency. Please note that this will _also_ skip installation of other recommended dependencies.

## Configuring an unattended installation

Once Dokku is installed, if you are not using the web-installer, you'll want to configure the virtualhost setup as well as the push user. If you do not, your installation will be considered incomplete and you will not be able to deploy applications.

For Debian, unattended installation is described [Debian installation guide](/docs/getting-started/install/debian.md).

*You should also stop and disable the `dokku-installer` service to remove public access to adding SSH keys.*

Set up a domain using your preferred vendor and a wildcard domain pointing to the host running Dokku. You can manage this global domain using the [domains plugin](/docs/configuration/domains.md).

Follow the [user management documentation](/docs/deployment/user-management.md) in order to add SSH keys for users to Dokku, or to give other Unix accounts access to Dokku.

## VMs with less than 1 GB of memory

Having less than 1 GB of system memory available for Dokku and its containers may result in unexpected errors, such as `! [remote rejected] master -> master (pre-receive hook declined)` during installation of NPM dependencies (https://github.com/npm/npm/issues/3867).

To work around this issue, it might suffice to augment the Linux swap file size to a maximum of twice the physical memory size.

To resize the swap file of a 512 MB machine to 1 GB, follow these steps while in SSH within your machine:

```shell
cd /var
touch swap.img
chmod 600 swap.img

dd if=/dev/zero of=/var/swap.img bs=1024k count=1000
mkswap /var/swap.img
swapon /var/swap.img
free

echo "/var/swap.img    none    swap    sw    0    0" >> /etc/fstab
```

[Reference](https://www.digitalocean.com/community/tutorials/how-to-configure-virtual-memory-swap-file-on-a-vps)
