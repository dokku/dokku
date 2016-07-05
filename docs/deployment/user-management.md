# User Management

While it is possible to use password-based authorization to push to Dokku, it is preferable to use key-based authentication for security.

Users in dokku are managed via the `~/dokku/.ssh/authorized_keys` file. While you *can* manually edit this file, it is **highly** recommended that you follow the below steps to manage users on a dokku server.

## SSHCommand

Dokku uses the [`sshcommand`](https://github.com/dokku/sshcommand) utility to manage ssh keys for the dokku user. The following is the usage output for sshcommand.

```
sshcommand create <user> <command>             # creates a user forced to run command when SSH connects
sshcommand acl-add <user> <ssh-key-name>       # adds named SSH key to user from STDIN
sshcommand acl-remove <user> <ssh-key-name>    # removes SSH key by name
sshcommand help                                # displays the usage help message
```

In dokku's case, the `<user>` section is *always* `dokku`, as this is the system user that the dokku binary performs all it's actions. Keys are given unique names, which can be used in conjunction with the [user-auth](/dokku/development/plugin-triggers/#user-auth) plugin trigger to handle command authorization.

## Adding deploy users

You can add your public key to the dokku user's `~/dokku/.ssh/authorized_keys` file with the following command:

```shell
# from your local machine
# replace dokku.me with your domain name or the host's IP
# replace root with your server's root user
# USER is the username you use to refer to this particular key
cat ~/.ssh/id_rsa.pub | ssh root@dokku.me "sudo sshcommand acl-add dokku USER"
```

At it's base, the `sshcommand` *must* be run under a user with sudo access, as it sets keys for the dokku user.

For instance, if you stored your public key at `~/.ssh/id_rsa.pub-open` and are deploying to EC2 where the default root-enabled user is `ubuntu`, you can run the following command to add your key under the `superuser` username:

```shell
cat ~/.ssh/id_rsa.pub-open | ssh ubuntu@dokku.me "sudo sshcommand acl-add dokku superuser"
```

If you are using the vagrant installation, you can also use the `make vagrant-acl-add` target to add your public key to dokku (it will use your host username as the `USER`):

```shell
cat ~/.ssh/id_rsa.pub | make vagrant-acl-add
```

## Scoping commands to specific users

See the [user auth plugin trigger documentation](/dokku/development/plugin-triggers/#user-auth).
