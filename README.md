# Dokku

Docker powered mini-Heroku. The smallest PaaS implementation you've ever seen.

## Requirements

Assumes Ubuntu 13 right now. Ideally name the host so it gets/has a hostname of a domain you have pointing to it.

## Installing

    $ wget -qO- j.mp/dokku-bootstrap | bash

This may take around 5 minutes.

## Configuring

Set up a domain and a wildcard domain pointing to that host. Make sure `/home/git/DOMAIN` is set to this domain. 
By default it's set to whatever the hostname the host has.

You'll have to add a public key associated with a username as it says at the end of the bootstrapper. You'll do something
like this:

    $ cat ~/.ssh/id_rsa.pub | ssh root@progriumapp.com "gitreceive upload-key progrium"

That's it!

## Deploy an App

Right now Buildstep supports the Node.js and Ruby buildpacks. It's not hard to add more, go add more! Let's deploy
the Heroku Node.js sample app. All you have to do is add a remote to name the app. It's created on the fly.

    $ cd node-js-sample
    $ git remote add progrium git@progriumapp.com:node-js-app
    $ git push progrium master
    Counting objects: 296, done.
    Delta compression using up to 4 threads.
    Compressing objects: 100% (254/254), done.
    Writing objects: 100% (296/296), 193.59 KiB, done.
    Total 296 (delta 25), reused 276 (delta 13)
    remote: ----> Receiving node-js-app ... 
    remote: -----> Building node-js-app ...
    remote: Node.js app detected
    remote: -----> Resolving engine versions
    
    ... blah blah blah ...
    
    remote: -----> Application deployed:
    remote:        http://node-js-app.progriumapp.com

You're done!

## Components

 * [Docker](https://github.com/dotcloud/docker) - Container runtime and manager
 * [Buildstep](https://github.com/progrium/buildstep) - Buildpack builder
 * [gitreceive](https://github.com/progrium/gitreceive) - Git push interface

## Contributors

 * Jeff Lindsay <progrium@gmail.com>

## License

MIT