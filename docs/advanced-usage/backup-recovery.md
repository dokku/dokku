# Backup and Recovery

> The backup plugin was deprecated in 0.4.x, below are backup recommendations for 0.5.x and later.

## Backup and Migration Tutorial

Because Dokku is git based, rebuilding a deployed app is as easy as pushing from git. You can push to a new server by updating the `dokku` remote in you local app's repo.

```shell
git remote rm dokku
git remote add dokku dokku@[dokku.me:dokku.me]
git push dokku [master]
```
## Databases

Some plugins, like the official [dokku-postgres](https://github.com/dokku/dokku-postgres) plugin, have built-in commands that allow non-volatile data be exported and imported.

For [dokku-postgres](https://github.com/dokku/dokku-postgres), use:

```shell
dokku postgres:export [db_name] > [db_name].dump
dokku postgres:import [db_name] < [db_name].dump
```

## Volumes and Static Assets

Dokku doesn't enforce a [300mb](https://devcenter.heroku.com/articles/slug-compiler#slug-size) limit on apps, but it's best practice to keep binary assets outside of git. Since containers are considered volatile in Dokku, external stores like s3 or storage mounts should be used for non-volatile items like user uploads. The Dokku storage core plugin can be used to mount local directories / volumes inside the docker container.

See the [persistent storage documentation](/docs/advanced-usage/persistent-storage.md) for more information on how to attach persistent storage to your application.

## Disaster Recovery

The best plan for disaster recovery is to always keep multiple (remote) copies of your local repo, static assets and periodic database dumps. In case of an emergency when your git repo and backups are completely lost, you can recover the last pushed copy from your remote Dokku server (assuming you still have the ssh key).

```shell
mkdir [app-name] ; cd !$
git init && git remote add dokku dokku@[dokku.me:app-name]
git pull dokku/master && git checkout dokku/master
```
