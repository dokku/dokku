# User Management

> New as of 0.7.0

```
ssh-keys:add <name> [/path/to/key]             # Add a new public key by pipe or path
ssh-keys:list                                  # List of all authorized Dokku public ssh keys
ssh-keys                                       # Manage public ssh keys that are allowed to connect to Dokku
ssh-keys:remove <name>                         # Remove SSH public key by name
```

When pushing to Dokku, ssh key based authorization is the preferred authentication method, for ease of use and increased security.

Users in Dokku are managed via the `~/dokku/.ssh/authorized_keys` file. It is **highly** recommended that you follow the  steps below to manage users on a Dokku server.

> Users of older versions of Dokku should use the `sshcommand` binary to manage keys. Please refer to the Dokku documentation for your version for more details.

## Adding deploy users

You can add your public key to Dokku with the following command:

```shell
dokku ssh-keys:add KEY_NAME path/to/id_rsa.pub
```

`KEY_NAME` is the username prefer to use to refer to this particular key. Including the word `admin` in the name will grant the user privileges to add additional keys remotely.

Key names are unique, and attempting to re-use a name will result in an error.

> The unique `KEY_NAME` is for ease of identifying the keys. The ssh (git) user is *always* `dokku`, as this is the system user that the `dokku` binary uses to perform all it's actions.

As key names are unique, they can be used to remove a public ssh key:

```SHELL
dokku ssh-keys:remove KEY_NAME
```

Admin users and root can also add keys remotely:

```shell
cat ~/.ssh/id_rsa.pub | ssh dokku@dokku.me ssh-keys:add KEY_NAME
```

Finally, if you are using the vagrant installation, you can also use the `make vagrant-acl-add` target to add your public key to Dokku (it will use your host username as the `USER`):

```shell
cat ~/.ssh/id_rsa.pub | make vagrant-acl-add
```

## Scoping commands to specific users

Keys are given unique names, which can be used in conjunction with the [user-auth](/dokku/development/plugin-triggers/#user-auth) plugin trigger to handle command authorization. Please see the documentation on that trigger for more information.
