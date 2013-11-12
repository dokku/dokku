#Advanced installation

## Development

If you plan on developing dokku, the easiest way to install from your own repository is cloning
the repository and calling the install script. Example:

    $ git clone https://github.com/yourusername/dokku.git
    $ cd dokku
    $ sudo make install

The `Makefile` allows source URLs to be overridden to include customizations from your own
repositories. The `DOCKER_URL`, `PLUGINHOOK_URL`, `SSHCOMMAND_URL` and `STACK_URL`
environment variables may be set to override the defaults (see the `Makefile` for how these
apply). Example:

    $ sudo SSHCOMMAND_URL=https://raw.github.com/yourusername/sshcommand/master/gitreceive make install

## Bootstrap a server from your own repository

The bootstrap script allows the dokku repository URL to be overridden to bootstrap a host from
your own clone of dokku using the `DOKKU_REPO` environment variable. Example:

    $ wget https://raw.github.com/progrium/dokku/master/bootstrap.sh
    $ chmod +x bootstrap.sh
    $ sudo DOKKU_REPO=https://github.com/yourusername/dokku.git ./bootstrap.sh

## Custom buildstep build

Dokku ships with a pre-built version of version of the [buildstep] component by
default. If you want to build your own version you can specify that with an env
variable.

    $ git clone https://github.com/progrium/dokku.git
    $ cd dokku
    $ sudo BUILD_STACK=true make install

[buildstep]: https://github.com/progrium/buildstep
