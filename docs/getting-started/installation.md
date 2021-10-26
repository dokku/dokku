# Getting Started with Dokku

## What is Dokku?

Dokku is an extensible, open source Platform as a Service that runs on a single server of your choice.

To start using Dokku, you'll need a system that meets the following minimum requirements:

- A fresh installation of [Ubuntu 18.04/20.04 x64](https://www.ubuntu.com/download), [Debian 9+ x64](https://www.debian.org/distrib/) or [CentOS 7 x64](https://www.centos.org/download/) *(experimental)* with the FQDN set <sup>[1]</sup>
- At least 1 GB of system memory <sup>[2]</sup>

Dokku is designed for usage on a fresh VM installation, and should install all necessary dependencies if installing via the bootstrap method.

### Installing the latest stable version

#### 1. Install Dokku

To install the latest stable version of Dokku, you can run the following shell commands:

```shell
# for debian systems, installs Dokku via apt-get
wget https://raw.githubusercontent.com/dokku/dokku/v0.26.0/bootstrap.sh;
sudo DOKKU_TAG=v0.26.0 bash bootstrap.sh
```

The installation process takes about 5-10 minutes, depending upon internet connection speed.

If you're using Debian 9+ or Ubuntu 18.04, make sure your package manager is configured to install a sufficiently recent version of nginx<sup>[3]</sup>, otherwise, the installation may fail due to `unmet dependencies` relating nginx.

#### 2. Optionally connect a domain to your server

Dokku optionally supports one or more domain names. If you do not own a domain name, you may either purchase one or skip this step.

When connecting a domain, either a single domain or a wildcard may be associated to the server's IP. A wildcard domain will allow access to apps via `$APP.domain.tld`, whereas associating only a single domain name will result in apps being access via `domain.tld:$RANDOM_PORT`. Please see the [dns documentation](/docs/networking/dns.md) and [domains documentation](/docs/configuration/domains.md) for more details.

#### 3. Setup SSH key and Virtualhost Settings

Once the installation is complete, you should configure an ssh key and set your global domain:from-archive

```shell
# usually your key is already available under the current user's `~/.ssh/authorized_keys` file
cat ~/.ssh/authorized_keys | dokku ssh-keys:add admin

# you can use any domain you already have access to
dokku domains:set-global dokku.me
```

See the [user management](/docs/deployment/user-management.md#adding-ssh-keys) and [domains documentation](/docs/configuration/domains.md#customizing-hostnames) for more information.

Alternatively, instructions for an unattended installation are available in the [advanced install guide](/docs/getting-started/advanced-installation.md#configuring). 

#### 4. Deploy your first application

At this point, you should be able to run or [deploy to the Dokku installation](/docs/deployment/application-deployment.md).

### Installing via other methods

For various reasons, certain hosting providers may have other steps that should be preferred to the above. If hosted on any of the following popular hosts, please follow the linked to instructions:

- [DigitalOcean Installation Notes](/docs/getting-started/install/digitalocean.md)
- [DreamHost Cloud Installation Notes](/docs/getting-started/install/dreamhost.md)
- [Microsoft Azure Installation Notes](/docs/getting-started/install/azure.md)

As well, you may wish to customize your installation in some other fashion. or experiment with Vagrant. The guides below should get you started:

- [Debian Package Installation Notes](/docs/getting-started/install/debian.md)
- [Docker-based Installation Notes](/docs/getting-started/install/docker.md)
- [RPM Package Installation Notes](/docs/getting-started/install/rpm.md)
- [Vagrant Installation Notes](/docs/getting-started/install/vagrant.md)
- [Advanced Install Customization](/docs/getting-started/advanced-installation.md)
- [Automated deployment via ansible](https://github.com/dokku/ansible-dokku)

---

- <sup>[1]: To check whether your system has an fqdn set, run `sudo hostname -f`</sup>
- <sup>[2]: If your system has less than 1GB of memory, you can use [this workaround](/docs/getting-started/advanced-installation.md#vms-with-less-than-1gb-of-memory).</sup>
- <sup>[3]: nginx >= 1.8.0 can be installed via the [nginx repositories](https://www.nginx.com/resources/admin-guide/installing-nginx-open-source/), or by adding [this PPA](https://launchpad.net/~nginx/+archive/ubuntu/stable) if you're using Ubuntu. nginx >= 1.11.5 is necessary for HTTP/2 support</sup>
