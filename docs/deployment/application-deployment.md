# Deploying to Dokku

> Note: This walkthrough uses the hostname `dokku.me` in commands. When deploying to your own server, you should substitute the domain `dokku.me` for the domain name or IP address associated with your server. Users of the Vagrant VM included with Dokku can use `dokku.me` which points to the IP of the VM.

## Deploy tutorial

Once you have configured Dokku with at least one user, you can deploy applications using `git push`. To quickly see Dokku deployment in action, try using [the Heroku Ruby on Rails "Getting Started" app](https://github.com/heroku/ruby-getting-started).

```shell
# from your local machine
# SSH access to github must be enabled on this host
git clone https://github.com/heroku/ruby-getting-started
```

### Create the app

SSH into the Dokku host and create the application as follows:

```shell
# on the Dokku host
dokku apps:create ruby-getting-started
```

### Create the backing services

Dokku by default **does not** provide datastores (e.g. MySQL, PostgreSQL) on a newly created app. You can add datastore support by installing plugins, and the Dokku project [provides official plugins](/docs/community/plugins.md#official-plugins-beta) for common datastores.

The Getting Started app requires a PostgreSQL service, so install the plugin and create the related service as follows:

```shell
# on the Dokku host
# install the postgres plugin
# plugin installation requires root, hence the user change
sudo dokku plugin:install https://github.com/dokku/dokku-postgres.git

# create a postgres service with the name railsdatabase
dokku postgres:create railsdatabase
```

Each service may take a few moments to create.

### Linking backing services to applications

Once the services have been created, you then set the `DATABASE_URL` environment variable by linking the service, as follows:

```shell
# on the Dokku host
# each official datastore offers a `link` method to link a service to any application
dokku postgres:link railsdatabase ruby-getting-started
```

Dokku supports linking a single service to multiple applications as well as linking only one service per application.

### Deploy the app

> Warning: Your app should respect the `PORT` environment variable, otherwise it may not respond to web requests. You can find more information in the [port management documentation](/docs/networking/port-management.md).**

Now you can deploy the `ruby-getting-started` app to your Dokku server. All you have to do is add a remote to name the app. Applications are created on-the-fly on the Dokku server.

```shell
# from your local machine
# the remote username *must* be dokku or pushes will fail
cd ruby-getting-started
git remote add dokku dokku@dokku.me:ruby-getting-started
git push dokku main:master
```

> Note: Some tools may not support the short-upstream syntax referenced above, and you may need to prefix
> the upstream with the scheme `ssh://` like so: `ssh://dokku@dokku.me:ruby-getting-started`
> Please see the [Git](https://git-scm.com/docs/git-clone#_git_urls_a_id_urls_a) documentation for more details.

> Note: Your private key should be registered with `ssh-agent` in your local development environment. If you get a `permission denied` error when pushing, you can register your private key as follows: `ssh-add -k ~/<your private key>`.

After running `git push dokku main:master`, you should have output similar to this in your terminal:

```
Counting objects: 231, done.
Delta compression using up to 8 threads.
Compressing objects: 100% (162/162), done.
Writing objects: 100% (231/231), 36.96 KiB | 0 bytes/s, done.
Total 231 (delta 93), reused 147 (delta 53)
-----> Cleaning up...
-----> Building ruby-getting-started from herokuish...
-----> Adding BUILD_ENV to build environment...
-----> Ruby app detected
-----> Compiling Ruby/Rails
-----> Using Ruby version: ruby-2.2.1
-----> Installing dependencies using 1.9.7
       Running: bundle install --without development:test --path vendor/bundle --binstubs vendor/bundle/bin -j4 --deployment
       Fetching gem metadata from https://rubygems.org/...........
       Fetching version metadata from https://rubygems.org/...
       Fetching dependency metadata from https://rubygems.org/..
       Using rake 10.4.2

...

=====> Application deployed:
       http://ruby-getting-started.dokku.me
```

Once the deploy is complete, the application's web URL will be generated as above.

Dokku supports deploying applications in a few ways:

- [Heroku buildpacks](https://devcenter.heroku.com/articles/buildpacks) via [Herokuish](https://github.com/gliderlabs/herokuish#buildpacks): See the [herokuish buildpacks documentation](/docs/deployment/builders/herokuish-buildpacks.md) to learn about the different ways to specify a buildpack.
  - This is the default method used by Dokku.
- [Dockerfile](https://docs.docker.com/reference/builder/): See the [dockerfile documentation](/docs/deployment/builders/dockerfiles.md) to learn about the different ways to configure Dockerfile-based deploys.
- [Docker Image](https://docs.docker.com/get-started/overview/#docker-objects): See the [docker image documentation](/docs/deployment/methods/images.md) to learn about how to deploy a Docker Image.

### Setting up SSL

> While SSL certificates can be imported, automated SSL via Letsencrypt requires that all domains on an app correctly point at your server's public ip address. Please keep this in mind when using Letsencrypt.

For many users, responding to requests via `https` will be desirable. Dokku has a complete [ssl plugin](/docs/configuration/ssl.md) built in that can be used to import SSL certificates (below is a short example, please refer to the [ssl documentation](/docs/configuration/ssl.md) for more information):

```shell
dokku certs:add ruby-getting-started server.crt server.key
```

As an alternative, the Dokku project offers an optional letsencrypt plugin that can be used to automate SSL certificate retrieval and renewal.

```shell
# on the Dokku host
# install the letsencrypt plugin
# plugin installation requires root, hence the user change
sudo dokku plugin:install https://github.com/dokku/dokku-letsencrypt.git

# configure the plugin
dokku config:set --global DOKKU_LETSENCRYPT_EMAIL=your-email@your.domain.com

# set a custom domain that you own for your application
dokku domains:set ruby-getting-started ruby-getting-started.your.domain.com

# enable letsencrypt
dokku letsencrypt:enable ruby-getting-started

# enable auto-renewal
dokku letsencrypt:cron-job --add
```

### Skipping deployment

If you only want to rebuild and tag a container, you can skip the deployment phase by setting `$DOKKU_SKIP_DEPLOY` to `true` by running:

``` shell
# on the Dokku host
dokku config:set ruby-getting-started DOKKU_SKIP_DEPLOY=true
```

### Redeploying or restarting

If you need to redeploy or restart your app: 

```shell
# on the Dokku host
dokku ps:rebuild ruby-getting-started
```

See the [process scaling documentation](/docs/processes/process-management.md) for more information on how to manage your app processes.

### Deploying with private Git submodules

Dokku uses Git locally (i.e. not a Docker image) to build its own copy of your app repo, including submodules, as the `dokku` user. This means that in order to deploy private Git submodules, you need to put your deploy key in `/home/dokku/.ssh/` and potentially add `github.com` (or your VCS host key) into `/home/dokku/.ssh/known_hosts`. You can use the following test to confirm your setup is correct:

```shell
# on the Dokku host
su - dokku
ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts
ssh -T git@github.com
```

> Warning: if the buildpack or Dockerfile build process require SSH key access for other reasons, the above may not always apply.

## Deploying to subdomains

If you do not enter a fully qualified domain name when pushing your app, Dokku deploys the app to `<remotename>.yourdomain.tld` as follows:

```shell
# from your local machine
# the remote username *must* be dokku or pushes will fail
# the below example assumes your app server domain or IP is dokku.me. Push in the form of: dokku@{serveripordomain}:{dokkuappname} 
git remote add dokku dokku@dokku.me:ruby-getting-started
git push dokku main:master
```

```
remote: -----> Application deployed:
remote:        http://ruby-getting-started.dokku.me
```

You can also specify the fully qualified name as follows:

```shell
# from your local machine
# the remote username *must* be dokku or pushes will fail
git remote add dokku dokku@dokku.me:app.dokku.me
git push dokku main:master
```

```
remote: -----> Application deployed:
remote:        http://app.dokku.me
```

This is useful when you want to deploy to the root domain:

```shell
# from your local machine
# the remote username *must* be dokku or pushes will fail
git remote add dokku dokku@dokku.me:dokku.me
git push dokku main:master
```

```
... deployment ...

remote: -----> Application deployed:
remote:        http://dokku.me
```

## Dokku/Docker container management compatibility

Dokku is, at its core, a Docker container manager. Thus, it does not necessarily play well with other out-of-band processes interacting with the Docker daemon.

Prior to every deployment, Dokku will execute a cleanup function. As of 0.5.x, the cleanup removes all containers with the `dokku` label where the status is either `dead` or `exited` (previous versions would remove _all_ `dead` or `exited` containers). The cleanup function also removes all images with `dangling` status.

## Adding deploy users

See the [user management documentation](/docs/deployment/user-management.md) for more information on how to manage users with access to your Dokku server.

## Default vhost

See the [domains documentation](/docs/configuration/domains.md#default-site) for more information on how to manage the default site.

## Deploying non-master branch

See the [Git documentation](/docs/deployment/methods/git.md#changing-the-deploy-branch) for more information on deploying a non-master branch to your application.

## Dockerfile deployment

See the [Dockerfile documentation](/docs/deployment/builders/dockerfiles.md) for information Dokku's Dockerfile support.

## Image tagging

See the [image tagging documentation](/docs/deployment/methods/images.md) for more information on how Docker images can be tagged and deployed for a given application.

## Specifying a custom buildpack

See the [herokuish buildpack documentation](/docs/deployment/builders/herokuish-buildpacks.md) for more information on how to specify a set of custom buildpacks for your application.

## Removing a deployed app

See the [application management documentation](/docs/deployment/application-management.md#removing-a-deployed-app) for more information on how to remove an application from your Dokku server.

## Renaming a deployed app

See the [application management documentation](/docs/deployment/application-management.md#renaming-a-deployed-app) for more information on how an application can be renamed and the impact of doing so upon the application and associated resources.

## Zero downtime deploy

See the [zero-downtime deploy documentation](/docs/deployment/zero-downtime-deploys.md) for more information on how Dokku enables zero-downtime deploys.
