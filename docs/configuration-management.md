# Configuration management

Typically an application will require some configuration to run properly. Dokku supports application configuration via environment variables. Environment variables may contain private data, such as passwords or API keys, so it is not recommended to store them in your application's repository.

The `config` plugin provides the following commands to manage your variables:

```
config <app> - display the config vars for an app
config:get <app> KEY - display a config value for an app
config:set <app> KEY1=VALUE1 [KEY2=VALUE2 ...] - set one or more config vars
config:unset <app> KEY1 [KEY2 ...] - unset one or more config vars
```

The variables are available both at run time and during the application build/compilation step. You no longer need a `user-env` plugin as Dokku handles this functionality in a way equivalent to how Heroku handles it.

> Note: Global `BUILD_ENV` files are currently migrated into a global `ENV` file and sourced before app-specific variables. This means that app-specific variables will take precedence over global variables. Configuring your global `ENV` file is manual, and should be considered potentially dangerous as configuration applies to all applications.
