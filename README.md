# Dokku

Docker powered mini-Heroku. The smallest PaaS implementation you've ever seen.

[![Build Status](https://travis-ci.org/progrium/dokku.png?branch=master)](https://travis-ci.org/progrium/dokku)

## Requirements

Ubuntu 14.04 x64 or 12.04 x64. Ideally have a domain ready to point to your host. It's designed for and is probably best to use a fresh VM. The bootstrapper will install everything it needs.

**Note: Support for 12.04 will be sunsetting in the near future after dokku & 14.04 have been more thoroughly tested.**

## Installing

### Stable

    $ wget -qO- https://raw.github.com/progrium/dokku/v0.2.3/bootstrap.sh | sudo DOKKU_TAG=v0.2.3 bash

**Note**: Users on 12.04 will need to run `apt-get install -y python-software-properties` before bootstrapping stable.

### Development

    $ wget -qO- https://raw.github.com/progrium/dokku/master/bootstrap.sh | sudo bash

This may take around 5 minutes. Certainly better than the several hours it takes to bootstrap Cloud Foundry.

You may also wish to take a look at the [advanced installation](http://progrium.viewdocs.io/dokku/advanced-installation) document for additional installation options.

## Configuring

Set up a domain and a wildcard domain pointing to that host. Make sure `/home/dokku/VHOST` is set to this domain. By default it's set to whatever hostname the host has. This file is only created if the hostname can be resolved by dig (`dig +short $(hostname -f)`). Otherwise you have to create the file manually and set it to your preferred domain. If this file still is not present when you push your app, dokku will publish the app with a port number (i.e. `http://example.com:49154` - note the missing subdomain).

You'll have to add a public key associated with a username by doing something like this from your local machine:

    $ cat ~/.ssh/id_rsa.pub | ssh progriumapp.com "sudo sshcommand acl-add dokku progrium"

That's it!

## Deploy an App

Now you can deploy apps on your Dokku. Let's deploy the [Heroku Node.js sample app](https://github.com/heroku/node-js-sample). All you have to do is add a remote to name the app. It's created on-the-fly.

    $ cd node-js-sample
    $ git remote add progrium dokku@progriumapp.com:node-js-app
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

Dokku only supports deploying from its master branch, so if you'd like to deploy a different local branch use: ```git push progrium <local branch>:master``` 

Right now Buildstep supports buildpacks for Node.js, Ruby, Python, [and more](https://github.com/progrium/buildstep#supported-buildpacks). It's not hard to add more, [go add more](https://github.com/progrium/buildstep#adding-buildpacks)!
Please check the documentation for your particular build pack as you may need to include configuration files (such as a Procfile) in your project root.

## Remote commands

Dokku commands can be run over ssh. Anywhere you would run `dokku <command>`, just run `ssh -t dokku@progriumapp.com <command>`
The `-t` is used to request a pty. It is highly recommended to do so.
To avoid the need to type the `-t` option each time, simply create/modify a section in the `.ssh/config` on the client side, as follows :

    Host progriumapp.com
    RequestTTY yes

## Run a command in the app environment

It's possible to run commands in the environment of the deployed application:

    $ dokku run node-js-app ls -alh
    $ dokku run <app> <cmd>

## Plugins

Dokku itself is built out of plugins. Checkout the wiki for information about
creating your own and a list of existing plugins:

https://github.com/progrium/dokku/wiki/Plugins

## Removing a deployed app

SSH onto the server, then execute:

    $ dokku delete myapp

## Environment variable management

Typically an application will require some environment variables to run properly. Environment variables may contain private data, such as passwords or API keys, so it is not recommend to store them in your application's repository.

The `config` plugin provides the following commands to manage your variables:
```
config <app> - display the config vars for an app  
config:get <app> KEY - display a config value for an app  
config:set <app> KEY1=VALUE1 [KEY2=VALUE2 ...] - set one or more config vars
config:unset <app> KEY1 [KEY2 ...] - unset one or more config vars
```

## TLS/SPDY support

Dokku provides easy TLS/SPDY support out of the box. This can be done app-by-app or for all subdomains at once. Note that whenever TLS support is enabled SPDY is also enabled.

### Per App

To enable TLS connection to to one of your applications, copy or symlink the `.crt`/`.pem` and `.key` files into the application's `/home/dokku/:app/tls` folder (create this folder if it doesn't exist) as `server.crt` and `server.key` respectively.

Redeployment of the application will be needed to apply TLS configuration. Once it is redeployed, the application will be accessible by `https://` (redirection from `http://` is applied as well).

### All Subdomains

To enable TLS connections for all your applications at once you will need a wildcard TLS certificate.

To enable TLS across all apps, copy or symlink the `.crt`/`.pem` and `.key` files into the  `/home/dokku/tls` folder (create this folder if it doesn't exist) as `server.crt` and `server.key` respectively. Then, enable the certificates by editing `/etc/nginx/conf.d/dokku.conf` and uncommenting these two lines (remove the #):

    ssl_certificate /home/dokku/tls/server.crt;
    ssl_certificate_key /home/dokku/tls/server.key;

The nginx configuration will need to be reloaded in order for the updated TLS configuration to be applied. This can be done either via the init system or by re-deploying the application. Once TLS is enabled, the application will be accessible by `https://` (redirection from `http://` is applied as well).

**Note**: TLS will not be enabled unless the application's VHOST matches the certificate's name. (i.e. if you have a cert for *.example.com TLS won't be enabled for something.example.org or example.net)

### HSTS Header

The [HSTS header](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) is an HTTP header that can inform browsers that all requests to a given site should be made via HTTPS. dokku does not, by default, enable this header. It is thus left up to you, the user, to enable it for your site.

Beware that if you enable the header and a subsequent deploy of your application results in an HTTP deploy (for whatever reason), the way the header works means that a browser will not attempt to request the HTTP version of your site if the HTTPS version fails.

## Upgrading

Dokku is in active development. You can update the deployment step and the build step separately.

**Note**: If you are upgrading from a revision prior to [27d4bc8c3c](https://github.com/progrium/dokku/commit/27d4bc8c3c19fe580ef3e65f2f85b85101cd83e4), follow the instructions in [this wiki entry](https://github.com/progrium/dokku/wiki/Migrating-to-Dokku-0.2.0).

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

You can use [Github Issues](https://github.com/progrium/dokku/issues), check [Troubleshooting](https://github.com/progrium/dokku/wiki/Troubleshooting) on the wiki, or join us on [freenode in #dokku](https://webchat.freenode.net/?channels=%23dokku)

## Components

 * [Docker](https://github.com/dotcloud/docker) - Container runtime and manager
 * [Buildstep](https://github.com/progrium/buildstep) - Buildpack builder
 * [pluginhook](https://github.com/progrium/pluginhook) - Shell based plugins and hooks
 * [sshcommand](https://github.com/progrium/sshcommand) - Fixed commands over SSH

Looking to keep codebase as simple and hackable as possible, so try to keep your line count down.

## Things this project won't do

 * **Multi-host.** Not a huge leap, but this isn't the project for it. Have a look at [Flynn](https://flynn.io/).
 * **Multitenancy.** It's ready for it, but again, have a look at [Flynn](https://flynn.io/).
 * **Client app.** Given the constraints, running commands remotely via SSH is fine.

## License

MIT
