# Build Management

> New as of 0.19.0

```
builds:cancel <app>                            # Cancel a running build for an app
builds:list <app>                              # List all running builds
builds:output <app>                           #  Shows build output
builds:report [<app>] [<flag>]                 # Displays a build report for one or more apps
```

## Usage

### Listing running deploys

### Viewing the status of a deploy

### Viewing build output for a deploy

### Canceling a running deploy

It can be useful to kill a deploy if that deploy does not appear to be progressing, is impacting other apps through system resource utilization, or if a successful deploy will result in app errors. To do so, the `builds:cancel` command can be used:

```shell
dokku builds:cancel node-js-app
```

This command will send a `QUIT` signal to the Process Group ID of the process handling the deploy, and should terminate all processes within that process tree. Finally, it will unlock the deploy so that a new deploy may be immediately invoked.

> Warning: This may also result in invalid app state depending upon when the app deploy was killed.
