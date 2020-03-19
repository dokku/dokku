# Clients

Given the constraints, running Dokku commands remotely via SSH is fine. For certain configurations, the extra complication of manually invoking ssh can be a burden.

While Dokku does not yet have an official client, there are a multitude of ways in which you can interact with your Dokku installation. The easiest is to use the **bash** client, though you may wish to use another.

## (bash, zsh, etc.) `dokku_client.sh`

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

Configure the `DOKKU_HOST` environment variable or run `dokku` from a repository with a git remote named `dokku` pointed at your Dokku host in order to use the script as normal.

You can also configure a `DOKKU_PORT` environment variable if you are running ssh on a non-standard port. This defaults to `22`.

## (nodejs) dokku-toolbelt

Dokku-toolbelt is a node-based cli wrapper that proxies requests to the Dokku command running on remote hosts. You can install it via the following shell command (assuming you have nodejs and npm installed):

```shell
npm install -g dokku-toolbelt
```

See [documentation here](https://www.npmjs.com/package/dokku-toolbelt) for more information.

## (python) dokku-client

dokku-client is an extensible python-based cli wrapper for remote Dokku hosts. You can install it via the following shell command (assuming you have python and pip installed):

```shell
pip install dokku-client
```

See [documentation here](https://github.com/adamcharnock/dokku-client) for more information.

## (ruby) Dokku CLI

Dokku CLI is a rubygem that acts as a client for your Dokku installation. You can install it via the following shell command (assuming you have ruby and rubygems installed):

```shell
gem install dokku-cli
```

See [documentation here](https://github.com/SebastianSzturo/dokku-cli) for more information.

## (ruby) DokkuClient

DokkuClient is another rubygem that acts as a client for your Dokku installation with built-in support for certain external plugins. You can install it via the following shell command (assuming you have ruby and rubygems installed):

```shell
gem install dokku_client
```

See [documentation here](https://github.com/netguru/dokku_client) for more information.

## (ruby) Dokkufy

Dokkufy is a rubygem that handles automation of certain tasks, such as Dokku setup, plugin installation, etc. You can install it via the following shell command (assuming you have ruby and rubygems installed):

```shell
gem install dokkufy
```

See [documentation here](https://github.com/cbetta/dokkufy) for more information.

## (ruby) Dockland

Dockland is a rubygem that acts as a client for your Dokku installation. You can install it via the following shell command (assuming you have ruby and rubygems installed):

```shell
gem install dockland
```

See [documentation here](https://github.com/uetchy/dockland) for more information.
