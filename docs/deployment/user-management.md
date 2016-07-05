# User Management

When pushing to Dokku, ssh key based authorization is the preferred authentication method, for ease of use and increased security.

Users in Dokku are managed via the `~/dokku/.ssh/authorized_keys` file. It is **highly** recommended that you follow the  steps below to manage users on a Dokku server.

## Dokku ssh-keys command

The `dokku ssh-keys` command(s) allow you to manage ssh keys used to push to the Dokku server. The following is the usage output for `dokku ssh-keys`:

```
$ dokku ssh-keys:help
Usage: dokku ssh-keys[:COMMAND]

Manage public ssh keys that are allowed to connect to Dokku

Additional commands:
    ssh-keys:add <name> [/path/to/key]   Add a new public key by pipe or path
    ssh-keys:list                        List of all authorized Dokku public ssh keys
    ssh-keys                             Manage public ssh keys that are allowed to connect to Dokku
    ssh-keys:remove <name>               Remove SSH public key by name
```

Keys are given unique names, which can be used in conjunction with the [user-auth](/dokku/development/plugin-triggers/#user-auth) plugin trigger to handle command authorization. In Dokku's case, the unique _name_ is just for ease of identifying the keys, the ssh (git) user is *always* `dokku`, as this is the system user that the `dokku` binary uses to perform all it's actions. 

## Adding deploy users

You can add your public key to Dokku with the following command:

`NAME` is the username prefer to use to refer to this particular key. Including the word `admin` in the name will grant the user privileges to add additional keys remotely.

```shell
$ dokku ssh-keys:add <NAME> <PATH/TO/KEY>
```

Admin users and root can also add keys remotely: 
```shell
cat <PATH/TO/KEY> | ssh dokku@dokku.me ssh-keys:add <NAME>
```

If you are using the vagrant installation, you can also use the `make vagrant-acl-add` target to add your public key to dokku (it will use your host username as the `USER`):

```shell
cat ~/.ssh/id_rsa.pub | make vagrant-acl-add
```

## Scoping commands to specific users

See the [user auth plugin trigger documentation](/dokku/development/plugin-triggers/#user-auth).
