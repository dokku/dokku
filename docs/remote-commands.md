# Remote Commands

Dokku commands can be run over ssh. Anywhere you would run `dokku <command>`, just run `ssh -t dokku@dokku.me <command>`
The `-t` is used to request a pty. It is highly recommended to do so.
To avoid the need to type the `-t` option each time, simply create/modify a section in the `.ssh/config` on the client side, as follows:

```
Host dokku.me
RequestTTY yes
```

## Behavioral modifiers

Dokku also supports certain command-line arguments that augment it's behavior. If using these over ssh, you must use the form `ssh -t dokku@dokku.me -- <command>`
in order to avoid ssh interpretting dokku arguments for itself.

```shell
--quiet                suppress output headers
--trace                enable DOKKU_TRACE for current execution only
--rm|--rm-container    remove docker container after successful dokku run <app> <command>
--force                force flag. currently used in apps:destroy
```

## Clients

You may optionally use a client to connect to your dokku server. Most clients use the configured `git remote` to locate the dokku server, though some allow for overriding this via an environment variable or flag.

Please refer to the [clients](/docs/community/clients.md) list for more details.
