# Environment Variables

Typically an application will require some configuration to run properly. Dokku supports application configuration via environment variables. Environment variables may contain private data, such as passwords or API keys, so it is not recommended to store them in your application's repository.

The `config` plugin provides the following commands to manage your variables:

```
config (<app>|--global)                                   Display all global or app-specific config vars
config:get (<app>|--global) KEY                           Display a global or app-specific config value
config:set (<app>|--global) KEY1=VALUE1 [KEY2=VALUE2 ...] Set one or more config vars
config:unset (<app>|--global) KEY1 [KEY2 ...]             Unset one or more config vars
```

The variables are available both at run time and during the application build/compilation step.

> Note: Global `ENV` files are sourced before app-specific `ENV` files. This means that app-specific variables will take precedence over global variables. Configuring your global `ENV` file is manual, and should be considered potentially dangerous as configuration applies to all applications.

You can set multiple environment variables at once:

```shell
dokku config:set node-js-app ENV=prod COMPILE_ASSETS=1
```

When setting variables with whitespaces, you need to escape them:

```shell
dokku config:set node-js-app KEY=\"VAL\ WITH\ SPACES\"
```

When setting or unsetting environment variables, you may wish to avoid an application restart. This is useful when developing plugins or when setting multiple environment variables in a scripted manner. To do so, use the `--no-restart` flag:

```shell
dokku --no-restart config:set node-js-app ENV=prod
```

If you wish to have the variables output in an `eval`-compatible form, you can use the `--export` flag:

```shell
dokku config node-js-app --export
# outputs variables in the form:
#
#   export ENV='prod'
#   export COMPILE_ASSETS='1'

# source in all the node-js-app app environment variables
eval $(dokku config node-js-app --export)
```

You can also output the variables in a single-line for usage in command-line utilities with the `--shell` flag:

```shell
dokku config node-js-app --shell

# outputs variables in the form:
#
#   ENV='prod' COMPILE_ASSETS='1'
```
