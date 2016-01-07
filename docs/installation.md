# Getting Started with Dokku

## What is Dokku?

Dokku is an extensible, open source Platform as a Service that runs on a single server of your choice.

To start using Dokku, you'll need a system that meets the following minimum requirements:

- A fresh installation of [Ubuntu 14.04 x64](http://www.ubuntu.com/download/) with the FQDN set <sup>[1]</sup>
- At least `1GB` of system memory <sup>[2]</sup>

You can *optionally* have a domain name pointed at the host's IP, though this is not necessary.

Dokku is designed for usage on a fresh installation of Ubuntu, and should install all necessary dependencies if installing via the bootstrap method.

### Installing the latest stable version

To install the latest stable version of dokku, you can run the following shell commands:

```shell
# installs dokku via apt-get
wget https://raw.githubusercontent.com/dokku/dokku/v0.4.10/bootstrap.sh
sudo DOKKU_TAG=v0.4.10 bash bootstrap.sh
```

The installation process takes about 5-10 minutes, depending upon internet connection speed.

Once the installation is complete, you can open a browser to setup your SSH key and virtualhost settings. Open your browser of choice and navigate to the host's IP address - or the domain you assigned to that IP previously - and configure dokku via the web admin.

Once you save your settings, the web admin will self-terminate and you should be able to run or deploy to the dokku installation.

### Installing via other methods

For various reasons, certain hosting providers may have other steps that should be preferred to the above. If hosted on any of the following popular hosts, please follow the linked to instructions:

- [Digital Ocean Installation Notes](http://dokku.viewdocs.io/dokku/getting-started/install/digitalocean)
- [Linode Installation Notes](http://dokku.viewdocs.io/dokku/getting-started/install/linode/)
- [Microsoft Azure Installation Notes](http://dokku.viewdocs.io/dokku/getting-started/install/azure/)

As well, you may wish to customize your installation in some other fashion. or experiment with vagrant. The guides below should get you started:

- [Debian Package Installation Notes](http://dokku.viewdocs.io/dokku/getting-started/install/debian)
- [Vagrant Installation Notes](http://dokku.viewdocs.io/dokku/getting-started/install/vagrant)
- [Advanced Install Customization](http://dokku.viewdocs.io/dokku/advanced-installation)

---

- <sup>[1]: To check whether your system has an fqdn set, run `sudo hostname -f`</sup>
- <sup>[2]: If your system has less than 1GB of memory, you can use ([this workaround](http://dokku.viewdocs.io/dokku/advanced-installation)).</sup>
