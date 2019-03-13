# Application Management

> New as of 0.3.1

```
apps:clone <old-app> <new-app>                 # Clones an app
apps:create <app>                              # Create a new app
apps:destroy <app>                             # Permanently destroy an app
apps:exists <app>                              # Checks if an app exists
apps:list                                      # List your apps
apps:lock <app>                                # Locks an app for deployment
apps:locked <app>                              # Checks if an app is locked for deployment
apps:rename <old-app> <new-app>                # Rename an app
apps:report [<app>] [<flag>]                   # Display report about an app
apps:unlock <app>                              # Unlocks an app for deployment
```

## Usage

### Listing applications

> New as of 0.8.1. Use the `apps` command for older versions.

You can easily list all available applications using the `apps:list` command:

```shell
dokku apps:list
```

```
=====> My Apps
node-js-app
python-app
```

Note that you can easily hide extra output from Dokku commands by using the `--quiet` flag, which makes it easier to parse on the command line.

```shell
dokku --quiet apps:list
```

```
node-js-app
python-app
```

### Checking if an application exists

For CI/CD pipelines, it may be useful to see if an application exists before creating a "review" application for a specific branch. You can do so via the `apps:exists` command:

```shell
dokku apps:exists  node-js-app
```

```
App does not exist
```

The `apps:exists` command will return non-zero if the application does not exist, and zero if it does.

### Manually creating an application

A common pattern for deploying applications to Dokku is to configure an application before deploying it. You can do so via the `apps:create` command:

```shell
dokku apps:create node-js-app
```

```
Creating node-js-app... done
```

Once created, you can configure the application as normal, and deploy the application whenever ready. This is useful for cases where you may wish to do any of the following kinds of tasks:

- Configure domain names and SSL certificates.
- Create and link datastores.
- Set environment variables.

### Removing a deployed app

In some cases, you may need to destroy an application, whether it is because the application is temporary or because it was misconfigured. In these cases, you can use the `apps:destroy` command. Performing any destructive actions in Dokku requires confirmation, and this command will ask for the name of the application being deleted before doing so.

```shell
dokku apps:destroy node-js-app
```

```
 !     WARNING: Potentially Destructive Action
 !     This command will destroy node-js-app (including all add-ons).
 !     To proceed, type "node-js-app"

> node-js-app
Destroying node-js-app (including all add-ons)
```

This will prompt you to verify the application's name before destroying it. You may also use the `--force` flag to circumvent this verification process:

```shell
dokku --force apps:destroy node-js-app
```

```
Destroying node-js-app (including all add-ons)
```

Destroying an application will unlink all linked services and destroy any config related to the application. Note that linked services will retain their data for later use (or removal).

### Renaming a deployed app

> New as of 0.4.7

You can rename a deployed app using the `apps:rename` command. Note that the application *must* have been deployed at least once, or the rename will not complete successfully:

```shell
dokku apps:rename node-js-app io-js-app
```

```
Destroying node-js-app (including all add-ons)
-----> Cleaning up...
-----> Building io-js-app from herokuish...
-----> Adding BUILD_ENV to build environment...
-----> Node.js app detected

-----> Creating runtime environment

...

=====> Application deployed:
       http://io-js-app.ci.dokku.me

Renaming node-js-app to io-js-app... done
```

This will copy all of your app's contents into a new app directory with the name of your choice, delete your old app, then rebuild the new version of the app and deploy it. All of your config variables, including database urls, will be preserved.

### Cloning an existing app

> New as of 0.11.5

You can clone an existing app using the `apps:clone` command.  Note that the application *must* have been deployed at least once, or cloning will not complete successfully:

```shell
dokku apps:clone node-js-app io-js-app
```

```
Cloning node-js-app to io-js-app... done
```

This will copy all of your app's contents into a new app directory with the name of your choice and then rebuild the new version of the app and deploy it with the following caveats:

- All of your environment variables, including database urls, will be preserved.
- Custom domains are not applied to the new app.
- SSL certificates will not be copied to the new app.
- Port mappings with the scheme `https` and host-port `443` will be skipped.

> Warning: If you have exposed specific ports via `docker-options` plugin, or performed anything that cannot be done against multiple applications, `apps:clone` may result in errors.

By default, Dokku will deploy this new application, though you can skip the deploy by using the `--skip-deploy` flag:

```shell
dokku apps:clone --skip-deploy node-js-app io-js-app
```

Finally, if the application already exists, you may wish to ignore errors resulting from attempting to clone over it. To do so, you can use the `--ignore-existing` flag. A warning will be emitted, but the command will return `0`.

```shell
dokku apps:clone --ignore-existing node-js-app io-js-app
```

### Locking app deploys

> New as of 0.11.6

If you wish to disable deploying for a period of time, this can be done via deploy locks. Normally, deploy locks exist only for the duration of a deploy so as to avoid deploys from colliding, but a deploy lock can be created by running the `apps:lock` command.


```shell
dokku apps:lock node-js-app
```

```
-----> Deploy lock created
```

### Unlocking app deploys

> New as of 0.11.6

In some cases, it may be necessary to remove an existing deploy lock. This can be performed via the `apps:unlock` command.

> Warning: Removing the deploy lock *will not* stop in progress deploys. At this time, in progress deploys will need to be manually terminated by someone with server access.

```shell
dokku apps:unlock node-js-app
```

```
 !     A deploy may be in progress.
 !     Removing the app lock will not stop in progress deploys.
-----> Deploy lock removed.
```

### Checking lock status

> New as of 0.13.0

In some cases, you may wish to inspect the state of an app lock. To do so, you can issue an `apps:lock` command. This will exit non-zero if there is no app lock in place.


```shell
dokku apps:locked node-js-app
```

```
Deploy lock does not exist
```

### Displaying reports for an app

> New as of 0.8.1

You can get a report about the deployed apps using the `apps:report` command:

```shell
dokku apps:report
```

```
=====> node-js-app
       App dir:             /home/dokku/node-js-app
       Git sha:             dbddc3f
       Deploy source:       git
       Locked:              false
=====> python-sample
not deployed
=====> ruby-sample
       App dir:             /home/dokku/ruby-sample
       Git sha:             a2d477c
       Deploy source:       git
       Locked:              false
```

You can run the command for a specific app also.

```shell
dokku apps:report node-js-app
```

```
=====> node-js-app
       App dir:             /home/dokku/node-js-app
       Git sha:             dbddc3f
       Deploy source:       git
       Locked:              false
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku apps:report node-js-app --git-sha
```
