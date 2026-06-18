# Backup and Recovery

> The backup plugin was deprecated in 0.4.x, below are backup recommendations for 0.5.x and later.

The best plan for disaster recovery is to always keep multiple (remote) copies of your local repo, static assets and periodic database dumps. Backups should be regularly tested for data integrity and completeness.

## Centralized export and import

> [!IMPORTANT]
> The `backup:export` and `backup:import` commands were introduced in 0.39.0.

Dokku can export an app (and/or service, and/or the global configuration) into a single `tar.gz` archive and restore it again, either on the same host or a fresh one. Each plugin serializes its own slice of state, so the archive captures config, domains, ports, docker-options, process scale, the git repository, and any other plugin's data.

```shell
# export everything (all apps, all services, global config) to /tmp
dokku backup:export

# export a single app
dokku backup:export --app node-js-app

# export specific services (datastore plugins must implement the service-list trigger)
dokku backup:export --service postgres:mydb --service redis:cache

# write the archive somewhere else and include persistent storage volume data
dokku backup:export --app node-js-app --backup-dir /mnt/backups --include-storage
```

The full path to the created archive is the only thing written to `stdout`; all progress and warnings are written to `stderr`, so the path can be captured directly:

```shell
BACKUP_FILE="$(dokku backup:export --app node-js-app)"
```

To restore, pass the archive to `backup:import`:

```shell
# restore everything in the archive
dokku backup:import /tmp/dokku-backup-full-20260618T120000Z.tar.gz

# restore only a single app or service from the archive
dokku backup:import --app node-js-app "$BACKUP_FILE"
dokku backup:import --service postgres:mydb "$BACKUP_FILE"
```

Importing is destructive. If an app or service already exists, the import aborts and asks you to re-run with `--force` to replace it:

```shell
dokku backup:import --force "$BACKUP_FILE"
```

A backup records the third-party plugins that were installed (by name and git remote). By default they are reinstalled from their remotes first, before any other restore step, so that datastore and other plugins exist before their state is restored. Pass `--skip-install-plugins` to only report them instead:

```shell
dokku backup:import --skip-install-plugins "$BACKUP_FILE"
```

### Notes and caveats

- **Secrets are stored unencrypted.** The archive contains environment variables and TLS material in plaintext. Store and transfer it securely.
- **Disk space.** Export buffers the archive to the `--backup-dir` (default `/tmp`); import extracts the whole archive before restoring. Ensure roughly twice the archive size is free.
- **Persistent storage data** is only included when `--include-storage` is passed, and only for dokku-managed storage directories. Mount declarations are always captured.
- **Datastore services** are provided by third-party plugins. They are auto-discovered for a full export only when their plugin implements the `datastore-list` and `service-list` triggers; otherwise pass `--service TYPE:NAME` explicitly.
- **Access control.** Callers can only export and import apps and services they have access to via the [user-auth](/docs/development/plugin-triggers.md#user-auth) plugin trigger.
- **Generated config** (nginx vhost files, proxy config) and **log history** are not backed up; they are regenerated when the app is redeployed.
- **Portability.** The declarative `config/*.yml` slices are [docket](https://github.com/dokku/docket) (>= 0.6.0) recipes; a backup also contains aggregate `tasks.yml` recipes that `docket apply` can consume for out-of-band restores.
- **Restore order.** An import reinstalls recorded plugins first, then restores global state, then services, then each app. Each app is created, has its config restored (`backup-app-import`, then `post-backup-app-import` for config that depends on another plugin's restored config such as `domains`), and is finally redeployed. Because the redeploy is last, all config is restored before the app is built, and plugins that need the running app (for example issuing TLS certificates) act during the normal deploy. See the [import-ordering note in the plugin triggers docs](/docs/development/plugin-triggers.md#backup-pre-and-post-hooks).

## TLDR

> [!WARNING]
> This method has many caveats. Please read this entire document before assuming these backups work as expected, and test your backups on a regular basis.

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

At this point, all datastores should be individually started and checked for data integrity. Once this is complete, individual apps can be rebuilt. Please consult the [process management documentation](/docs/processes/process-management.md#rebuilding-apps) for more information on how to rebuild apps.

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

### Docker Networks

Docker networks generated by Dokku should be recreated. Running `dokku network:report` will output all networks in use by various apps, which can then be used to recreate them via `dokku network:create $NETWORK`.

Networks created by tools other than Dokku may be created as they initially were.

### Docker Image and Tar-based Apps

These apps may fail to rebuild via the normal `ps:rebuild` method. Redeploy these apps by running the original commands used to deploy them

### Datastores

> Please note that point-in-time backups of the `/var/lib/dokku/services` directory may contain partially written data due to how various datastores work. Consult the official datastore documentation for the best documentation surrounding proper backup and restore procedures.

Some plugins, like the official [dokku-postgres](https://github.com/dokku/dokku-postgres) plugin, have built-in commands that allow non-volatile data be exported and imported.

For [dokku-postgres](https://github.com/dokku/dokku-postgres), use:

```shell
dokku postgres:export [db_name] > [db_name].dump
dokku postgres:import [db_name] < [db_name].dump
```

Additionally, data for official datastores is located in the `/var/lib/dokku/services` directory. If the directory is restored and the plugin is available, a `dokku $SERVICE:start` may be enough to restart the service with the underlying data, so long as the datastore version does not change and the underlying data is not corrupt. If this is the case, it may be necessary to re-import all the data onto a fresh version of the datastore service.

### Plugins

The plugin directory is contained at the `/var/lib/dokku/plugins` directory. Core plugins will automatically be included in new installs, but custom plugins may not. The aforementioned `tar` creation command will back all plugins up, and the `tar` extract command will restore the plugins.

Note that restoring a plugin will not trigger any `install` or `dependencies` triggers. You will need to run these manually. See the [plugin management documentation](/docs/advanced-usage/plugin-management.md#installing-a-plugin) for more information on how to trigger these two hooks.

### Volumes and Static Assets

Dokku doesn't enforce a [300mb](https://devcenter.heroku.com/articles/slug-compiler#slug-size) limit on apps, but it's best practice to keep binary assets outside of git. Since containers are considered volatile in Dokku, external stores like s3 or storage mounts should be used for non-volatile items like user uploads. The Dokku storage core plugin can be used to mount local directories / volumes inside the docker container.

System administrators are highly encouraged to store persistent data in app-specific subdirectories of the path `/var/lib/dokku/data/storage`. This will help ensure restores of the aforementioned primary Dokku directories will restore service to all apps.

See the [persistent storage documentation](/docs/advanced-usage/persistent-storage.md) for more information on how to attach persistent storage to your app.

## Recovering app code

In case of an emergency when your git repo and backups are completely lost, you can recover the last pushed copy from your remote Dokku server (assuming you still have the ssh key).

```shell
mkdir [app-name] ; cd !$
git init && git remote add dokku dokku@[dokku.me:app-name]
git pull dokku/master && git checkout dokku/master
```
