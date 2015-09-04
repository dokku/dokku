# Deploy an App

Now you can deploy apps on your Dokku. Let's deploy the [Heroku Node.js sample app](https://github.com/heroku/node-js-sample). All you have to do is add a remote to name the app. It's created on-the-fly.

```
$ cd node-js-sample
$ git remote add dokku dokku@dokku.me:node-js-app
$ git push dokku master
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
       http://node-js-app.dokku.me
```

You're done!

Dokku only supports deploying from its master branch, so if you'd like to deploy a different local branch use: ```git push dokku <local branch>:master```

Right now Herokuish supports buildpacks for Node.js, Ruby, Python, [and more](https://github.com/gliderlabs/herokuish#buildpacks).
Please check the documentation for your particular build pack as you may need to include configuration files (such as a Procfile) in your project root.

## Deploying to server over SSH

Pushing to the dokku remote may prompt you to input a password for the dokku user. It's preferable, however, to use key-based authentication, and you can add your public key to the dokku user's authorized_keys file with:

```
cat ~/.ssh/id_rsa.pub | ssh [sudouser]@[yourdomain].com "sudo sshcommand acl-add dokku [description]"
```

## Deploying with private git submodules

Dokku uses git locally (i.e. not a docker image) to build its own copy of your app repo, including submodules. This is done as the `dokku` user. Therefore, in order to deploy private git submodules, you'll need to drop your deploy key in `~dokku/.ssh` and potentially add github.com (or your VCS host key) into `~dokku/.ssh/known_hosts`. A decent test like this should help confirm you've done it correctly.

```
su - dokku
ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts
ssh -T git@github.com
```

## Specifying a custom buildpack

If buildpack detection isn't working well for you or you want to specify a custom buildpack for one repository you can create & commit a file in the root of your git repository named `.env` containing `export BUILDPACK_URL=<repository>` before pushing. This will tell herokuish to fetch the specified buildpack and use it instead of relying on the built-in buildpacks & their detection methods.

## Dockerfile deployment

Deployment of `Dockerfile` repos is supported as of v0.3.15. Simply place a Dockerfile in the root of your repo and push to dokku. If you are converting from a heroku/dokku buildpack deployment, ensure you are not setting `$BUILDPACK_URL` or including a `.buildpacks` file in the root of your repo.

By default, we will extract the first `EXPOSE` port and tell nginx to proxy your app to that port. Alternatively, you can set `$DOKKU_DOCKERFILE_PORT` in your app's dokku configuration.

By default no arguments are passed to `docker run` when deploying the container and the `CMD` or `ENTRYPOINT` defined in the `Dockerfile` are executed. You can take advantage of docker ability of overriding the `CMD` or passing parameters to your `ENTRYPOINT` setting `$DOKKU_DOCKERFILE_START_CMD`. Let's say for example you are deploying a base nodejs image, with the following `ENTRYPOINT`:

```
ENTRYPOINT ["node"]
```

You can do:

```
dokku config:set DOKKU_DOCKERFILE_START_CMD="--harmony server.js"
```

To tell docker what to run.

Setting `$DOKKU_DOCKERFILE_CACHE_BUILD` to `true` or `false` will enable or disable docker's image layer cache. Lastly, for more granular build control, you may also pass any `docker build` option to `docker`, by setting `$DOKKU_DOCKER_BUILD_OPTS`.

## Default vhost

You might notice the default vhost for Nginx will be one of the apps. If an app doesn't exist, it will use this vhost and it may be confusing for it to go to another app. You can create a default vhost using a configuration under `sites-enabled`. You just have to change one thing in the main nginx.conf:

Swap both conf.d include line and the sites-enabled include line. From:
```
include /etc/nginx/conf.d/*.conf;
include /etc/nginx/sites-enabled/*;
```
to
```
include /etc/nginx/sites-enabled/*;
include /etc/nginx/conf.d/*.conf;
```

Alternatively, you may push an app to your dokku host with a name like "00-default". As long as it lists first in `ls /home/dokku/*/nginx.conf | head`, it will be used as the default nginx vhost.

## Deploying to subdomains

The name of remote repository is used as the name of application to be deployed, as for example above:

    $ git remote add dokku dokku@dokku.me:node-js-app
    $ git push dokku master

Is deployed to,

    remote: -----> Application deployed:
    remote:        http://node-js-app.dokku.me

You can also specify fully qualified names, say `app.dokku.me`, as

    $ git remote add dokku dokku@dokku.me:app.dokku.me
    $ git push dokku master

So, after deployment the application will be available at,

    remote: -----> Application deployed:
    remote:        http://app.dokku.me

This is in particular useful, then you want to deploy to root domain, as

    $ git remote add dokku dokku@dokku.me:dokku.me
    $ git push dokku master

    ... deployment ...

    remote: -----> Application deployed:
    remote:        http://dokku.me

## Zero downtime deploy

Following a deploy, dokku will now wait `DOKKU_DEFAULT_CHECKS_WAIT` seconds (default: `10`), and if the container is still running, then route traffic to the new container.

This can be problematic for applications whose boot up time can vary and can lead to `502 Bad Gateway` errors.

Dokku provides a way to run a set of more precise checks against the new container, and only switch traffic over if all checks complete successfully.

To specify checks, add a `CHECKS` file to the root of your project directory. This is a text file with one line per check. Empty lines and lines starting with `#` are ignored.

A check is a relative URL and may be followed by expected content from the page, for example:

```
/about      Our Amazing Team
```

Dokku will wait `DOKKU_CHECKS_WAIT` seconds (default: `5`) before running the checks to give server time to start. For shorter/longer wait, change the `DOKKU_CHECKS_WAIT` environment variable.  This can also be overridden in the CHECKS file by setting WAIT=nn.

Dokku will wait `DOKKU_WAIT_TO_RETIRE` seconds (default: `60`) before stopping the old container such that no existing connections to it are dropped.

Dokku will retry the checks DOKKU_CHECKS_ATTEMPTS times until the checks are successful or DOKKU_CHECKS_ATTEMPTS is exceeded.  In the latter case, the deployment is considered failed. This can be overridden in the CHECKS file by setting ATTEMPTS=nn.

Checks can be skipped entirely by setting `DOKKU_SKIP_ALL_CHECKS` to `true` either globally or per application. You can choose to skip only default checks by setting `DOKKU_SKIP_DEFAULT_CHECKS` to `true` either globally or per application.

See [checks-examples.md](checks-examples.md) for examples and output.

## Removing a deployed app

SSH onto the server, then execute:

```shell
dokku apps:destroy myapp
```

## Image tagging

The dokku tags plugin allows you to add docker image tags to the currently deployed app image for versioning and subsequent deployment.

```
tags <app>                                       List all app image tags
tags:create <app> <tag>                          Add tag to latest running app image
tags:deploy <app> <tag>                          Deploy tagged app image
tags:destroy <app> <tag>                         Remove app image tag
```

Example:
```
root@dokku:~# dokku tags node-js-app
=====> Image tags for dokku/node-js-app
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
dokku/node-js-app   latest              936a42f25901        About a minute ago   1.025 GB

root@dokku:~# dokku tags:create node-js-app v0.9.0
=====> Added v0.9.0 tag to dokku/node-js-app

root@dokku:~# dokku tags node-js-app
=====> Image tags for dokku/node-js-app
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
dokku/node-js-app   latest              936a42f25901        About a minute ago   1.025 GB
dokku/node-js-app   v0.9.0              936a42f25901        About a minute ago   1.025 GB

root@dokku:~# dokku tags:deploy node-js-app v0.9.0
-----> Releasing node-js-app (dokku/node-js-app:v0.9.0)...
-----> Deploying node-js-app (dokku/node-js-app:v0.9.0)...
-----> Running pre-flight checks
       For more efficient zero downtime deployments, create a file CHECKS.
       See http://progrium.viewdocs.io/dokku/checks-examples.md for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
=====> node-js-app container output:
       Detected 512 MB available memory, 512 MB limit per process (WEB_MEMORY)
       Recommending WEB_CONCURRENCY=1
       > node-js-sample@0.1.0 start /app
       > node index.js
       Node app is running at localhost:5000
=====> end node-js-app container output
-----> Running post-deploy
-----> Configuring node-js-app.dokku.me...
-----> Creating http nginx.conf
-----> Running nginx-pre-reload
       Reloading nginx
-----> Shutting down old containers in 60 seconds
=====> 025eec3fa3b442fded90933d58d8ed8422901f0449f5ea0c23d00515af5d3137
=====> Application deployed:
       http://node-js-app.dokku.me

```

## Dokku/Docker Container Management Compatibility

Dokku is, at it's core, a docker container manager. Thus, it does not necessarily play well with other out-of-band processes interacting with the docker daemon. One thing to note as in [issue #1220](https://github.com/progrium/dokku/issues/1220), dokku executes a cleanup function prior to every deployment. This function removes all exited containers and all 'unattached' images.
