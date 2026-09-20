# Registry Management

> [!IMPORTANT]
> New as of 0.25.0

```
registry:auth-status [--password-stdin] <app>|--global <server> [<username> [<password>]] # Reports whether the stored registry credential matches the requested state
registry:login [--global|--password-stdin] [<app>] <server> <username> [<password>]       # Login to a docker registry
registry:logout [--global] [<app>] <server>                                               # Logout from a docker registry
registry:report [<app>] [<flag>]                                                          # Displays a registry report for one or more apps
registry:set <app>|--global <key> (<value>)                                               # Set or clear a registry property for an app
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

### Checking the configured auth state

> [!IMPORTANT]
> New as of 0.38.29

The `registry:auth-status` command reports whether a stored registry credential matches a desired state without exposing it. This allows external tooling such as configuration management systems to converge idempotently, rather than re-running `registry:login` on every pass to discover that nothing needed to change.

```shell
# check whether an app is configured with the expected credentials
dokku registry:auth-status node-js-app ghcr.io $USERNAME $REGISTRY_PAT_TOKEN

# check whether the global credentials match
dokku registry:auth-status --global docker.io $USERNAME $PASSWORD

# check whether any credential is configured for a server
dokku registry:auth-status node-js-app ghcr.io
```

As with `registry:login`, the password may be provided via `STDIN`. This is the preferred form, as it keeps the password out of the process list:

```shell
echo "$PASSWORD" | dokku registry:auth-status --password-stdin node-js-app ghcr.io $USERNAME
```

The result is reported via the exit code, and nothing is printed when the state could be determined:

| Exit code | Description |
| --------- | ----------- |
| `0` | The stored credential matches the requested state. |
| `1` | There is no stored credential for the specified server. |
| `2` | There is a stored credential for the specified server, but it does not match the requested state. |
| `3` | The command was called with invalid arguments and the state could not be checked. |
| `4` | A credential is stored for the specified server, but it cannot be compared. |
| `20` | The specified app does not exist. |

The distinction between `1` and `2` allows a caller to tell a server that needs a credential created apart from a server whose existing credential would be replaced, without issuing a second call.

When called without a username, the requested state is that no credential exists for the server. An existing credential is therefore reported as `2`, and `1` is never returned. This is the check to perform before deciding whether a `registry:logout` is necessary.

Exit code `4` is returned when a credential exists but Dokku cannot read the secret it holds - either because a docker credential helper is configured for that server and stores the secret outside of `config.json`, or because the registry issued an identity token in place of a password. In both cases a username that differs from the one supplied is still reported as `2`, so `4` means only that an otherwise matching credential could not be confirmed. Callers should treat `4` as "unknown" rather than retrying the login indefinitely.

Unlike `registry:login`, the `--global` flag is required in order to check the global credentials - the deprecated implicit-global form is not accepted. A per-app check reads only the credentials created by `registry:login <app>` and does not fall back to the global credentials, so an app configured with a credential identical to the global one is still reported as needing one of its own. To see which credentials would actually be used for an app, read the `computed-auth-servers` field of `registry:report`.

> [!WARNING]
> A `0` exit code means the stored credential matches the one supplied, not that it still authenticates against the registry. A password that has been revoked upstream compares equal here and fails at push or pull time.

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

Instead of setting the image repository name on a per-app basis, it can be set via a template with the `image-repo-template` property. The property can be set globally or for a specific app, with the per-app value overriding the global value when both are present:

```shell
# globally
dokku registry:set --global image-repo-template "my-awesome-prefix/{{ .AppName }}"

# per-app
dokku registry:set node-js-app image-repo-template "my-awesome-prefix/{{ .AppName }}-prod"
```

Dokku uses a Golang template and has access to the `AppName` variable as shown above. The per-app `image-repo` property always takes precedence over the rendered template when both are set.

Setting the property value to an empty string will reset the value to the system default. Resetting the value can be done per app or globally.

```shell
# per-app
dokku registry:set node-js-app image-repo-template

# globally
dokku registry:set --global image-repo-template
```

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

#### Recovering from local image GC

When `push-on-release` is enabled, Dokku treats the remote registry as the canonical store for app images. If a local image disappears - for example because a `docker image prune` cron ran, the host rebooted, or an operator removed it manually - subsequent commands that need the image (`ps:restart`, `ps:scale`, `dokku run`, `domains:add`, `certs:add`, etc.) will pull the missing tag back from the registry automatically. Per-app registry credentials configured via `registry:login` are honored during the pull. This recovery covers the deployed numeric tag pushed by the registry plugin - a missing `latest` tag will be ignored.

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

## Properties

### Settable properties

> [!NOTE]
> The `Report flags` column lists the CLI argument names accepted by `registry:report`. The JSON keys emitted by `registry:report --format json` are the same names with the leading `--registry-` stripped (e.g. `image-repo`, `global-server`, `computed-push-on-release`). Legacy keys with the `registry-` prefix (e.g. `registry-image-repo`) are also emitted during the 0.38.x deprecation window and will be removed in a future major release.

| Property | Scope | Default | Report flags | Description |
|---|---|---|---|---|
| `image-repo` | app only | `dokku/<app>` | `--registry-image-repo`, `--registry-computed-image-repo` | Repository name used when pushing the app's image (overrides the global template) |
| `image-repo-template` | app + global | none | `--registry-image-repo-template`, `--registry-global-image-repo-template`, `--registry-computed-image-repo-template` | Go template used to compute the per-app image repository when `image-repo` is unset |
| `push-extra-tags` | app + global | none | `--registry-push-extra-tags`, `--registry-global-push-extra-tags`, `--registry-computed-push-extra-tags` | Comma-separated list of additional tags pushed alongside the deploy tag |
| `push-on-release` | app + global | `false` | `--registry-push-on-release`, `--registry-global-push-on-release`, `--registry-computed-push-on-release` | When `true`, pushes the image to the registry on every successful build |
| `server` | app + global | none | `--registry-server`, `--registry-global-server`, `--registry-computed-server` | Registry server host (e.g. `ghcr.io`) used when pushing images |

### Read-only report fields

> [!IMPORTANT]
> New as of 0.38.29

The following `registry:report` fields describe credentials created by `registry:login` and are not settable via `registry:set`.

| Report flag | Scope | Description |
|---|---|---|
| `--registry-auth-servers` | app only | Comma-separated list of servers the app has its own credentials for |
| `--registry-global-auth-servers` | app + global | Comma-separated list of servers the global credentials cover |
| `--registry-computed-auth-servers` | app + global | Comma-separated list of servers whose credentials would be used for the app |

Servers are listed under the name they would be passed to `registry:login`, so a Docker Hub credential is reported as `docker.io` rather than the index server address docker stores it under. Only the server names are reported; no part of a credential is ever emitted.

Docker is pointed at one config directory or the other rather than a merge of the two, so an app with any credential of its own shadows the global credentials entirely. The computed value is therefore the app's own list when it has one and the global list otherwise, rather than a per-server merge of the two.
