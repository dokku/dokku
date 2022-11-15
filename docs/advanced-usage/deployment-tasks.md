# Deployment Tasks

> New as of 0.5.0

## Usage

### Overview

Sometimes you need to run a command on deployment time, but before an app is completely deployed. Common use cases include:

- Checking a database is initialized
- Running database migrations
- Any commands required to set up the server (e.g. something like a Django `collectstatic`)

To support this, Dokku provides support for a special `release` command within your app's `Procfile`, as well as a special `scripts.dokku` key inside of your app's `app.json` file. Be aware that all commands are run within the context of the built docker image - no commands affect the host unless there are volume mounts attached to your app.

Each "phase" has different expectations and limitations:

- `app.json`: `scripts.dokku.predeploy`
    - When to use: This should be used if your app does not support arbitrary build commands and you need to make changes to the built image.
    - Are changes committed to the image at this phase: Yes
    - Example use-cases
        - Bundling assets in a slightly different way
        - Installing a custom package from source or copying a binary into place
- `app.json`: `scripts.dokku.postdeploy`
    - When to use: This should be used in conjunction with external systems to signal the completion of your deploy.
    - Are changes committed to the image at this phase: No
    - Example use-cases
        - Notifying slack that your app is deployed
        - Coordinating traffic routing with a central load balancer
- `app.json`: `scripts.postdeploy`
    - When to use: This should be used when you wish to run a command _once_, after the app is created and not on subsequent deploys to the app.
    - Are changes committed to the image at this phase: No
    - Example use-cases
        - Setting up OAuth clients and DNS
        - Loading seed/test data into the app’s test database
- `Procfile`: `release`
    - When to use: This should be used in conjunction with external systems to signal the completion of your app image build.
    - Are changes committed to the image at this phase: No
    - Example use-cases
        - Sending CSS, JS, and other assets from your app’s slug to a CDN or S3 bucket
        - Priming or invalidating cache stores
        - Running database migrations

Additionally, if using a Dockerfile with an `ENTRYPOINT`, the deployment task is passed to that entrypoint as is. The exceptions are if the entrypoint is one of the following:

- `["/tini", "--"]`
- `["/bin/tini", "--"]`
- `["/usr/bin/tini", "--"]`
- `["/usr/local/bin/tini", "--"]`

Please keep the above in mind when utilizing deployment tasks.

> To execute commands on the host during a release phase, see the [plugin creation documentation](/docs/development/plugin-creation.md) docs for more information on building your own custom plugin.

### Changing the `app.json` location

When deploying a monorepo, it may be desirable to specify the specific path of the `app.json` file to use for a given app. This can be done via the `app-json:set` command. If a value is specified and that file does not exist within the repository, Dokku will continue the build process as if the repository has no `app.json` file.

```shell
dokku app-json:set node-js-app appjson-path second-app.json
```

The default value may be set by passing an empty value for the option:

```shell
dokku app-json:set node-js-app appjson-path
```

The `appjson-path` property can also be set globally. The global default is `app.json`, and the global value is used when no app-specific value is set.

```shell
dokku app-json:set --global appjson-path global-app.json
```

The default value may be set by passing an empty value for the option.

```shell
dokku app-json:set --global appjson-path
```

### Displaying app-json reports for an app

> New as of 0.25.0

You can get a report about the app's storage status using the `app-json:report` command:

```shell
dokku app-json:report
```

```
=====> node-js-app app-json information
       App-json computed appjson path: app2.json
       App-json global appjson path:   app.json
       App-json appjson path:          app2.json
=====> python-sample app-json information
       App-json computed appjson path: app.json
       App-json global appjson path:   app.json
       App-json appjson path:
=====> ruby-sample app-json information
       App-json computed appjson path: app.json
       App-json global appjson path:   app.json
       App-json appjson path:
```

You can run the command for a specific app also.

```shell
dokku app-json:report node-js-app
```

```
=====> node-js-app app-json information
       App-json computed appjson path: app2.json
       App-json global appjson path:   app.json
       App-json appjson path:          app2.json
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku app-json:report node-js-app --app-json-appjson-path
```

```
app2.json
```

### Deployment tasks

#### `app.json` deployment tasks

Dokku provides limited support for the `app.json` manifest from Heroku (documentation available [here](https://devcenter.heroku.com/articles/app-json-schema)). The keys available for use with Deployment Tasks are:

- `scripts.dokku.predeploy`: This is run _after_ an app's docker image is built, but _before_ any containers are scheduled. Changes made to your image are committed at this phase.
- `scripts.dokku.postdeploy`: This is run _after_ an app's containers are scheduled. Changes made to your image are _not_ committed at this phase.
- `scripts.postdeploy`: This is run _after_ an app's containers are scheduled. Changes made to your image are _not_ committed at this phase.

For buildpack-based deployments, the location of the `app.json` file should be at the root of your repository. Dockerfile-based app deploys should have the `app.json` in the configured `WORKDIR` directory; otherwise Dokku defaults to the buildpack app behavior of looking in `/app`.

> Warning: Any failed `app.json` deployment task will fail the deploy. In the case of either phase, a failure will not affect any running containers.

The following is an example `app.json` file. Please note that only the `scripts.dokku.predeploy` and `scripts.dokku.postdeploy` tasks are supported by Dokku at this time. All other fields will be ignored and can be omitted.

```json
{
  "scripts": {
    "dokku": {
      "predeploy": "touch /app/predeploy.test",
      "postdeploy": "curl https://some.external.api.service.com/deployment?state=success"
    },
    "postdeploy": "curl https://some.external.api.service.com/created?state=success"
  }
}
```

#### Procfile Release command

> New as of 0.14.0

The `Procfile` also supports a special `release` command which acts in a similar way to the [Heroku Release Phase](https://devcenter.heroku.com/articles/release-phase). This command is executed _after_ an app's docker image is built, but _before_ any containers are scheduled. This is also run _after_ any command executed by `scripts.dokku.predeploy`.

To use the `release` command, simply add a `release` stanza to your Procfile.

```Procfile
release: curl https://some.external.api.service.com/deployment?state=built
```

Unlike the `scripts.dokku.predeploy` command, changes made during by the `release` command are _not_ persisted to disk.

> Warning: scaling the release command up will likely result in unspecified issues within your deployment, and is highly discouraged.
