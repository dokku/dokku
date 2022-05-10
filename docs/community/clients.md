# Clients

Given the constraints, running Dokku commands remotely via SSH is fine. For certain configurations, the extra complication of manually invoking ssh can be a burden.

The easiest way to interact with Dokku remotely is by using the official client. Documented below are the various clients that you may wish to use.

## Official Client

See the [remote commands documentation](/docs/deployment/remote-commands.md) for more information on how to install and use the official client.

## (nodejs) dokku-toolbelt

Dokku-toolbelt is a node-based cli wrapper that proxies requests to the Dokku command running on remote hosts. You can install it via the following shell command (assuming you have nodejs and npm installed):

```shell
npm install -g dokku-toolbelt
```

See [documentation here](https://www.npmjs.com/package/dokku-toolbelt) for more information.

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
