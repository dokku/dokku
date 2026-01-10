# Registry Management

> [!IMPORTANT]
> New as of 0.25.0

```
registry:login [--global|--password-stdin] [<app>] <server> <username> [<password>] # Login to a docker registry
registry:logout [--global] [<app>] <server>                                         # Logout from a docker registry
registry:report [<app>] [<flag>]                                                    # Displays a registry report for one or more apps
registry:set <app>|--global <key> (<value>)                                         # Set or clear a registry property for an app
```

The registry plugin enables interacting with remote registries, which is useful when either deploying images via `git:from-image` or when interacting with custom schedulers to deploy built image artifacts.

## Usage

### Logging into a registry

The `registry:login` command can be used to log into a docker registry. Credentials can be stored globally (for all apps) or on a per-app basis.

#### Global login

To log in globally (credentials shared by all apps), use the `--global` flag:

```shell
# hub.docker.com
dokku registry:login --global docker.io $USERNAME $PASSWORD

# digitalocean
# the username and password are both defined as the same api token
dokku registry:login --global registry.digitalocean.com $DIGITALOCEAN_API_TOKEN $DIGITALOCEAN_API_TOKEN

# github container registry
# see the following link for information on retrieving a personal access token
#   https://docs.github.com/en/packages/guides/pushing-and-pulling-docker-images#authenticating-to-github-container-registry
dokku registry:login --global ghcr.io $USERNAME $REGISTRY_PAT_TOKEN

# quay
# a robot user may be used to login
dokku registry:login --global quay.io $USERNAME $PASSWORD
```

> [!NOTE]
> For backwards compatibility, if the `--global` flag is omitted and only three arguments are provided (server, username, password), the command will behave as a global login but will show a deprecation warning.

#### Per-app login

To log in for a specific app, specify the app name as the first argument:

```shell
# log into docker.io for a specific app
dokku registry:login node-js-app docker.io $USERNAME $PASSWORD

# log into ghcr.io for a specific app
dokku registry:login node-js-app ghcr.io $USERNAME $REGISTRY_PAT_TOKEN
```

Per-app credentials are stored in `/var/lib/dokku/config/registry/$APP/config.json` and are automatically used for docker operations (build, push, pull) for that specific app.

#### Password via stdin

For security reasons, the password may also be specified as stdin by specifying the `--password-stdin` flag. This is supported for both global and per-app logins:

```shell
# global login via stdin
echo "$PASSWORD" | dokku registry:login --global --password-stdin docker.io $USERNAME

# per-app login via stdin
echo "$PASSWORD" | dokku registry:login node-js-app --password-stdin docker.io $USERNAME
```

For certain Docker registries - such as Amazon ECR or Google's GCR registries - users may instead wish to use a docker credential helper to automatically authenticate against a server; please see the documentation regarding the credential helper in question for further setup instructions.

### Logging out from a registry

The `registry:logout` command can be used to log out from a docker registry:

```shell
# global logout
dokku registry:logout --global docker.io

# per-app logout
dokku registry:logout node-js-app docker.io
```

When an app is destroyed, any per-app registry credentials are automatically removed.

### Setting a remote server

To specify a remote server registry for pushes, set the `server` property via the `registry:set` command. The default value for this property is empty string. Setting the value to `docker.io` or `hub.docker.com` will result in the computed value being empty string (as that is the default, implicit registry), while any non-zero length value will have a `/` appended to it if there is not one already.

```shell
dokku registry:set node-js-app server docker.io
```

This property can be set for a single app or globally via the `--global` flag. When set globally, the app-specific value will always overide the global value. The default global value for this property is empty string.

```shell
dokku registry:set --global server docker.io
```

Setting the property value to an empty string will reset the value to the system default. Resetting the value can be done per app or globally.

```shell
# per-app
dokku registry:set node-js-app server

# globally
dokku registry:set --global server
```

The following are the values that should be used for common remote servers:

- Amazon Elastic Container Registry:
    - value: `$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/`
    - notes: The `$AWS_ACCOUNT_ID` and `$AWS_REGION` should match the values for your account and region, respectively. Additionally, an IAM profile that allows `push` access to the repository specified by `image-repo` should be attached to your Dokku server.
- Azure Container Registry:
    - value `$REGISTRY_NAME.azurecr.io/`
    - notes: The `$AKS_REGISTRY_NAME` should match the name of the registry created on your account.
- Docker Hub:
    - value: `docker.io/`
    - notes: Requires owning the namespace used in the `image-repo` value.
- Digitalocean:
    - value: `registry.digitalocean.com/`
    - notes: Requires setting the correct `image-repo` value for your registry.
- Github Container Registry:
    - value: `ghcr.io/`
    - notes: Requires that the authenticated user has access to the namespace used in the `image-repo` value.
- Quay.io:
    - value: `quay.io/`

### Specifying an image repository name

By default, Dokku uses the value `dokku/$APP_NAME` as the image repository that is pushed and deployed. For certain registries, the `dokku` namespace may not be available to your user. In these cases, the value can be set by changing the value of the `image-repo` property via the `registry:set` command.

```shell
dokku registry:set node-js-app image-repo my-awesome-prefix/node-js-app
```

Setting the property value to an empty string will reset the value to the system default. Resetting the value has to be done per-app.

```shell
# per-app
dokku registry:set node-js-app image-repo
```

### Templating the image repository name

Instead of setting the image repository name on a per-app basis, it can be set via a template globally with the `image-repo-template` property:

```shell
dokku registry:set --global image-repo-template "my-awesome-prefix/{{ .AppName }}"
```

Dokku uses a Golang template and has access to the `AppName` variable as shown above.

### Pushing images on build

To push the image on release, set the `push-on-release` property to `true` via the `registry:set` command. The default value for this property is `false`. Setting the property to `true` will result in the image being tagged with an ID that is incremented with every release. This tag will be what is used for running app code.

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

### Push extra tags 

To push the image on release with extra tags, set the `push-extra-tags` to a comma-separated list of tags via the `registry:set` command. The default value for this property is empty. Setting the property will result in the image being tagged with extra tags every release.

```shell
# multiple-tags
dokku registry:set node-js-app push-extra-tags foo,bar
```
The `push-extra-tags` can be set to a single tag too.

```shell
# single tag
dokku registry:set node-js-app push-extra-tags foo
```

This property can be set for a single app or globally via the `--global` flag. When set globally, the app-specific value will always overide the global value. The default global value for this property is `false`.

```shell
dokku registry:set --global push-extra-tags foo,bar
```

Setting the property value to an empty string will reset the value to the system default. Resetting the value can be done per app or globally.

```shell
# per-app
dokku registry:set node-js-app push-extra-tags

# globally
dokku registry:set --global push-extra-tags
```
