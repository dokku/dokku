# Plugin Management

> New as of 0.4.0

```
plugin:disable <name>                    # Disable an installed plugin (third-party only)
plugin:enable <name>                     # Enable a previously disabled plugin
plugin:install [--core|git-url [--committish tag|branch|commit|--name custom-plugin-name]]           # Optionally download git-url (with custom tag/committish) & run install trigger for active plugins (or only core ones)
plugin:installed <name>                  # Checks if a plugin is installed
plugin:install-dependencies [--core]     # Run install-dependencies trigger for active plugins (or only core ones)
plugin:list                              # Print active plugins
plugin:trigger <args...>.                # Trigger an arbitrary plugin hook
plugin:uninstall <name>                  # Uninstall a plugin (third-party only)
plugin:update [name [committish]]        # Optionally update named plugin from git (with custom tag/committish) & run update trigger for active plugins
```

```shell
# for 0.3.x
cd /var/lib/dokku/plugins
git clone <git url>
dokku plugins-install
```

> These commands require `root` permissions as the `install` and `install-dependencies` plugin triggers may utilize commands such as `apt-get`. For non-core plugins, please inspect those plugins before running the following command as `root` user.

## Usage

### Listing Plugins

Installed plugins can be listed via the `plugin:list` command:

```shell
dokku plugin:list
```

```
plugn: dev
  00_dokku-standard    0.26.8 enabled    dokku core standard plugin
  20_events            0.26.8 enabled    dokku core events logging plugin
  app-json             0.26.8 enabled    dokku core app-json plugin
  apps                 0.26.8 enabled    dokku core apps plugin
  build-env            0.26.8 enabled    dokku core build-env plugin
  buildpacks           0.26.8 enabled    dokku core buildpacks plugin
  certs                0.26.8 enabled    dokku core certificate management plugin
  checks               0.26.8 enabled    dokku core checks plugin
  common               0.26.8 enabled    dokku core common plugin
  config               0.26.8 enabled    dokku core config plugin
  docker-options       0.26.8 enabled    dokku core docker-options plugin
  domains              0.26.8 enabled    dokku core domains plugin
  enter                0.26.8 enabled    dokku core enter plugin
  git                  0.26.8 enabled    dokku core git plugin
  logs                 0.26.8 enabled    dokku core logs plugin
  network              0.26.8 enabled    dokku core network plugin
  nginx-vhosts         0.26.8 enabled    dokku core nginx-vhosts plugin
  plugin               0.26.8 enabled    dokku core plugin plugin
  proxy                0.26.8 enabled    dokku core proxy plugin
  ps                   0.26.8 enabled    dokku core ps plugin
  repo                 0.26.8 enabled    dokku core repo plugin
  resource             0.26.8 enabled    dokku core resource plugin
  scheduler-docker-local 0.26.8 enabled    dokku core scheduler-docker-local plugin
  shell                0.26.8 enabled    dokku core shell plugin
  ssh-keys             0.26.8 enabled    dokku core ssh-keys plugin
  storage              0.26.8 enabled    dokku core storage plugin
  tags                 0.26.8 enabled    dokku core tags plugin
  tar                  0.26.8 enabled    dokku core tar plugin
  trace                0.26.8 enabled    dokku core trace plugin
```

> Warning: All plugin commands other than `plugin:list` and `plugin:help` require sudo access and must be run directly from the Dokku server.

### Checking if a plugin is installed

You can check if a plugin has been installed via the `plugin:installed` command:

```shell
dokku plugin:installed postgres
```

### Installing a plugin

Installing a plugin is easy as well using the `plugin:install` command. This command will also trigger the `install` pluginhook on all existing plugins.

The most common usage is to install a plugin from a url. This url may be any of the following:

- `git`: For git+ssh based plugin repository clones.
- `ssh`: For git+ssh based plugin repository clones.
- `file`: For copying plugins from a path on disk.
- `https`: For http based plugin repository clones.

Additionally, any urls with the extensions `.tar.gz` or `.tgz` are treated as Gzipped Tarballs for installation purposes and will be downloaded and extracted into place.

```shell
dokku plugin:install https://github.com/dokku/dokku-postgres.git
```

```
-----> Cloning plugin repo https://github.com/dokku/dokku-postgres.git to /var/lib/dokku/plugins/available/postgres
Cloning into 'postgres'...
remote: Counting objects: 646, done.
remote: Total 646 (delta 0), reused 0 (delta 0), pack-reused 646
Receiving objects: 100% (646/646), 134.24 KiB | 0 bytes/s, done.
Resolving deltas: 100% (406/406), done.
Checking connectivity... done.
-----> Plugin postgres enabled
```

For git-based plugin installation, a commit SHA-like object may be specified (tag/branch/commit sha) via the `--committish` argument and Dokku will attempt to install the specified commit object.

```shell
# where 2.0.0 is a potential git tag
dokku plugin:install https://github.com/dokku/dokku-postgres.git --committish 2.0.0
```

Plugin names are interpolated based on the repository name minus the `dokku-` prefix. If the plugin being installed has a name other than what matches the repository name - or another name is desired - the `--name` flag can be used to override this interpolation.

```shell
dokku plugin:install https://github.com/dokku/smoke-test-plugin.git --name smoke-test-plugin
```

The `--core` flag may also be indicated as the sole argument, though it is only for installation of core plugins, and thus not useful for end-user installations.

```shell
dokku plugin:install --core
```

Finally, all flags may be omitted to trigger the `install` procedures for both core and third-party plugins:

```shell
dokku plugin:install
```

### Installing plugin dependencies

In some cases, plugins will have system-level dependencies. These are not automatically installed via `plugin:install`, and must be separately via the `plugin:install-dependencies` command. This will run through all the `dependencies` trigger for all plugins.

```shell
dokku plugin:install-dependencies
```

This command may also target _just_ core plugins via the `--core` flag. This is usually only useful for source-based installs of Dokku.

```shell
dokku plugin:install-dependencies --core
```

### Updating a plugin

An installed, third-party plugin can be updated can updated via the `plugin:update` command. This should be done after any upgrades of Dokku as there may be changes in the internal api that require an update of how the plugin interfaces with Dokku.

Please note that this command is only valid for plugin installs that were backed by a git-repository.

```shell
dokku plugin:update postgres
```

```
Plugin (postgres) updated
```

An optional commit SHA-like object may be specified.

```shell
dokku plugin:update postgres 2.0.0
```

### Uninstalling a plugin

Third party plugins can be uninstalled using the `plugin:uninstall` command:

```shell
dokku plugin:uninstall postgres
```

```
-----> Plugin postgres uninstalled
```

### Disabling a plugin

Disabling a plugin can also be useful for debugging whether a third-party plugin is causing issues in a Dokku installation. Another common use case is for disabling core functionality for replacement with a third-party plugin.

```shell
dokku plugin:disable postgres
```

```
-----> Plugin postgres disabled
```

### Enabling a plugin

Disabled plugins can be re-enabled via the `plugin:enable` command.

```shell
dokku plugin:enable postgres
```

```
-----> Plugin postgres enabled
```

### Triggering a plugin trigger

The `plugin:trigger` can be used to call any internal plugin trigger. This may have unintended consequences, and thus should only be called for development or debugging purposes.

```shell
dokku plugin:trigger some-internal-trigger args-go-here
```
