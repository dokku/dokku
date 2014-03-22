# Davis

A Dokku powered mini-Travis(-CI). A tiny, under-featured CI purpose built for an engineering CTF-alike event.

Forked from Dokku and suffering from a muddled identity. Lots of commands rely on Dokku commands and jargon, but with some unstated redefinitions to make things apply in a CI environment.

## Requirements

Assumes Ubuntu 13 or 12.04 x64 right now. Ideally have a domain ready to point to your host. It's designed for and is probably best to use a fresh VM. The bootstrapper will install everything it needs.

**Note: There are known issues with docker and Ubuntu 13.10 ([1](https://github.com/dotcloud/docker/issues/1300), [2](https://github.com/dotcloud/docker/issues/1906)) - use of 12.04 is recommended until these issues are resolved.**

## Installation

    $ git clone
    $ git checkout davis
    $ sudo make install

## Configuring

You'll have to add a public key associated with a username by doing something like this from your local machine:

    $ cat ~/.ssh/id_rsa.pub | ssh progriumapp.com "sudo sshcommand acl-add dokku progrium"

Davis ships with a pre-built version of version of the buildstep component by default. If you want to build your own version you can specify that with an env variable. Specify `BUILD_STACK=true` when you run `sudo make install` to build it fresh.

## Upgrading

Davis is in active development. You can update the deployment step and the build step separately.

To update the deploy step (this is updated less frequently):

    $ cd ~/dokku
    $ git pull origin davis
    $ sudo make install

Nothing needs to be restarted. Changes will take effect on the next push.

To update the build step:

    $ git clone https://github.com/progrium/buildstep.git
    $ cd buildstep
    $ git pull origin master
    $ sudo make build

This will build a fresh Ubuntu Quantal image, install a number of packages, and eventually replace the Docker image for buildstep.

## Run some tests

Now you can run some tests on Davis. Simply add a remote to the name of your app and the job will be created on the fly.

    $ cd my-project
    $ git remote add ci-server dokku@ci-server.biz:my-project
    $ git push ci-server master
    Counting objects: 296, done.
    Delta compression using up to 4 threads.
    Compressing objects: 100% (254/254), done.
    Writing objects: 100% (296/296), 193.59 KiB, done.
    Total 296 (delta 25), reused 276 (delta 13)
    -----> Cleaning up ...
    -----> Building  ...
    -----> Deploying $APP ...
    =====> Tests complete.

You're done!

Right now Davis uses the [stack-puzzle](https://github.com/lonnen/stack-puzzle) to create the test-runner environment.

## Remote commands

Davis commands can be run over ssh. Anywhere you would run `dokku <command>`, just run `ssh -t dokku@ci-server.biz <command>`
The `-t` is used to request a pty. It is highly recommended to do so.
To avoid the need to type the `-t` option each time, simply create/modify a section in the `.ssh/config` on the client side, as follows :

    Host ci-server.biz
    RequestTTY yes


## Run a command in the app environment

It's possible to run commands in the environment of the running tests:

    $ dokku run my-project ls -alh
    $ dokku run <app> <cmd>


## Plugins

Davis is based on Dokku, which is itself is built out of plugins. Take a look in the `plugins/` directory for a full enumeration. Load order is lexical, so numbers can be prepended to the name to force earlier execution.

## Removing a set of tests

SSH onto the server, then execute:

    $ dokku delete my-project


## Environment variable management

Typically an application will require some environment variables to run properly. Environment variables may contain private data, such as passwords or API keys, so it is not recommend to store them in your application's repository.

The core `config` plugin provides the following commands to manage your variables:
```
config <app> - display the config vars for an app  
config:get <app> KEY - display a config value for an app  
config:set <app> KEY1=VALUE1 [KEY2=VALUE2 ...] - set one or more config vars
config:unset <app> KEY1 [KEY2 ...] - unset one or more config vars
```

## TLS support

Davis provides easy TLS support from the box, mostly as an artifact of it's Dokku heritage. To enable TLS connections to your running tests, copy the `.crt` and `.key` files into the `/home/dokku/:app/ssl` folder (notice, file names should be `server.crt` and `server.key`, respectively). New test runs will force redirection to `https`.

Chances are you'll never need this feature. It would only be of value to the most rubegoldbergian test suites. It persists because it's more work to remove it than to document it and ignore it.

## Support

You can try [Github Issues](https://github.com/lonnen/dokku/issues), check [Troubleshooting](https://github.com/progrium/lonnen/wiki/Troubleshooting) on the wiki, or find lonnen around the freenode or mozilla irc networks.

The dokku folk might be able to help also, but this fork is starting to deviate in some extreme ways.

## Components

 * [Docker](https://github.com/dotcloud/docker) - Container runtime and manager
 * [Buildstep](https://github.com/progrium/buildstep) - Buildpack builder
 * [pluginhook](https://github.com/progrium/pluginhook) - Shell based plugins and hooks
 * [sshcommand](https://github.com/progrium/sshcommand) - Fixed commands over SSH
 * [stack-puzzle](https://github.com/lonnen/stack-puzzle) - The environment where tests run

Looking to keep codebase as simple and hackable as possible, so try to keep your line count down.

## Things this project won't do

 * **Multi-host.** Not a huge leap, but this isn't the project for it. Have a look at [Flynn](https://flynn.io/).
 * **Multitenancy.** It's ready for it, but again, have a look at [Flynn](https://flynn.io/).
 * **Client app.** Given the constraints, running commands remotely via SSH is fine.
 * **Work** For the time being it's being developer for a bespoke environment as part of an ephemeral coding challenge. It might have some interesting ideas, but in many places its tightly bound to some strange assumptions for the contest that will hamper it from being useful to anyone else.

## License

MIT
