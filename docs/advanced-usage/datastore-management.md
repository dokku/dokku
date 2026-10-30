# Datastore Management

```
datastore:app-links <datastore-type> [<app>]                                                       # list all datastore service links for a given app
datastore:backup <datastore-type> <service> <bucket-name> [--id <id>]...                           # create a backup of the datastore service to an existing s3 bucket
datastore:backup-auth <datastore-type> <service> <aws-access-key-id> <aws-secret-access-key>...    # set up authentication for backups on the datastore service
datastore:backup-deauth <datastore-type> <service> [--id <id>]                                     # remove backup authentication for the datastore service
datastore:backup-logs <datastore-type> <service> [-t|--tail [<tail-num>]]                          # print the most recent output of the scheduled backups of the service
datastore:backup-schedule <datastore-type> <service> <schedule> <bucket-name>...                   # schedule a backup of the datastore service
datastore:backup-schedule-cat <datastore-type> <service> [--id <id>]                               # cat the crontab line of a scheduled backup for the service
datastore:backup-set-encryption <datastore-type> <service> <passphrase> [--id <id>]                # set encryption for all future backups of datastore service
datastore:backup-set-public-key-encryption <datastore-type> <service> <public-key-id> [--id <id>]  # set GPG Public Key encryption for all future backups of datastore service
datastore:backup-unschedule <datastore-type> <service> [--id <id>|--all]                           # unschedule the backup of the datastore service
datastore:backup-unset-encryption <datastore-type> <service> [--id <id>]                           # unset encryption for future backups of the datastore service
datastore:backup-unset-public-key-encryption <datastore-type> <service> [--id <id>]                # unset GPG Public Key encryption for future backups of the datastore service
datastore:clone <datastore-type> <service> <new-service> [--clone-flags...]                        # create container <new-name> then copy data from <name> into <new-name>
datastore:connect <datastore-type> <service>                                                       # connect to the service via the <datastore-type> connection tool
datastore:create <datastore-type> <service> [--create-flags...]                                    # create a datastore service
datastore:create-database <datastore-type> <service> <database>                                    # create a database in the datastore service
datastore:create-user <datastore-type> <service> <user> [<password>]                               # create a user in the datastore service
datastore:destroy <datastore-type> <service> [-f|--force]                                          # delete the datastore service/data/container if there are no links left
datastore:destroy-database <datastore-type> <service> <database> [-f|--force]                      # destroy a database created in the datastore service
datastore:destroy-user <datastore-type> <service> <user> [-f|--force]                              # destroy a user created in the datastore service
datastore:enter <datastore-type> <service>                                                         # enter or run a command in a running datastore service container
datastore:exists <datastore-type> <service>                                                        # check if the datastore service exists
datastore:export <datastore-type> <service> [-f|--file <path>] [--force]                           # export a dump of the datastore service database
datastore:expose <datastore-type> <service> <ports...>                                             # expose a datastore service on custom host:port if provided (random port on the 0.0.0.0 interface if otherwise unspecified)
datastore:help [<datastore-type> [<command>]]                                                      # print the help for every datastore, one datastore or one of its commands
datastore:import <datastore-type> <service> [-f|--file <path>]                                     # import a dump into the datastore service database
datastore:info <datastore-type> [<service>] [--info-flags...]                                      # print the service information
datastore:label <datastore-type> <service> <key=value>...                                          # add docker labels to the service container
datastore:link <datastore-type> <service> [<app>] [--link-flags...]                                # link the datastore service to the app
datastore:linked <datastore-type> <service> [<app>]                                                # check if the datastore service is linked to an app
datastore:links <datastore-type> <service>                                                         # list all apps linked to the datastore service
datastore:list <datastore-type>                                                                    # list all datastore services
datastore:logs <datastore-type> <service> [-t|--tail [<tail-num>]]                                 # print the most recent log(s) for this service
datastore:mount <datastore-type> [--replace] <service> <source:container-dir[:options]>...         # mount a host path or docker volume into the service container
datastore:pause <datastore-type> <service>                                                         # pause a running datastore service
datastore:promote <datastore-type> <service> [<app>]                                               # promote service <service> as DATABASE_URL in <app>
datastore:reexpose <datastore-type> <service>                                                      # reexpose a datastore service, applying its expose settings
datastore:reset <datastore-type> <service> [-f|--force]                                            # delete all data in the datastore service, keeping the service and its links
datastore:restart <datastore-type> <service>                                                       # graceful shutdown and restart of the datastore service container
datastore:set <datastore-type> <service> <key> [<value>]...                                        # set or clear a property for a service
datastore:set-user-password <datastore-type> <service> <user> [<password>]...                      # change the password of a user created in the datastore service
datastore:start <datastore-type> <service>                                                         # start a previously stopped datastore service
datastore:stop <datastore-type> <service>                                                          # stop a running datastore service
datastore:unexpose <datastore-type> <service>                                                      # unexpose a previously exposed datastore service
datastore:unlabel <datastore-type> [--all] <service> [<key>...]                                    # remove one or all docker labels from the service container
datastore:unlink <datastore-type> <service> [<app>] [-n|--no-restart]                              # unlink the datastore service from the app
datastore:unmount <datastore-type> [--all] <service> [<source:container-dir>...]                   # remove one or all mounts from the service container
datastore:upgrade <datastore-type> <service> [--upgrade-flags...]                                  # upgrade service <service> to the specified versions
datastore:version                                                                                  # print the version of dokku-datastore the plugin runs
```

The `datastore` plugin manages datastore services such as postgres, mysql and redis, and links them to apps. Every command takes the datastore type as its first argument, and the datastores it serves are the ones the [official datastore plugins](/docs/community/plugins.md#official-plugins) provide: clickhouse, couchdb, elasticsearch, graphite, mariadb, meilisearch, memcached, mongo, mysql, nats, omnisci, postgres, pushpin, rabbitmq, redis, rethinkdb, solr and typesense.

The plugin is built from [dokku-datastore](https://github.com/dokku/dokku-datastore), which the official datastore plugins also run, and keeps its services in the same place they do, under `/var/lib/dokku/services`.

## Usage

### Creating and linking a service

A service is created with `datastore:create` and linked to an app with `datastore:link`, which sets the variable the app connects with, such as `DATABASE_URL` for postgres or `REDIS_URL` for redis, and restarts the app.

```shell
dokku datastore:create postgres lollipop
dokku datastore:link postgres lollipop node-js-app
```

`datastore:unlink` removes the link, and `datastore:destroy` removes a service once nothing is linked to it.

```shell
dokku datastore:unlink postgres lollipop node-js-app
dokku datastore:destroy postgres lollipop
```

Deleting an app unlinks it from every service it is linked to.

### Getting help

`datastore:help` lists the datastore types and the commands they share. Not every datastore implements every command, so the help for one datastore lists only the commands it implements, including ones only it has, such as `connect-admin` for mongo. The help for one of a datastore's commands describes its flags and shows examples.

```shell
dokku datastore:help
dokku datastore:help postgres
dokku datastore:help postgres create
```

### Scheduled backups

A backup scheduled with `datastore:backup-schedule` is added to the dokku crontab as a `dokku datastore:backup` command, and is managed with the [cron plugin](/docs/processes/scheduled-cron-tasks.md) like any other scheduled task.

```shell
dokku datastore:backup-auth postgres lollipop AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY
dokku datastore:backup-schedule postgres lollipop "0 3 * * *" my-s3-bucket
```

### Images

Installing dokku does not pull the image of every datastore. A datastore's images are pulled the first time a command needs them, and on later installs and upgrades once the datastore has a service.

### Official datastore plugins

The official datastore plugins, such as [dokku-postgres](https://github.com/dokku/dokku-postgres), may be installed alongside the `datastore` plugin. While one is enabled, the `datastore` plugin leaves that datastore to it when dokku runs plugin triggers, such as unlinking an app that is being deleted or writing scheduled backups into the crontab. Commands for that datastore should be run through the official plugin until it is removed.

The `datastore` plugin does not read the datastore definitions an official plugin ships, so a host running a modified official plugin should keep using it.

#### Migrating from an official datastore plugin

Moving a datastore from its official plugin to the `datastore` plugin leaves its services, containers, data, settings and links where they are, and does not stop running services or the apps linked to them. Using postgres as an example:

1. Check that `dokku datastore:version` is at least the version dokku-postgres runs, which is printed by `/var/lib/dokku/data/postgres/dokku-datastore version`. If the official plugin runs a newer version, wait for a dokku release that ships it.
2. Uninstall the official plugin.

    ```shell
    dokku plugin:uninstall postgres
    ```

3. Run every plugin's install trigger, which sets the host up for postgres, brings its services up to date and rewrites the crontab so scheduled backups run through `dokku datastore:backup`.

    ```shell
    dokku plugin:install
    ```

4. Change anything that runs `dokku postgres:<command> <service>` to run `dokku datastore:<command> postgres <service>`.

To move back, install the official plugin again from a release running at least the version `dokku datastore:version` prints.
