# Configuration management

Typically an application will require some configuration to run properly. Dokku supports application configuration via environment variables. Environment variables may contain private data, such as passwords or API keys, so it is not recommended to store them in your application's repository.

The `config` plugin provides the following commands to manage your variables:

```
config <app> - display the config vars for an app
config:get <app> KEY - display a config value for an app
config:set <app> KEY1=VALUE1 [KEY2=VALUE2 ...] - set one or more config vars
config:unset <app> KEY1 [KEY2 ...] - unset one or more config vars
```

The variables are by default available only at run time, i.e. not when the application is being built/compiled. However there are [plugins](http://progrium.viewdocs.io/dokku/plugins) that can change that.
