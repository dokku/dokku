# Environment Variables

Typically an application will require some configuration to run properly. Dokku supports application configuration via environment variables. Environment variables may contain private data, such as passwords or API keys, so it is not recommended to store them in your application's repository.

The `config` plugin provides the following commands to manage your variables:

```
config:show (<app>|--global)                                                          Pretty-print an app or global environment
config:bundle (<app>|--global) [--merged]                                             Bundle environment into tarfile
config:clear (<app>|--global)                                                         Clears environment variables
config:export (<app>|--global) [--format <format>]                                    Export a global or app environment
config:get (<app>|--global) KEY                                                       Display a global or app-specific config value
config:keys (<app>|--global) [--merged]                                               Show keys set in environment
config:set [--encoded] [--no-restart] (<app>|--global) KEY1=VALUE1 [KEY2=VALUE2 ...]  Set one or more config vars
config:unset [--no-restart] (<app>|--global) KEY1 [KEY2 ...]                          Unset one or more config vars
```

> For security reasons - and as per [docker recommendations](https://github.com/docker/docker/issues/13490) - Dockerfile-based deploys have variables available _only_ during runtime, as noted in [this issue](https://github.com/dokku/dokku/issues/1860). Consider using [build arguments](/docs/deployment/builders/dockerfiles.md#build-time-configuration-variables) to expose variables during build-time for Dockerfile apps.

Environment variables are available both at run time and during the application build/compilation step for buildpack-based deploys.

For buildpack deploys, Dokku will create a  `/app/.env` file that can be used for legacy buildpacks. Note that this is _not_ updated when `config:set` or `config:unset` is called, and is only written during a `deploy` or `ps:rebuild`. Developers are encouraged to instead read from the application environment directly, as the proper values will be available then.

> [!NOTE]
> Global environment variables are sourced before app-specific environment variables. This means that app-specific variables will take precedence over global variables. Configuring global environment variables should be considered potentially dangerous as configuration applies to all applications.

You can set multiple environment variables at once:

```shell
dokku config:set node-js-app APP_ENV=prod COMPILE_ASSETS=1
```

Whitespace and special characters get tricky. If you are using dokku locally you don't need to do any special escaping. If you are using dokku over ssh you will need to backslash-escape spaces:

```shell
dokku config:set node-js-app KEY="VAL\ WITH\ SPACES"
```

Dokku can also read base64 encoded values. That's the easiest way to set a value with newlines or spaces. To set a value with newlines you need to base64 encode it first and pass the `--encoded` flag:

```shell
dokku config:set --encoded node-js-app KEY="$(base64 -w 0 ~/.ssh/id_rsa)"
```

When setting or unsetting environment variables, you may wish to avoid an application restart. This is useful when developing plugins or when setting multiple environment variables in a scripted manner. To do so, use the `--no-restart` flag:

```shell
dokku config:set --no-restart node-js-app APP_ENV=prod
```

If you wish to have the variables output in an `eval`-compatible form, you can use the `config:export` command

```shell
dokku config:export node-js-app
# outputs variables in the form:
#
#   export APP_ENV='prod'
#   export COMPILE_ASSETS='1'

# source in all the node-js-app app environment variables
eval $(dokku config:export node-js-app)
```

You can control the format of the exported variables with the `--format` flag. `--format=shell` will output the variables in a single-line for usage in command-line utilities:

```shell
dokku config:export --format shell node-js-app

# outputs variables in the form:
#
#   APP_ENV='prod' COMPILE_ASSETS='1'
```

## Setting Environment Variables via app.json

Environment variables can also be declared in an `app.json` file in your repository root. This is useful for setting default values, generating secrets, or requiring certain variables to be set before deployment.

```json
{
  "env": {
    "SIMPLE_VAR": "default_value",
    "SECRET_KEY": {
      "description": "A secret key for signing tokens",
      "generator": "secret"
    },
    "DATABASE_URL": {
      "description": "PostgreSQL connection string",
      "required": true
    }
  }
}
```

Environment variables from `app.json` are processed during the first deploy, before the predeploy script runs. Variables can be configured as:

- **Simple string values**: Set as defaults if not already configured
- **Generated secrets**: Use `"generator": "secret"` to auto-generate a 64-character hex string
- **Required variables**: Use `"required": true` to prompt for or require a value
- **Synced variables**: Use `"sync": true` to update the value on every deploy

For full details on the `env` schema and behavior, see the [app.json documentation](/docs/appendices/file-formats/app-json.md#env).

## Special Config Variables

The following config variables have special meanings and can be set in a variety of ways. Unless specified via global app config, the values may not be passed into applications. Usage of these values within applications should be considered unsafe, as they are an internal configuration values that may be moved to the internal properties system in the future.

> [!WARNING]
> This list is not exhaustive, and may vary from version to version.

| Name                           | Default                         | How to modify                                                                                                                                    | Description                                                                                                |
| ------------------------------ | ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------- |
| `DOKKU_ROOT`                   | `~dokku`                        | `/etc/environment`                                                                                                                               | The root directory where dokku will store application repositories, as well as certain configuration files. |
| `DOKKU_IMAGE`                  | `gliderlabs/herokuish`          | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | The default image to use when building herokuish containers. Deprecated in favor of using `buildpacks:set-property` |
| `DOKKU_LIB_ROOT`               | `/var/lib/dokku`                | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | The directory where plugins, certain data, and general configuration is stored. |
| `PLUGIN_PATH`                  | `$DOKKU_LIB_ROOT/plugins"`      | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | The top-level directory where plugins are stored. |
| `PLUGIN_AVAILABLE_PATH`        | `$PLUGIN_PATH/available"`       | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | The directory that holds all available plugins, including core. |
| `PLUGIN_ENABLED_PATH`          | `$PLUGIN_PATH/enabled"`         | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | The directory that holds all enabled plugins, including core. |
| `PLUGIN_CORE_PATH`             | `$DOKKU_LIB_ROOT/core-plugins"` | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | The directory that stores all core plugins. |
| `PLUGIN_CORE_AVAILABLE_PATH`   | `$PLUGIN_CORE_PATH/available"`  | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | The directory that stores all available core plugins. |
| `PLUGIN_CORE_ENABLED_PATH`     | `$PLUGIN_CORE_PATH/enabled"`    | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | The directory that stores all enabled core plugins. |
| `DOKKU_LOGS_DIR`               | `/var/log/dokku`                | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | Where dokku logs should be written to. |
| `DOKKU_LOGS_HOST_DIR`          | `$DOKKU_LOGS_DIR`               | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | A path on the host that will be mounted into the vector logging container. |
| `DOKKU_EVENTS_LOGFILE`         | `$DOKKU_LOGS_DIR/events.log`    | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | Where the events log file is written to. |
| `DOKKU_APP_NAME`               | none                            | `--app APP` flag                                                                                                                                 | Name of application to work on. Respected by core plugins. |
| `DOKKU_APPS_FORCE_DELETE`      | none                            | `--force` flag                                                                                                                                   | Whether to force delete an application. Also used by other plugins for destructive actions. |
| `DOKKU_CHECKS_URL`             | `https://dokku.com/docs/deployment/zero-downtime-deploys/` | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                | Url displayed during deployment when no CHECKS file exists. |
| `DOKKU_QUIET_OUTPUT`           | none                            | `--quiet` flag                                                                                                                                   | Silences certain header output for `dokku` commands. |
| `DOKKU_RM_CONTAINER`           | none                            | `dokku config:set` <br />                                                                                                                        | Deprecated: Whether to keep `dokku run` containers around or not. |
| `DOKKU_TRACE`                  | none                            | `dokku trace:on`   <br /> `dokku trace:off` <br /> `--trace` flag                                                                                | Turn on very verbose debugging. |
| `DOKKU_SKIP_CLEANUP`           |                                 | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | When a deploy is triggered, if this is set to a non-empty value, then old docker containers and images will not be removed. Falls back to the `builder:set skip-cleanup` property. |
| `DOKKU_SYSTEM_GROUP`           | `dokku`                         | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | System group to chown files as. |
| `DOKKU_SYSTEM_USER`            | `dokku`                         | `/etc/environment` <br /> `~dokku/.dokkurc` <br /> `~dokku/.dokkurc/*`                                                                           | System user to chown files as. |

## Deprecated Environment Variables

The following environment variables have been migrated to plugin properties. Existing values will be automatically migrated on upgrade. Use the corresponding property commands going forward.

| Deprecated Env Var | Replacement Command |
|---|---|
| `DOKKU_APP_PROXY_TYPE` | `dokku proxy:set <app> type <value>` |
| `DOKKU_APP_RESTORE` | `dokku ps:set <app> restore <true\|false>` |
| `DOKKU_APP_SHELL` | `dokku scheduler:set <app> shell <value>` |
| `DOKKU_CHECKS_DISABLED` | `dokku checks:disable <app> [proctypes]` |
| `DOKKU_CHECKS_ENABLED` | `dokku checks:enable <app> [proctypes]` |
| `DOKKU_CHECKS_SKIPPED` | `dokku checks:skip <app> [proctypes]` |
| `DOKKU_CHECKS_WAIT` | `dokku checks:set <app> wait <value>` |
| `DOKKU_CHECKS_TIMEOUT` | `dokku checks:set <app> timeout <value>` |
| `DOKKU_CHECKS_ATTEMPTS` | `dokku checks:set <app> attempts <value>` |
| `DOKKU_DEFAULT_CHECKS_WAIT` | `dokku checks:set --global default-wait <value>` |
| `DOKKU_DISABLE_APP_AUTOCREATION` | `dokku apps:set --global disable-autocreation <true\|false>` |
| `DOKKU_DISABLE_PROXY` | `dokku proxy:disable <app>` / `dokku proxy:enable <app>` |
| `DOKKU_DOCKERFILE_START_CMD` | `dokku ps:set <app> dockerfile-start-cmd <value>` |
| `DOKKU_PARALLEL_ARGUMENTS` | Removed. No longer supported. |
| `DOKKU_PROXY_PORT` | `dokku proxy:set <app> proxy-port <value>` |
| `DOKKU_PROXY_SSL_PORT` | `dokku proxy:set <app> proxy-ssl-port <value>` |
| `DOKKU_SKIP_ALL_CHECKS` | `dokku checks:disable <app>` |
| `DOKKU_SKIP_CLEANUP` | `dokku builder:set <app> skip-cleanup <true\|false>` |
| `DOKKU_SKIP_DEFAULT_CHECKS` | `dokku checks:skip <app>` |
| `DOKKU_SKIP_DEPLOY` | `dokku ps:set <app> skip-deploy <true\|false>` |
| `DOKKU_START_CMD` | `dokku ps:set <app> start-cmd <value>` |
