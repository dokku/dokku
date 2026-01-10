# Getting Started with Dokku

## What is Dokku?

Dokku is an extensible, open source Platform as a Service that runs on a single server of your choice. Dokku supports building apps on the fly from a `git push` via either Dockerfile or by auto-detecting the language with Buildpacks, and then starts containers based on your built image. Using technologies such as nginx and cron, Web processes are automatically routed to, while background processes and automated cron tasks are also managed by Dokku.

## System Requirements

To start using Dokku, you'll need a system that meets the following minimum requirements:

- Operating Systems: [Ubuntu 22.04/24.04](https://www.ubuntu.com/download) or [Debian 11+ x64](https://www.debian.org/distrib/)
- Supported Architectures: `AMD64` (`x86_64`) and `arm64` (`armv8`)
- Minimum Memory:
    - Docker Scheduler: 1GB of system memory, or [add swap memory](/docs/getting-started/advanced-installation.md#vms-with-less-than-1-gb-of-memory)
    - K3s Scheduler: 2GB of system memory on every node in the cluster

Finally, we recommend attaching at least one domain name to your server. This is not required, but using a domain name will make app access easier. When connecting a domain, either a single domain or a wildcard may be associated to the server's IP.

- Wildcard domain (`*.domain.tld` A Record): will allow access to apps via `$APP.domain.tld`.
- Single domain (`domain.tld` A or CNAME Record): will result in apps being accessed via `domain.tld:$RANDOM_PORT`.

Please see the [dns documentation](/docs/networking/dns.md) and [domains documentation](/docs/configuration/domains.md) for more details.

## Installing the latest stable version

This is the simple method of installing Dokku. For alternative methods of installation, see the [advanced install guide](/docs/getting-started/advanced-installation.md#configuring).

### 1. Install Dokku

To install the latest stable version of Dokku, you can run the following shell commands:

```shell
# for debian systems, installs Dokku via apt-get
wget -NP . https://dokku.com/install/v0.37.5/bootstrap.sh
sudo DOKKU_TAG=v0.37.5 bash bootstrap.sh
```

The installation process takes about 5-10 minutes, depending upon internet connection speed.

### 2. Setup SSH key and Virtualhost Settings

Once the installation is complete, you should configure an ssh key and set your global domain.

```shell
# usually your key is already available under the current user's `~/.ssh/authorized_keys` file
cat ~/.ssh/authorized_keys | sudo dokku ssh-keys:add admin

# you can use any domain you already have access to
# this domain should have an A record or CNAME pointing at your server's IP
dokku domains:set-global dokku.me

# you can also use the ip of your server
dokku domains:set-global 10.0.0.2

# finally, you can use sslip.io to get subdomain support
# as you would with a regular domain name
# this would be done by appending `.sslip.io` to your ip address
dokku domains:set-global 10.0.0.2.sslip.io
```

See the [user management](/docs/deployment/user-management.md#adding-ssh-keys) and [domains documentation](/docs/configuration/domains.md#customizing-hostnames) for more information.

### 3. Deploy your first application

At this point, you should be able to [deploy to the Dokku installation](/docs/deployment/application-deployment.md).
