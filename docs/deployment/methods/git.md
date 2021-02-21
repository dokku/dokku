# Git Deployment

> Subcommands new as of 0.12.0

```
git:allow-host <host>                              # Adds a host to known_hosts
git:auth <host> [<username> <password>]            # Configures netrc authentication for a given git server
git:sync [--build] <app> <repository> [<git-ref>]  # Clone or fetch an app from remote git repo
git:initialize <app>                               # Initialize a git repository for an app
git:public-key                                     # Outputs the dokku public deploy key
git:report [<app>] [<flag>]                        # Displays a git report for one or more apps
git:set <app> <key> (<value>)                      # Set or clear a git property for an app
```

Git-based deployment has been the traditional method of deploying applications in Dokku. As of v0.12.0, Dokku introduces a few ways to customize the experience of deploying via `git push`. A Git-based deployment currently supports building applications via:

- [Cloud Native Buildpacks](/docs/deployment/methods/cloud-native-buildpacks.md)
- [Herokuish Buildpack](/docs/deployment/methods/herokuish-buildpacks.md)
- [Dockerfiles](/docs/deployment/methods/dockerfiles.md)

## Usage

> Warning: Pushing from a shallow clone is not currently supported and may have undefined behavior. Please unshallow your local repository before pushing to a Dokku app to avoid potential errors in the deployment process.

### Initializing an application

When an application is created via `git push`, Dokku will create the proper `pre-receive` hook in order to execute the build pipeline. In certain cases - such as when fronting deploys with the [`git-http-backend`](https://git-scm.com/docs/git-http-backend) - this may not be correctly created. As an alternative, the `git:initialize` command can be used to trigger this creation:

```shell
# on the Dokku host

# overrides any existing pre-receive hook
dokku git:initialize node-js-app
```

In order for the above command to succeed, the application *must* already exist. 

> Warning: If the `pre-receive` hook was customized in any way, this will overwrite that hook with the current defaults for Dokku.

### Changing the deploy branch

By default, Dokku will deploy code pushed to the `master` branch. In order to quickly deploy a different local branch, the following Git command can be used:

```shell
# on the local machine

# where `SOME_BRANCH_NAME` is the name of the branch
git push dokku SOME_BRANCH_NAME:master
```

In `0.12.0`, the correct way to change the deploy branch is to use the `git:set` Dokku command.

```shell
# on the Dokku host

# override for all applications
dokku git:set --global deploy-branch SOME_BRANCH_NAME

# override for a specific app
# where `SOME_BRANCH_NAME` is the name of the branch
dokku git:set node-js-app deploy-branch SOME_BRANCH_NAME
```

Pushing multiple branches can also be supported by creating a [receive-branch](/docs/development/plugin-triggers.md#receive-branch) plugin trigger in a custom plugin.

### Configuring the `GIT_REV` environment variable

> New as of 0.12.0

Application deployments will include a special `GIT_REV` environment variable containing the current deployment sha being deployed. For rebuilds, this SHA will remain the same.

To configure the name of the `GIT_REV` environment variable, run the `git:set` command as follows:

```shell
# on the Dokku host

# override for a specific app
dokku git:set node-js-app rev-env-var DOKKU_GIT_REV
```

This behavior can be disabled entirely on a per-app basis by setting the `rev-env-var` value to an empty string:

```shell
# on the Dokku host

# override for a specific app
dokku git:set node-js-app rev-env-var ""
```

### Keeping the `.git` directory

By default, Dokku will remove the contents of the `.git` before triggering a build for a given app. This is generally a safe default as shipping the entire source code history of your app in the deployed image artifact is unnecessary as it increases bloat and potentially can leak information if there are any security issues with your app code.

To enable the `.git` directory, run the `git:set` command as follows:

```shell
# on the Dokku host

# keep the .git directory during builds
dokku git:set node-js-app keep-git-dir true
```

The default behavior is to delete this directory and it's contents. To revert to the default behavior, the `keep-git-dir` value can be set to either an empty string or `false`.

```shell
# on the Dokku host

# delete the .git directory during builds (default)
dokku git:set node-js-app keep-git-dir false

# delete the .git directory during builds (default)
dokku git:set node-js-app keep-git-dir ""
```

Please keep in mind that setting `keep-git-dir` to `true` may result in unstaged changes shown within the built container due to the build process generating application changes within the built app directory.

### Initializing an app repository from a remote repository

> New as of 0.23.0

A Dokku app repository can be initialized or updated from a remote git repository via the `git:sync` command. This command will either clone or fetch updates from a remote repository and has undefined behavior if the history cannot be fast-fowarded to the referenced repository reference. Any repository that can be cloned by the `dokku` user can be specified.

> The application must exist before the repository can be initialized

```shell
dokku git:sync node-js-app https://github.com/heroku/node-js-getting-started.git
```

The `git:sync` command optionally takes an optional third parameter containing a git reference, which may be a branch, tag, or specific commit.

```shell
# specify a branch
dokku git:sync node-js-app https://github.com/heroku/node-js-getting-started.git main

# specify a tag
dokku git:sync node-js-app https://github.com/heroku/node-js-getting-started.git 1

# specify a commit
dokku git:sync node-js-app https://github.com/heroku/node-js-getting-started.git 97e6c72491c7531507bfc5413903e0e00e31e1b0
```

By default, this command does not trigger an application build. To do so during a `git:sync`, specify the `--build` flag.

```shell
dokku git:sync --build node-js-app https://github.com/heroku/node-js-getting-started.git
```

### Initializing from private repositories

> New as of 0.24.0

Initializing from a private repository requires one of the following:

- A Public SSH Key (`id_rsa.pub` file) configured on the remote server, with the associated private key (`id_rsa`) in the Dokku server's `/home/dokku/.ssh/` directory.
- A configured [`.netrc`](https://www.gnu.org/software/inetutils/manual/html_node/The-_002enetrc-file.html) entry.

Dokku provides the `git:auth` command which can be used to configure a `netrc` entry for the remote server. This command can be used to add or remove configuration for any remote server.

```shell
# add credentials for github.com
dokku git:auth github.com username personal-access-token

# remove credentials for github.com
dokku git:auth github.com
```

For syncing to a private repository stored on a remote Git product such as Github or Gitlab, Dokku's recommendation is to use a personal access token on a bot user where possible. Please see your service's documentation for information regarding the recommended best practices.

### Allowing remote repository hosts

By default, the Dokku host may not have access to a server containing the remote repository. This can be initialized via the `git:allow-host` command.

```shell
dokku git:allow-host github.com
```

Note that this command is currently not idempotent and may add duplicate entries to the `~dokku/.ssh/known_hosts` file.

### Verifying the cloning public key

In order to clone a remote repository, the remote server should have the Dokku host's public key configured. This plugin does not currently create this key, but if there is one available, it can be shown via the `git:public-key` command.

```shell
dokku git:public-key
```

If there is no key, an error message is shown that displays the command that can be run on the Dokku server to generate a new public/private ssh key pair.
