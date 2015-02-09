# Installation

## Requirements

- A fresh VM running Ubuntu `14.04 x64`.
- (optional) A domain pointed at your VM before installation.

## Installing the latest Stable version

To install the latest stable release, you can run the following command as a user that has access to `sudo`:

```shell
wget -qO- https://raw.github.com/progrium/dokku/v0.3.14/bootstrap.sh | sudo DOKKU_TAG=v0.3.14 bash
```

- [Linode Installation](http://progrium.viewdocs.io/dokku/install/linode)
- [Vagrant Installation](http://progrium.viewdocs.io/dokku/install/vagrant)
- [Advanced Install Customization](http://progrium.viewdocs.io/dokku/install/advanced)

## Configuring

Set up a domain and a wildcard domain pointing to that host. Make sure `/home/dokku/VHOST` is set to this domain. By default it's set to whatever hostname the host has. This file is only created if the hostname can be resolved by dig (`dig +short $(hostname -f)`). Otherwise you have to create the file manually and set it to your preferred domain. If this file still is not present when you push your app, dokku will publish the app with a port number (i.e. `http://example.com:49154` - note the missing subdomain).

You'll have to add a public key associated with a username by doing something like this from your local machine:

    $ cat ~/.ssh/id_rsa.pub | ssh dokku.me "sudo sshcommand acl-add dokku $USER"

If you are using the vagrant installation, you can use the following command to add your public key to dokku:

    $ cat ~/.ssh/id_rsa.pub | make vagrant-acl-add

That's it!
