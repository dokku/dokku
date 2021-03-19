# Backup and Recovery
----

!!! attention "The backup plugin was deprecated in 0.4.x, below are backup recommendations for 0.5.x and later."

The best plan for disaster recovery is to always keep multiple (remote) copies of your local repo, static assets and periodic database dumps. Backups should be regularly tested for data integrity and completeness.

## TLDR

!!! warning
    This method has many caveats. Please read this entire document before assuming these backups work as expected, and test your backups on a regular basis.

### Creating a backup

To create a backup, run the following command at a time when not executing any Dokku commands or app deployments:

```shell
export BACKUP_TIME=$(date +%Y-%m-%d-%H-%M)
sudo mkdir -p /var/lib/dokku/services
sudo chown dokku:dokku /var/lib/dokku/services
mkdir -p /tmp/dokku-backups/
sudo tar -czvf "/tmp/dokku-backups/${BACKUP_TIME}.tar.gz" /home/dokku /var/lib/dokku/config /var/lib/dokku/data /var/lib/dokku/services /var/lib/dokku/plugins
```

This will create a point-in-time backup of your entire Dokku installation in the `/tmp/dokku-backups` directory. This backup may be moved offsite to another location via rsync, sftp, or some other protocol.

It is recommended that backups are regularly cleaned from the originating server and tested as necessary.

### Restoring a backup

To extract the backup onto another server, copy the backup to the server and extract it using the following command.

```shell
sudo tar -xzvf path/to/dokku/backup.tar.gz -C /
```

At this point, all datastores should be individually started and checked for data integrity. Once this is complete, individual applications can be rebuilt. Please consult the [process management documentation](/processes/process-management#rebuilding-apps) for more information on how to rebuild applications.

## Caveats

### App config

Application config is largely held in a small number of places:

- `/var/lib/dokku/config`: Properties set and managed by plugins
- `/var/lib/dokku/data`: Files generated or extracted by various plugins
- `/home/dokku`: Certain parts of Dokku core store data in this location

Compressing these directories when no Dokku commands are running is enough to ensure a complete backup of the system.

### Code Repositories

Because Dokku is git based, rebuilding a deployed app is as easy as pushing from git. You can push to a new server by updating the `dokku` remote in you local app's repo.

```shell
git remote rm dokku
git remote add dokku dokku@[dokku.me:dokku.me]
git push dokku [master]
```

### Docker Image and Tar-based Apps

These applications may fail to rebuild via the normal `ps:rebuild` method. Redeploy these apps by running the original commands used to deploy them

### Datastores

!!! info
    Please note that point-in-time backups of the `/var/lib/dokku/services` directory may contain partially written data due to how various datastores work. Consult the official datastore documentation for the best documentation surrounding proper backup and restore procedures.

Some plugins, like the official [dokku-postgres](https://github.com/dokku/dokku-postgres) plugin, have built-in commands that allow non-volatile data be exported and imported.

For [dokku-postgres](https://github.com/dokku/dokku-postgres), use:

```shell
dokku postgres:export [db_name] > [db_name].dump
dokku postgres:import [db_name] < [db_name].dump
```

Additionally, data for official datastores is located in the `/var/lib/dokku/services` directory. If the directory is restored and the plugin is available, a `dokku $SERVICE:start` may be enough to restart the service with the underlying data, so long as the datastore version does not change and the underlying data is not corrupt. If this is the case, it may be necessary to re-import all the data onto a fresh version of the datastore service.

### Plugins

The plugin directory is contained at the `/var/lib/dokku/plugins` directory. Core plugins will automatically be included in new installs, but custom plugins may not. The aforementioned `tar` creation command will back all plugins up, and the `tar` extract command will restore the plugins.

Note that restoring a plugin will not trigger any `install` or `dependencies` triggers. You will need to run these manually. See the [plugin management documentation](/advanced-usage/plugin-management#installing-a-plugin) for more information on how to trigger these two hooks.


### Volumes and Static Assets

Dokku doesn't enforce a [300mb](https://devcenter.heroku.com/articles/slug-compiler#slug-size) limit on apps, but it's best practice to keep binary assets outside of git. Since containers are considered volatile in Dokku, external stores like s3 or storage mounts should be used for non-volatile items like user uploads. The Dokku storage core plugin can be used to mount local directories / volumes inside the docker container.

System administrators are highly encouraged to store persistent data in app-specific subdirectories of the path `/var/lib/dokku/data/storage`. This will help ensure restores of the aforementioned primary Dokku directories will restore service to all apps.

See the [persistent storage documentation](/advanced-usage/persistent-storage) for more information on how to attach persistent storage to your application.

## Recovering app code

In case of an emergency when your git repo and backups are completely lost, you can recover the last pushed copy from your remote Dokku server (assuming you still have the ssh key).

```shell
mkdir [app-name] ; cd !$
git init && git remote add dokku dokku@[dokku.me:app-name]
git pull dokku/master && git checkout dokku/master
```
