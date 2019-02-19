# Tarball Deployment

> New as of 0.4.0

```
tar:from <app> <url>                           # Loads an app tarball from url
tar:in <app>                                   # Reads an tarball containing the app from stdin
```

> When triggering `dokku ps:rebuild APP` on an application deployed via the `tar` plugin, the following may occur:
>
> - Applications previously deployed via another method (`git`): The application may revert to a state before the latest custom image tag was deployed.
> - Applications that were only ever deployed via the `tar` plugin: The application will be properly rebuilt.
>
> Please use the appropriate `tar` command when redeploying an application deployed via tarball.

## Usage

### Deploying from a tarball

In some cases, it may be useful to deploy an application from a tarball. For instance, if you implemented a non-Git based deployment plugin, tarring the generated artifact may be an easier route to interface with the existing Dokku infrastructure.

You can place the tarball on an external webserver and deploy via the `tar:from` command.

```shell
dokku tar:from node-js-app https://dokku.me/releases/node-js-app/v1
```

### Deploying via stdin

As an alternative, a deploy can be trigged from a tarball read from stdin using the `tar:in` command:

```shell
# run from the generated artifact directory
tar c . $* | dokku tar:in node-js-app
```
