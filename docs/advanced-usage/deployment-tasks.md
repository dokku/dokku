# Deployment Tasks

> New as of 0.5.0

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

## `app.json` deployment tasks

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

## Procfile Release command

> New as of 0.14.0

The `Procfile` also supports a special `release` command which acts in a similar way to the [Heroku Release Phase](https://devcenter.heroku.com/articles/release-phase). This command is executed _after_ an app's docker image is built, but _before_ any containers are scheduled. This is also run _after_ any command executed by `scripts.dokku.predeploy`.

To use the `release` command, simply add a `release` stanza to your Procfile.

```Procfile
release: curl https://some.external.api.service.com/deployment?state=built
```

Unlike the `scripts.dokku.predeploy` command, changes made during by the `release` command are _not_ persisted to disk.

> Warning: scaling the release command up will likely result in unspecified issues within your deployment, and is highly discouraged.
