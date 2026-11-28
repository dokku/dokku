# Environment Variable Groups

> [!IMPORTANT]
> New as of 0.39.0

```
envgroup:add [--index <index>] [--restart] <group> <app>                    # Adds an app to an env group
envgroup:bundle <group>                                                     # Bundles an env group's environment into a tarfile
envgroup:clear [--restart] <group>                                          # Clears an env group's environment variables
envgroup:create <group>                                                     # Creates an env group
envgroup:destroy [--force] [--restart] <group>                              # Destroys an env group
envgroup:exists <group>                                                     # Checks if an env group exists
envgroup:export [--format <format>] <group>                                 # Exports an env group's environment
envgroup:get [--quoted] <group> KEY                                         # Displays an env group config value
envgroup:import [--format <format>] [--replace] [--restart] <group> [FILE|-] # Imports an env group's environment from a file
envgroup:info <group>                                                       # Displays the apps in an env group and its number of keys
envgroup:keys <group>                                                       # Shows the keys set in an env group
envgroup:list                                                               # Lists all env groups
envgroup:remove [--restart] <group> <app>                                   # Removes an app from an env group
envgroup:report [<app>] [<flag>]                                            # Displays an envgroup report for one or more apps
envgroup:set [--encoded] [--restart] <group> KEY1=VALUE1 [KEY2=VALUE2 ...]  # Sets one or more env group config vars
envgroup:show <group>                                                       # Pretty-prints an env group's environment
envgroup:unset [--restart] <group> KEY1 [KEY2 ...]                          # Unsets one or more env group config vars
```

An env group is a named set of [environment variables](/docs/configuration/environment-variables.md) that is shared by every app added to it. This is useful when several apps need the same values - for example test credentials shared by a set of development environments - without having to set them on each app individually.

## Precedence

An app's environment is built from the following layers, with each layer overriding the one before it:

1. The global environment, set with `config:set --global`.
2. The env groups the app belongs to, in the order they were added. A group added later overrides a group added earlier.
3. The app's own environment, set with `config:set <app>`.

Env group values are included wherever the merged environment of an app is used: at runtime, during builds, and in the output of `config:show`, `config:keys`, `config:export` and `config:bundle` when the `--merged` flag is passed. They are never written into the app's own environment, so `config:get <app>` and the commands above without `--merged` only show the app's own values.

## Usage

### Creating and destroying groups

Groups must be created before they can be used. Group names follow the same rules as app names: they must begin with a lowercase alphanumeric character, and may only contain lowercase alphanumerics, dots, and hyphens.

```shell
dokku envgroup:create dev
```

```
-----> Created env group dev
```

A group can be checked for existence with `envgroup:exists`, which exits non-zero when the group does not exist:

```shell
dokku envgroup:exists dev
```

All groups can be listed with `envgroup:list`:

```shell
dokku envgroup:list
```

```
=====> Env groups
dev
staging
```

A group is destroyed with `envgroup:destroy`. A group that apps still belong to is not destroyed unless the `--force` flag is passed, in which case the group is first removed from each of those apps.

```shell
dokku envgroup:destroy --force dev
```

### Managing group variables

The commands for managing the variables in a group mirror their `config:*` counterparts, taking a group name in place of an app name:

```shell
dokku envgroup:set dev API_URL=https://api.example.com TEST_TOKEN=secret
dokku envgroup:set --encoded dev KEY="$(base64 -w 0 ~/.ssh/id_rsa)"
dokku envgroup:get dev API_URL
dokku envgroup:unset dev TEST_TOKEN
dokku envgroup:keys dev
dokku envgroup:show dev
dokku envgroup:export --format envfile dev
dokku envgroup:bundle dev > dev.tar
dokku envgroup:import dev ./dev.env
dokku envgroup:import --format json --replace dev - < dev.json
dokku envgroup:clear dev
```

`envgroup:export` supports the same formats as `config:export`, and `envgroup:import` supports the same `envfile` and `json` formats as `config:import`. With `--replace`, `envgroup:import` unsets every key not in the file, and leaves the group and its apps untouched when the group already holds exactly the imported keys.

### Adding apps to groups

An app is added to a group with `envgroup:add`, and removed from it with `envgroup:remove`. An app may belong to any number of groups.

```shell
dokku envgroup:add dev node-js-app
dokku envgroup:add shared node-js-app
dokku envgroup:remove shared node-js-app
```

Groups are appended to the end of the app's group list by default, giving them the highest precedence among the app's groups. The `--index` flag inserts the group at a 1-based position in the list instead:

```shell
# place the shared group first, so dev overrides it
dokku envgroup:add --index 1 shared node-js-app
```

The apps in a group, along with the number of keys it holds, can be viewed with `envgroup:info`:

```shell
dokku envgroup:info dev
```

```
=====> dev env group information
       Env group apps:  node-js-app python-app
       Env group keys:  2
```

### Applying changes

Changing a group's variables, adding an app to or removing an app from a group, or destroying a group in use by apps does not redeploy any app. Instead, the deployed apps affected by the change are listed, as they must be redeployed for the change to take effect:

```shell
dokku envgroup:set dev API_URL=https://api2.example.com
```

```
-----> Setting config vars
       API_URL:  https://api2.example.com
 !     The following apps must be redeployed for the change to take effect: node-js-app, python-app
 !     Pass --restart to redeploy them, or run ps:rebuild on each app
```

Pass the `--restart` flag to redeploy those apps as part of the command, in the same way that `config:set` redeploys an app:

```shell
dokku envgroup:set --restart dev API_URL=https://api2.example.com
```

> [!NOTE]
> `ps:restart` reuses the app's existing release, and apps built with herokuish keep the global and group values from that release in `/app/.profile.d`. Use `--restart` or `ps:rebuild` so that those apps pick up the new values.
>
> Apps built with herokuish also keep the merged environment from their last build in `/app/.env`. Keys removed from a group, or from an app by removing the group, remain set there until the app is rebuilt with `ps:rebuild`, as with `config:unset`.

### Backups

Env groups and their variables are included in the global scope of a [backup](/docs/advanced-usage/backup-recovery.md), and each app's group memberships are included in the app's scope. When an app's memberships are restored, any group that does not exist on the host is skipped with a warning.

### Displaying reports for an app

You can get a report about the env groups an app belongs to using the `envgroup:report` command:

```shell
dokku envgroup:report
```

```
=====> node-js-app envgroup information
       Envgroup groups:               shared dev
=====> python-app envgroup information
       Envgroup groups:               dev
```

You can run the command for a specific app also.

```shell
dokku envgroup:report node-js-app
```

```
=====> node-js-app envgroup information
       Envgroup groups:               shared dev
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku envgroup:report node-js-app --envgroup-groups
```

```
shared dev
```
