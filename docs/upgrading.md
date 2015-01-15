# Upgrading

Dokku is in active development. You can update the deployment step and the build step separately.

**Note**: If you are upgrading from a revision prior to [27d4bc8c3c](https://github.com/progrium/dokku/commit/27d4bc8c3c19fe580ef3e65f2f85b85101cd83e4), follow the instructions in [this wiki entry](https://github.com/progrium/dokku/wiki/Migrating-to-Dokku-0.2.0).

To update the deploy step (this is updated less frequently):

```shell
cd ~/dokku
git pull origin master
sudo make install
```

Nothing needs to be restarted. Changes will take effect on the next push / deployment.

To update the build step:

```shell
git clone https://github.com/progrium/buildstep.git
cd buildstep
git pull origin master
sudo make build
```

This will build a fresh Ubuntu Quantal image, install a number of packages, and
eventually replace the Docker image for buildstep.

## Migrating from 0.1.0 to 0.2.0

This should be summary of breaking changes between 0.1 and 0.2 version with instructions to upgrade.

### software-properties-common

software-properties-common is now a dependency for plugins. Before running dokku install script, make sure it is installed.

```shell
> sudo apt-get install software-properties-common
```

### Gitreceive removed

PR [#270](https://github.com/progrium/dokku/pull/270).

Starting with Dokku 0.2.0, [Gitrecieve](https://github.com/progrium/gitreceive) is replaced by a `git` plugin.

Dokku no longer uses the `git` user. Everything is done with the `dokku` user account.

This causes the remote url to change. Instead of pushing to `git@hostname:app`, you should now push to `dokku@hostname:app`.
The url must be modified using the `git remote set-url` command :

    git remote set-url deploy dokku@hostname:app

Where `deploy` is the name of the remote.

Additionally, the repositiories on the server must be migrated to work with dokku 0.2.0

There is an upgrade [script](https://gist.github.com/plietar/7201430), which is meant to automate this migration. To use it:

1. run the script as root
2. `git pull` to get the latest version of dokku
3. make install
4. `dokku ps:restartall`

TDB.

### Cache directory

Commit [#6350f373](https://github.com/progrium/dokku/commit/6350f373be2cef4f3bb90912099e1be6196522d1)

Buildstep have to be re-installed in order to support cache directory.

TDB.
