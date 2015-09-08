# Advanced installation

You can always install dokku straight from the latest - potentially unstable - master release via the following bash command:

```shell
wget -qO- https://raw.github.com/progrium/dokku/master/bootstrap.sh | sudo DOKKU_BRANCH=master bash
```

## Development

If you plan on developing dokku, the easiest way to install from your own repository is cloning the repository and calling the install script. Example:

```shell
git clone https://github.com/yourusername/dokku.git
cd dokku
sudo make install
```

The `Makefile` allows source URLs to be overridden to include customizations from your own repositories. The `DOCKER_URL`, `PLUGN_URL`, `SSHCOMMAND_URL` and `STACK_URL` environment variables may be set to override the defaults (see the `Makefile` for how these apply). Example:

```shell
sudo SSHCOMMAND_URL=https://raw.github.com/yourusername/sshcommand/master/sshcommand make install
```

## Bootstrap a server from your own repository

The bootstrap script allows the dokku repository URL to be overridden to bootstrap a host from your own clone of dokku using the `DOKKU_REPO` environment variable. Example:

```shell
wget https://raw.github.com/progrium/dokku/master/bootstrap.sh
chmod +x bootstrap.sh
sudo DOKKU_REPO=https://github.com/yourusername/dokku.git DOKKU_BRANCH=master ./bootstrap.sh
```

## Custom herokuish build

Dokku ships with a pre-built version of version of the [herokuish](https://github.com/gliderlabs/herokuish) component by default. If you want to build your own version you can specify that with an env variable.

```shell
git clone https://github.com/progrium/dokku.git
cd dokku
sudo BUILD_STACK=true STACK_URL=https://github.com/gliderlabs/herokuish.git make install
```

## Configuring

Once dokku is installed, if you are not using the web-installer, you'll want to configure a the virtualhost setup as well as the push user. If you do not, your installation will be considered incomplete and you will not be able to deploy applications.

Set up a domain and a wildcard domain pointing to that host. Make sure `/home/dokku/VHOST` is set to this domain. By default it's set to whatever hostname the host has. This file is only created if the hostname can be resolved by dig (`dig +short $(hostname -f)`). Otherwise you have to create the file manually and set it to your preferred domain. If this file still is not present when you push your app, dokku will publish the app with a port number (i.e. `http://example.com:49154` - note the missing subdomain).

You'll have to add a public key associated with a username by doing something like this from your local machine:

    $ cat ~/.ssh/id_rsa.pub | ssh dokku.me "sudo sshcommand acl-add dokku $USER"

If you are using the vagrant installation, you can use the following command to add your public key to dokku:

    $ cat ~/.ssh/id_rsa.pub | make vagrant-acl-add

That's it!

## VMs with less than 1GB of memory

Having less than 1GB of system memory available for dokku and its containers, for example Digital Ocean's small 512MB machines, might result in unexpected errors, such as **! [remote rejected] master -> master (pre-receive hook declined)** during installation of NPM dependencies (https://github.com/npm/npm/issues/3867).

To work around this issue, it might suffice to augment the linux swap file size to a maximum of twice the physical memory size.

To resize the swap file of a 512MB machine to 1GB, follow these steps while in SSH within your machine:

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
Reference: https://www.digitalocean.com/community/tutorials/how-to-configure-virtual-memory-swap-file-on-a-vps
