# Application Management

> [!IMPORTANT]
> New as of 0.3.1

```
apps:clone <old-app> <new-app>                 # Clones an app
apps:create <app>                              # Create a new app
apps:destroy <app>                             # Permanently destroy an app
apps:exists <app>                              # Checks if an app exists
apps:list [--format stdout|json]               # List your apps
apps:lock <app>                                # Locks an app for deployment
apps:locked <app>                              # Checks if an app is locked for deployment
apps:rename <old-app> <new-app>                # Rename an app
apps:report [<app>] [<flag>]                   # Display report about an app
apps:set [--global] <app> <key> (<value>)      # Set or clear an apps property for an app
apps:unlock <app>                              # Unlocks an app for deployment
```

## Usage

### Listing applications

> [!IMPORTANT]
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

You can also retrieve the list of apps as a JSON array by using the `--format json` flag, which is useful for programmatic consumption:

```shell
dokku apps:list --format json
```

```json
["node-js-app","python-app"]
```

When no apps exist, the JSON output is an empty array (`[]`).

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

### Disabling automatic app creation

By default, pushing to a git remote for an app that does not yet exist on the Dokku host will cause the app to be created automatically. On shared or production hosts this may be undesirable - operators may prefer that an explicit `apps:create` is required first. The `disable-autocreation` global property controls this behavior:

```shell
dokku apps:set --global disable-autocreation true
```

While set, pushes targeting an app that does not exist will be rejected. The default behavior may be restored by passing an empty value:

```shell
dokku apps:set --global disable-autocreation
```

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

The `--force` flag can also be specified on the command vs globally:

```shell
dokku apps:destroy --force node-js-app
```

```
Destroying node-js-app (including all add-ons)
```

Destroying an application will unlink all linked services and destroy any config related to the application. Note that linked services will retain their data for later use (or removal).

### Renaming a deployed app

> [!IMPORTANT]
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

By default, Dokku will deploy the renamed app, though you can skip the deploy by using the `--skip-deploy` flag:

```shell
dokku apps:rename --skip-deploy node-js-app io-js-app
```

Remember to also change your git remote on your local machine in order to make `git push dokku main` work again. For this you can use `git remote set-url`.

```shell
git remote set-url dokku dokku@dokku.me:io-js-app
```

### Cloning an existing app

> [!IMPORTANT]
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

> [!WARNING]
> If you have exposed specific ports via `docker-options` plugin, or performed anything that cannot be done against multiple applications, `apps:clone` may result in errors.

By default, Dokku will deploy this new app, though you can skip the deploy by using the `--skip-deploy` flag:

```shell
dokku apps:clone --skip-deploy node-js-app io-js-app
```

Finally, if the application already exists, you may wish to ignore errors resulting from attempting to clone over it. To do so, you can use the `--ignore-existing` flag. A warning will be emitted, but the command will return `0`.

```shell
dokku apps:clone --ignore-existing node-js-app io-js-app
```

### Locking app deploys

> [!IMPORTANT]
> New as of 0.11.6

If you wish to disable deploying for a period of time, this can be done via deploy locks. Normally, deploy locks exist only for the duration of a deploy so as to avoid deploys from colliding, but a deploy lock can be created by running the `apps:lock` command.

```shell
dokku apps:lock node-js-app
```

```
-----> Deploy lock created
```

### Unlocking app deploys

> [!IMPORTANT]
> New as of 0.11.6

In some cases, it may be necessary to remove an existing deploy lock. This can be performed via the `apps:unlock` command.

> [!WARNING]
> Removing the deploy lock *will not* stop in progress deploys. At this time, in progress deploys will need to be manually terminated by someone with server access.

```shell
dokku apps:unlock node-js-app
```

```
 !     A deploy may be in progress.
 !     Removing the app lock will not stop in progress deploys.
-----> Deploy lock removed.
```

### Checking lock status

> [!IMPORTANT]
> New as of 0.13.0

In some cases, you may wish to inspect the state of an app lock. To do so, you can issue an `apps:lock` command. This will exit non-zero if there is no app lock in place.

```shell
dokku apps:locked node-js-app
```

```
Deploy lock does not exist
```

### Displaying reports for an app

> [!IMPORTANT]
> New as of 0.8.1

You can get a report about the deployed apps using the `apps:report` command:

```shell
dokku apps:report
```

```
=====> node-js-app app information
       App created at:              1635126111
       App dir:                     /home/dokku/node-js-app
       App deploy source:           git
       App deploy source metadata:  cd7b8afccb202f222e7dc7b427553e71ba5ddafd
       App locked:                  false
=====> python-sample app information
       App created at:              1635126000
       App dir:                     /home/dokku/python-sample
       App deploy source:
       App deploy source metadata:
       App locked:                  false
=====> ruby-sample app information
       App created at:              1635122462
       App dir:                     /home/dokku/ruby-sample
       App deploy source:           git
       App deploy source metadata:  c60921ea2799ca108276414b95ea197f16798d51
       App locked:                  false
```

You can run the command for a specific app also.

```shell
dokku apps:report node-js-app
```

```
=====> node-js-app app information
       App dir:                     /home/dokku/node-js-app
       App deploy source:           git
       App deploy source metadata:  cd7b8afccb202f222e7dc7b427553e71ba5ddafd
       App locked:                  false
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku apps:report node-js-app --app-dir
```

## Properties

### Settable properties

> [!NOTE]
> The `Report flags` column lists the CLI argument names accepted by `apps:report`. The JSON keys emitted by `apps:report --format json` are the same names with the leading `--app-` stripped (e.g. `global-disable-autocreation`). Legacy keys with the `app-` prefix (e.g. `app-global-disable-autocreation`) are also emitted during the 0.38.x deprecation window and will be removed in a future major release.

| Property | Scope | Default | Report flags | Description |
|---|---|---|---|---|
| `disable-autocreation` | global only | `false` | `--app-global-disable-autocreation` | When `true`, pushes to a remote for an app that does not yet exist are rejected instead of auto-creating the app |

### Read-only flags

The following flags surface in `apps:report` but are not managed by `apps:set`:

| Flag | Description |
|---|---|
| `--app-created-at` | UNIX timestamp of app creation |
| `--app-deploy-source` | Source kind of the last deploy (`git`, `archive`, `docker-image`, `git-sync`) |
| `--app-deploy-source-metadata` | Free-form metadata for the deploy source (commit sha, image ref, URL) |
| `--app-dir` | Absolute path of the app root on disk |
| `--app-locked` | `true` while a deploy or rebuild holds the app lock |
