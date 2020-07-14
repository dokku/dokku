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

You can list all installed plugins using the `plugin:list` command:

```shell
dokku plugin:list
```

```
plugn: dev
  00_dokku-standard    0.21.1 enabled    dokku core standard plugin
  20_events            0.21.1 enabled    dokku core events logging plugin
  app-json             0.21.1 enabled    dokku core app-json plugin
  apps                 0.21.1 enabled    dokku core apps plugin
  build-env            0.21.1 enabled    dokku core build-env plugin
  buildpacks           0.21.1 enabled    dokku core buildpacks plugin
  certs                0.21.1 enabled    dokku core certificate management plugin
  checks               0.21.1 enabled    dokku core checks plugin
  common               0.21.1 enabled    dokku core common plugin
  config               0.21.1 enabled    dokku core config plugin
  docker-options       0.21.1 enabled    dokku core docker-options plugin
  domains              0.21.1 enabled    dokku core domains plugin
  enter                0.21.1 enabled    dokku core enter plugin
  git                  0.21.1 enabled    dokku core git plugin
  logs                 0.21.1 enabled    dokku core logs plugin
  network              0.21.1 enabled    dokku core network plugin
  nginx-vhosts         0.21.1 enabled    dokku core nginx-vhosts plugin
  plugin               0.21.1 enabled    dokku core plugin plugin
  proxy                0.21.1 enabled    dokku core proxy plugin
  ps                   0.21.1 enabled    dokku core ps plugin
  repo                 0.21.1 enabled    dokku core repo plugin
  resource             0.21.1 enabled    dokku core resource plugin
  scheduler-docker-local 0.21.1 enabled    dokku core scheduler-docker-local plugin
  shell                0.21.1 enabled    dokku core shell plugin
  ssh-keys             0.21.1 enabled    dokku core ssh-keys plugin
  storage              0.21.1 enabled    dokku core storage plugin
  tags                 0.21.1 enabled    dokku core tags plugin
  tar                  0.21.1 enabled    dokku core tar plugin
  trace                0.21.1 enabled    dokku core trace plugin
```

You can check if a plugin has been installed via the `plugin:installed` command:

```shell
dokku plugin:installed postgres
```

Installing a plugin is easy as well using the `plugin:install` command. This command will also trigger the `install` pluginhook on all existing plugins.

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

You can also uninstall a third-party plugin using the `plugin:uninstall` command:

```shell
dokku plugin:uninstall postgres
```

```
-----> Plugin postgres uninstalled
```

Enabling or disabling a plugin can also be useful in cases where you are debugging whether a third-party plugin is causing issues in your Dokku installation:

```shell
dokku plugin:disable postgres
```

```
-----> Plugin postgres disabled
```

```shell
dokku plugin:enable postgres
```

```
-----> Plugin postgres enabled
```

Finally, you can update an installed third-party plugin. This should be done after any upgrades of Dokku as there may be changes in the internal api that require an update of how the plugin interfaces with Dokku.

```shell
dokku plugin:update postgres
```

```
Plugin (postgres) updated
```
