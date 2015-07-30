# Installation

## Requirements

- A fresh VM running Ubuntu `14.04 x64`

Ubuntu 14.04 x64 x64. Ideally have a domain ready to point to your host. It's designed for and is probably best to use a fresh VM. The bootstrapper will install everything it needs.

## Installing the latest Stable version

To install the latest stable version of dokku, you can run the following bootstrapper command:

```shell
# installs dokku via apt-get
wget https://raw.github.com/progrium/dokku/v0.3.22/bootstrap.sh
sudo DOKKU_TAG=v0.3.22 bash bootstrap.sh

# By default, this will do cli-based setup, though you may *also*
# go to your server's IP and follow the web installer
```

For various reasons, certain hosting providers may have other steps that should be preferred to the above. If hosted on any of the following popular hosts, please follow the linked to instructions:

- [Digital Ocean Installation Notes](http://progrium.viewdocs.io/dokku/getting-started/install/digitalocean)
- [Linode Installation Notes](http://progrium.viewdocs.io/dokku/getting-started/install/linode)

As well, you may wish to customize your installation in some other fashion. or experiment with vagrant. The guides below should get you started:

- [Debian Package Installation Notes](http://progrium.viewdocs.io/dokku/getting-started/install/debian)
- [Vagrant Installation Notes](http://progrium.viewdocs.io/dokku/getting-started/install/vagrant)
- [Advanced Install Customization](http://progrium.viewdocs.io/dokku/advanced-installation)

### VMs with only up to 512mb memory
Having only 512mb of system-memory available for dokku and its containers, for example Digital Ocean's smallest machines, might result in unexpected errors, such as **! [remote rejected] master -> master (pre-receive hook declined)** during installation of NPM dependencies (https://github.com/npm/npm/issues/3867).

To work around this issue, it might suffice to augment the linux swap file size to a maximum of twice the physical memory size.

To resize the swap file of a 512MB machine to 1GB, follow these steps while in SSH within your machine:
```
root@my.droplet:/# cd /var
root@my.droplet:/# touch swap.img
root@my.droplet:/# chmod 600 swap.img

root@my.droplet:/# dd if=/dev/zero of=/var/swap.img bs=1024k count=1000
root@my.droplet:/# mkswap /var/swap.img
root@my.droplet:/# swapon /var/swap.img
root@my.droplet:/# free

root@my.droplet:/# echo "/var/swap.img    none    swap    sw    0    0" >> /etc/fstab
```
Reference: https://www.digitalocean.com/community/tutorials/how-to-configure-virtual-memory-swap-file-on-a-vps
