# Remote commands

Dokku commands can be run over ssh. Anywhere you would run `dokku <command>`, just run `ssh -t dokku@dokku.me <command>`
The `-t` is used to request a pty. It is highly recommended to do so.
To avoid the need to type the `-t` option each time, simply create/modify a section in the `.ssh/config` on the client side, as follows:

```
Host dokku.me
RequestTTY yes
```

## Run a command in the app environment

It's possible to run commands in the environment of the deployed application:

```bash
dokku run node-js-app ls -alh
dokku run <app> <cmd>
```
