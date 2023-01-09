# Remote Commands

Dokku commands can be run over SSH. Anywhere you would run `dokku <command>`, just run `ssh -t dokku@dokku.me <command>`
The `-t` is used to request a pty. It is highly recommended to do so.
To avoid the need to type the `-t` option each time, create/modify a section in the `.ssh/config` on the client side, as follows:

```ini
Host dokku.me
    RequestTTY yes
```

## Behavioral modifiers

Dokku also supports certain command line arguments that augment its behavior. If using these over SSH, you must use the form `ssh -t dokku@dokku.me -- <command>`
to avoid SSH interpreting Dokku arguments for itself.

```
--quiet                suppress output headers
--trace                enable DOKKU_TRACE for current execution only
--force                force flag. currently used in apps:destroy and other ":destroy" commands
```

## Official Client

You may optionally use the official client when connecting to the Dokku server.

Of all methods, this is the *most* official method of interacting with your Dokku installation. It is a bash script that interacts with a remote Dokku installation via `ssh`. It is available in `contrib/dokku_client.sh` in the root of the Dokku repository.

It can be installed either via the [Homebrew](https://brew.sh) package manager (macOS only), or manually.

### Installation via Homebrew

To install, simply run the following command:

```shell
brew install dokku/repo/dokku
```

### Manual installation

To install manually, simply clone the Dokku repository down and add the `dokku` alias pointing at the script:

```shell
git clone git@github.com:dokku/dokku.git ~/.dokku

# optional: make sure that the dokku_client.sh version matches your Dokku version
cd ~/.dokku
git checkout <tag/branch>

# add the following to either your
# .bashrc, .bash_profile, or .profile file
alias dokku='$HOME/.dokku/contrib/dokku_client.sh'
```

Alternatively, if using another shell such as **zsh**, create an alias command which invokes the script using **bash**:

```shell
# zsh: add the following to either .zshenv or .zshrc
alias dokku='bash $HOME/.dokku/contrib/dokku_client.sh'

# fish: add the following to ~/.config/fish/config.fish
alias dokku 'bash $HOME/.dokku/contrib/dokku_client.sh'

# csh: add the following to .cshrc
alias dokku 'bash $HOME/.dokku/contrib/dokku_client.sh'
```

### Usage

All commands have the application name automatically set via the `--app` flag on the remote server, and thus the app name does not need to be specified manually for core plugins.

The client supports several environment variables:

- `DOKKU_HOST` (default: `dokku` git remote): Used to interact with a specific remote server. Can be overridden via the `--remote` flag.
- `DOKKU_PORT` (default: `22`): Used to specify a port to connect to the Dokku server on.

It also supports several flags (all flags unspecified here are passed as is to the server):

- `--app`: Override the remote app in use.
- `--trace`: Enable trace mode.
- `--remote`: Override the remote server.
- `--global`: Unsets the "app" value. May not be supported for the specified command.

In addition, the following commands have special local side-effects:

- `apps:create`:
    - If no local `--app` flag is specified or detected from a `dokku` git remote, a random name is generated and used for the app.
    - The `dokku` git remote is set if not already set.
- `apps:destroy`:
    - Removes the local `dokku` git remote if set.

#### Specifying a remote

While many users will only ever push to a single dokku remote from a given repository, it may be useful to point the same repository at multiple remotes, especially in the case of pre-production environments. This can be done via the `remote` commands. These commands are _only_ available within the official remote client, and are not server-side commands.

To see your currently configured remote, run `dokku remote`. This will respect the `--remote` flag.

To list all available remotes, run `dokku remote:list`. This is equivalent to running `git remote`.

To add a remote, run `dokku remote:add $REMOTE_NAME dokku@dokku.me:$APP_NAME`, replacing `$REMOTE_NAME` with your desired remote name, the `$APP_NAME` with your desired app name and the host `dokku.me` with your Dokku host. This is equivalent to running `git remote add $REMOTE_NAME dokku@dokku.me:$APP_NAME`.

To remove a remote, run `dokku remote:remove $REMOTE_NAME`. This is equivalent to running `git remote remove $REMOTE_NAME`.

By default, the remote in use is `dokku`. To change the remote the client uses by default, run `dokku remote:set $REMOTE_NAME`, where `$REMOTE_NAME` is the name of your desired remote. This will not check the value specified, so that remote _must_ be available or errors will occur when running normal client commands. Note that specifying `--remote` when running dokku commands will override this configuration value.

To clear the configured remote, run `dokku remote:unset`. The client will unset the configured remote, and will default back to `dokku`, or the value of `--remote` if specified.

## Unofficial Clients

Please refer to the [community clients](/docs/community/clients.md) list for more details.
