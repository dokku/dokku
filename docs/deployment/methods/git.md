# GIT Deployment

> Subcommands new as of 0.12.0

```
git:initialize <app>                     # Initialize a git repository for an app
git:report [<app>] [<flag>]              # Displays a git report for one or more apps
git:set <app> <key> (<value>)            # Set or clear a git property for an app
```

GIT-based deployment has been the traditional method of deploying applications in Dokku. As of v0.12.0, Dokku introduces a few ways to customize the experience of deploying via `git push`. A GIT-based deployment currently supports building applications via both [Buildpack](/docs/deployment/methods/buildpacks.md) and [Dockerfile](/docs/deployment/methods/dockerfiles.md). 

## Usage

### Initializing an Application

When an application is created via `git push`, Dokku will create the proper `pre-receive` hook in order to execute the build pipeline. In certain cases - such as when fronting deploys with the [`git-http-backend`](https://git-scm.com/docs/git-http-backend) - this may not be correctly created. As an alternative, the `git:initialize` command can be used to trigger this creation:

```shell
# on the Dokku host

# overrides any existing pre-receive hook
dokku git:initialize node-js-app
```

In order for the above command to succeed, the application _must_ already exist. 

> Warning: If the pre-receive hook was customized in any way, this will overwrite that hook with the current defaults for Dokku.

### Changing the Deploy Branch

By default, Dokku will deploy code pushed to the `master` branch. In order to quickly deploy a different local branch, the following GIT command can be used:

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

Pushing multiple branches can also be supportec by creating a [receive-branch](/docs/development/plugin-triggers.md#receive-branch) plugin trigger in a custom plugin.

### Configuring the GIT_REV Environment Variable

> New as of 0.12.0

Application deployments will include a special `GIT_REV` environment variable containing the current deployment sha being deployed. For rebuilds, this sha will remain the same.

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
