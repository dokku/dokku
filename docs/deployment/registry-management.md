# Registry Management

> New as of 0.25.0

```
registry:login [--password-stdin] <server> <username> [<password>] # Login to a docker registry
registry:report [<app>] [<flag>]                                   # Displays a registry report for one or more apps
registry:set <app> <key> (<value>)                                 # Set or clear a registry property for an app
```

The registry plugin enables interacting with remote registries, which is useful when either deploying images via `git:from-image` or when interacting with custom schedulers to deploy built image artifacts.

## Usage

### Logging into a registry

The `registry:login` command can be used to log into a docker registry. The following are examples for logging into various common registries:

```shell
# hub.docker.com
dokku registry:login docker.io $USERNAME $PASSWORD

# digitalocean
# the username and password are both defined as the same api token
dokku registry:login registry.digitalocean.com $DIGITALOCEAN_API_TOKEN $DIGITALOCEAN_API_TOKEN

# github container registry
# see the following link for information on retrieving a personal access token
#   https://docs.github.com/en/packages/guides/pushing-and-pulling-docker-images#authenticating-to-github-container-registry
dokku registry:login ghcr.io $USERNAME $REGISTRY_PAT_TOKEN

# quay
# a robot user may be used to login
dokku registry:login quay.io $USERNAME $PASSWORD
```

For security reasons, the password may also be specified as stdin by specifying the `--password-stdin` flag. This is supported regardless of the registry being logged into.

```shell
echo "$PASSWORD" | dokku registry:login docker.io $USERNAME
```

For certain Docker registries - such as Amazon ECR or Google's GCR registries - users may instead wish to use a docker credential helper to automatically authenticate against a server; please see the documentation regarding the credential helper in question for further setup instructions.

### Pushing images on build

To push the image on release, set the `push-on-release` property to `true` via the `registry:set` command. The default value for this property is `false`.

```shell
dokku registry:set node-js-app push-on-release true
```

This property can be set for a single app or globally via the `--global` flag. When set globally, the app-specific value will always overide the global value. The default global value for this property is `false`.

```shell
dokku registry:set --global push-on-release true
```

Setting the property value to an empty string will reset the value to the system default. Resetting the value can be done per app or globally.

```shell
# per-app
dokku registry:set node-js-app push-on-release

# globally
dokku registry:set --global push-on-release
```
