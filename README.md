# Dokku

Docker powered mini-Heroku. The smallest PaaS implementation you've ever seen.

[![Build Status](https://travis-ci.org/progrium/dokku.png?branch=master)](https://travis-ci.org/progrium/dokku)

## Requirements

Assumes Ubuntu 13 x64 right now. Ideally have a domain ready to point to your host. It's designed for and is probably
best to use a fresh VM. The bootstrapper will install everything it needs.

## Installing

    $ wget -qO- https://raw.github.com/progrium/dokku/master/bootstrap.sh | sudo bash

This may take around 5 minutes. Certainly better than the several hours it takes to bootstrap Cloud Foundry.

## Configuring

Set up a domain and a wildcard domain pointing to that host. Make sure `/home/git/VHOST` is set to this domain. By default it's set to whatever the hostname the host has. This file only created if the hostname can be resolved by dig (`dig +short $HOSTNAME`). Otherwise you have to create the file manually and set it to your prefered domain. If this file still not present when you push your app, dokku will publish the app with a port number (i.e. `http://example.com:49154` - note the missing subdomain).

You'll have to add a public key associated with a username as it says at the end of the bootstrapper. You'll do something
like this from your local machine:

    $ cat ~/.ssh/id_rsa.pub | ssh progriumapp.com "sudo gitreceive upload-key progrium"

That's it!

## Deploy an App

Right now Buildstep supports buildpacks for Node.js, Ruby, Python, [and more](https://github.com/progrium/buildstep#supported-buildpacks). It's not hard to add more, [go add more](https://github.com/progrium/buildstep#adding-buildpacks)!
Please check the documentation for your particular build pack as you may need to include configuration files (such as a Procfile) in your project root.
Let's deploy the [Heroku Node.js sample app](https://github.com/heroku/node-js-sample). All you have to do is add a remote to name the app. It's created on-the-fly.

    $ cd node-js-sample
    $ git remote add progrium git@progriumapp.com:node-js-app
    $ git push progrium master
    Counting objects: 296, done.
    Delta compression using up to 4 threads.
    Compressing objects: 100% (254/254), done.
    Writing objects: 100% (296/296), 193.59 KiB, done.
    Total 296 (delta 25), reused 276 (delta 13)
    -----> Building node-js-app ...
           Node.js app detected
    -----> Resolving engine versions

    ... blah blah blah ...

    -----> Application deployed:
           http://node-js-app.progriumapp.com

You're done!

## Run a command in the app environment

It's possible to run commands in the environment of the deployed application:

    $ dokku run node-js-app ls -alh
    $ dokku run <app> <cmd>

## Plugins

Dokku itself is built out of plugins. Checkout the wiki for information about
creating your own and a list of existing plugins:

https://github.com/progrium/dokku/wiki/Plugins

## Removing a deployed app

Currently this is a manual process.

To remove an app, ssh to the server, then run:

    $ sudo docker ps
    # Then from the list, take repository name of your app and run:
    $ sudo docker stop app/node-js-sample
    # To find the ids of images to delete, run:
    $ sudo docker images
    # Then from that list, take the IDs corresponding to your app, and
    # those corresponding to no tag at all, and for each run:
    $ sudo docker rmi 123456789

## Environment setup

Typically application requires some environment variables to be set up for proper run. Environment variables might contain some private data, like passwords and API keys, so it's not recommend to store them as part of source code.

To setup environment for your application, create file `/home/git/APP_NAME/ENV`. This file is a script that would expose all required environment variables, like:

    export NODE_ENV=production
    export MONGODB_PASSWORD=password

Next time the application is deployed, those variables would be exposed by `start` script.

## SSL support

Dokku provides easy SSL support from the box. To enable SSL connection to your application, copy `.crt` and `.key` file into `/home/git/:app/ssl` folder (notice, file names should be `server.crt` and `server.key`, respectively). Redeployment of application will be needed to apply SSL configuration. Once it redeployed, application will be accessible by `https://` (redirection from `http://` is applied as well).

## Advanced installation (for development)

If you plan on developing dokku, the easiest way to install from your own repository is cloning
the repository and calling the install script. Example:

    $ git clone https://github.com/yourusername/dokku.git
    $ cd dokku
    $ sudo make all

The `Makefile` allows source URLs to be overridden to include customizations from your own
repositories. The DOCKER_URL, GITRECEIVE_URL, PLUGINHOOK_URL, SSHCOMMAND_URL and STACK_URL
environment variables may be set to override the defaults (see the `Makefile` for how these
apply). Example:

    $ sudo GITRECEIVE_URL=https://raw.github.com/yourusername/gitreceive/master/gitreceive make all

## Advanced installation (bootstrap a server from your own repository)

The bootstrap script allows the dokku repository URL to be overridden to bootstrap a host from
your own clone of dokku using the DOKKU_REPO environment variable. Example:

    $ wget https://raw.github.com/progrium/dokku/master/bootstrap.sh
    $ chmod +x bootstrap.sh
    $ sudo DOKKU_REPO=https://github.com/yourusername/dokku.git ./bootstrap.sh

## Advanced installation (custom buildstep build)

Dokku ships with a pre-built version of version of the [buildstep] component by
default. If you want to build your own version you can specify that with an env
variable.

    $ git clone https://github.com/progrium/dokku.git
    $ cd dokku
    $ sudo BUILD_STACK=true make all

[buildstep]: https://github.com/progrium/buildstep

## Upgrading

Dokku is in active development. You can update the deployment step and the build step separately.

To update the deploy step (this is updated less frequently):

    $ cd ~/dokku
    $ git pull origin master
    $ sudo make install

Nothing needs to be restarted. Changes will take effect on the next push / deployment.

To update the build step:

    $ git clone https://github.com/progrium/buildstep.git
    $ cd buildstep
    $ git pull origin master
    $ sudo make build

This will build a fresh Ubuntu Quantal image, install a number of packages, and
eventually replace the Docker image for buildstep.

## Support

You can use [Github Issues](https://github.com/progrium/dokku/issues), check [Troubleshooting](https://github.com/progrium/dokku/wiki/Troubleshooting) on the wiki, or join us on Freenode in #dokku

## Components

 * [Docker](https://github.com/dotcloud/docker) - Container runtime and manager
 * [Buildstep](https://github.com/progrium/buildstep) - Buildpack builder
 * [gitreceive](https://github.com/progrium/gitreceive) - Git push interface
 * [pluginhook](https://github.com/progrium/pluginhook) - Shell based plugins and hooks
 * [sshcommand](https://github.com/progrium/sshcommand) - Fixed commands over SSH

## Ideas for Improvements

 * Custom domain support for apps
 * Support more buildpacks (see Buildstep)
 * Use dokku as the system user instead of git
 * Heroku-ish commands to be run via SSH (like [Dokuen](https://github.com/peterkeen/dokuen#available-app-sub-commands))

Looking to keep codebase as simple and hackable as possible, so try to keep your line count down.

## Things this project won't do

 * **Multi-host.** Not a huge leap, but this isn't the project for it. Maybe as Super Dokku.
 * **Multitenancy.** It's ready for it, but again, probably for Super Dokku.
 * **Client app.** Given the constraints, running commands remotely via SSH is fine.

## License

MIT
