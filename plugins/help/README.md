# Dokku core help command

This plugin provides the `help` command, which outputs the usage of every
command provided by a Dokku plugin, as specified by the plugin's `help.txt`
and the output of the `help` command as by other plugins.

## help.txt format

Each line is interpreted to describe the usage, and human explanation, for a
command. Usage and explanation are separated by the first instance of 2 or more
spaces.

Lines starting with ":" will have the ":" replaced with the name of the plugin,
followed by a colon if there is a subcommand name immediately following the
colon.

For example, this `help.txt` for the `config` plugin:

```
: <app>  display the config vars for an app
:get <app> KEY  display a config value for an app
:set <app> KEY1=VALUE1 [KEY2=VALUE2 ...]  set one or more config vars
:unset <app> KEY1 [KEY2 ...]  unset one or more config vars
```

would be formatted as this when output by the `help` command:
```
    config <app>                                    display the config vars for an app
    config:get <app> KEY                            display a config value for an app
    config:set <app> KEY1=VALUE1 [KEY2=VALUE2 ...]  set one or more config vars
    config:unset <app> KEY1 [KEY2 ...]              unset one or more config vars
```

