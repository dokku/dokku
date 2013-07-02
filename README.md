# Dokku

Docker powered mini-Heroku. The smallest PaaS implementation you've ever seen.

[![Build Status](https://travis-ci.org/progrium/dokku.png?branch=master)](https://travis-ci.org/progrium/dokku)

## Requirements

Assumes Ubuntu 13 right now. Ideally have a domain ready to point to your host. It's designed for and is probably
best to use a fresh VM. The bootstrapper will install everything it needs.

## Installing

    $ wget -qO- https://raw.github.com/progrium/dokku/master/bootstrap.sh | sudo bash

This may take around 5 minutes. Certainly better than the several hours it takes to bootstrap Cloud Foundry.

## Configuring

Set up a domain and a wildcard domain pointing to that host. Make sure `/home/git/DOMAIN` is set to this domain. 
By default it's set to whatever the hostname the host has.

You'll have to add a public key associated with a username as it says at the end of the bootstrapper. You'll do something
like this from your local machine:

    $ cat ~/.ssh/id_rsa.pub | ssh progriumapp.com "sudo gitreceive upload-key progrium"

That's it!

## Deploy an App

Right now Buildstep supports buildpacks for Node.js, Ruby, Python, [and more](https://github.com/progrium/buildstep#supported-buildpacks). It's not hard to add more, [go add more](https://github.com/progrium/buildstep#adding-buildpacks)! Let's deploy
the Heroku Node.js sample app. All you have to do is add a remote to name the app. It's created on-the-fly.

    $ cd node-js-sample
    $ git remote add progrium git@progriumapp.com:node-js-app
    $ git push progrium master
    Counting objects: 296, done.
    Delta compression using up to 4 threads.
    Compressing objects: 100% (254/254), done.
    Writing objects: 100% (296/296), 193.59 KiB, done.
    Total 296 (delta 25), reused 276 (delta 13)
    remote: -----> Building node-js-app ...
    remote:        Node.js app detected
    remote: -----> Resolving engine versions
    
    ... blah blah blah ...
    
    remote: -----> Application deployed:
    remote:        http://node-js-app.progriumapp.com

You're done!

## Advanced installation (for development)

The bootstrap script allows source URLs to be overridden to include customizations from your own 
repositories. The GITRECEIVE_URL and DOKKU_REPO environment variables
may be set to override the defaults (see the bootstrap.sh script for how these apply). Example:

    $ wget j.mp/dokku-bootstrap
    $ chmod +x dokku-bootstrap
    $ sudo DOKKU_REPO=https://github.com/yourusername/dokku.git ./dokku-bootstrap

## Upgrading

Dokku is in active development. You can update the deployment step and the build step separately.
To update the deploy step (this is updated less frequently):

    $ cd ~/dokku
    $ git pull origin master
    $ sudo make install

Nothing needs to be restarted. Changes will take effect on the next push / deployment.

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
 * HTTPS support on default domain
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
