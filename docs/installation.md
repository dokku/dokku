# Installation

## Requirements

- A fresh VM running Ubuntu `14.04 x64`

Ubuntu 14.04 x64 x64. Ideally have a domain ready to point to your host. It's designed for and is probably best to use a fresh VM. The bootstrapper will install everything it needs.

## Installing the latest Stable version

To install the latest stable version of dokku, you can run the following bootstrapper command:

```bash
wget -qO- https://raw.github.com/progrium/dokku/v0.3.1/bootstrap.sh | sudo DOKKU_TAG=v0.3.1 bash
```

## Configuring

Set up a domain and a wildcard domain pointing to that host. Make sure `/home/dokku/VHOST` is set to this domain. By default it's set to whatever hostname the host has. This file is only created if the hostname can be resolved by dig (`dig +short $(hostname -f)`). Otherwise you have to create the file manually and set it to your preferred domain. If this file still is not present when you push your app, dokku will publish the app with a port number (i.e. `http://example.com:49154` - note the missing subdomain).

You'll have to add a public key associated with a username by doing something like this from your local machine:

    $ cat ~/.ssh/id_rsa.pub | ssh progriumapp.com "sudo sshcommand acl-add dokku progrium"

That's it!

## Advanced installation

### Development via bash script

```bash
wget -qO- https://raw.github.com/progrium/dokku/master/bootstrap.sh | sudo bash
```

This may take around 5 minutes. Certainly better than the several hours it takes to bootstrap Cloud Foundry.

You may also wish to take a look at the [advanced installation](http://progrium.viewdocs.io/dokku/advanced-installation) document for additional installation options.

### Development From Source

If you plan on developing dokku, the easiest way to install from your own repository is cloning
the repository and calling the install script. Example:

```bash
git clone https://github.com/yourusername/dokku.git
cd dokku
sudo make install
```

The `Makefile` allows source URLs to be overridden to include customizations from your own
repositories. The `DOCKER_URL`, `PLUGINHOOK_URL`, `SSHCOMMAND_URL` and `STACK_URL`
environment variables may be set to override the defaults (see the `Makefile` for how these
apply). Example:

```bash
sudo SSHCOMMAND_URL=https://raw.github.com/yourusername/sshcommand/master/sshcommand make install
```

### Bootstrap a server from your own repository

The bootstrap script allows the dokku repository URL to be overridden to bootstrap a host from
your own clone of dokku using the `DOKKU_REPO` environment variable. Example:

```bash
wget https://raw.github.com/progrium/dokku/master/bootstrap.sh
chmod +x bootstrap.sh
sudo DOKKU_REPO=https://github.com/yourusername/dokku.git ./bootstrap.sh
```

### Custom buildstep build

Dokku ships with a pre-built version of the [buildstep](https://github.com/progrium/buildstep) component by
default. If you want to build your own version you can specify that with an env
variable.

```bash
git clone https://github.com/progrium/dokku.git
cd dokku
sudo BUILD_STACK=true make install
```

### Install Dokku using Vagrant

- Download and install [VirtualBox](https://www.virtualbox.org/wiki/Downloads)
- Download and install [Vagrant](http://www.vagrantup.com/downloads.html)
- Clone Dokku

    ```
    git clone https://github.com/progrium/dokku.git
    ```

- Setup SSH hosts in your `/etc/hosts`

    ```
    10.0.0.2 dokku.me
    ```

- Setup SSH Config in `~/.ssh/config`

    ```
    Host dokku.me
        Port 2222
    ```

- Create VM
    ```
    # Optional ENV arguments:
    # - `BOX_NAME`
    # - `BOX_URI`
    # - `BOX_MEMORY`
    # - `DOKKU_DOMAIN`
    # - `DOKKU_IP`.
    vagrant up
    ```

- Copy your SSH key via `cat ~/.ssh/id_rsa.pub | pbcopy` and paste it into the dokku-installer at http://dokku.me . Change the `Hostname` field on the Dokku Setup screen to your domain and then check the box that says `Use virtualhost naming`. Then click *Finish Setup* to install your key. You'll be directed to application deployment instructions from here.

You are now ready to deploy an app or install plugins.

For a different, complete, example see https://github.com/RyanBalfanz/dokku-vagrant-example.

### Installing on Linode

#### Using StackScript

Deploy using the following StackScript:
* https://www.linode.com/stackscripts/view/8552

#### Without StackScript

* Build a Ubuntu 13.04 instance

* Follow these instructions: https://www.linode.com/wiki/index.php/PV-GRUB#Ubuntu_12.04_Precise

* If `apt-get update` no longer works:

    * Verify if apt-get is trying to use ipv6 instead of ipv4 (e.g. you read something like "[Connecting to us.archive.ubuntu.com (2001:67c:1562::14)]" and apt-get would not proceed). In that case, follow these instructions: http://unix.stackexchange.com/questions/9940/convince-apt-get-not-to-use-ipv6-method (append "precedence ::ffff:0:0/96  100" to /etc/gai.conf)

    * OR: change `/etc/apt/sources.list` to one mentioned in http://mirrors.ubuntu.com/mirrors.txt

* Run the following commands:

    ```bash
    apt-get update

    apt-get install lxc wget bsdtar linux-image-extra-$(uname -r)

    modprobe aufs
    ```
* After this, you can install dokku the default way:

    ```bash
    wget -qO- https://raw.github.com/progrium/dokku/master/bootstrap.sh | sudo bash
    ```
