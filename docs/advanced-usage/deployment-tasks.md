# Deployment Tasks

> New as of 0.5.0

Sometimes you need to run a command on at deployment time, but before an app is completely deployed.
Common use cases include:

* Checking a database is initialized
* Running database migrations
* Any commands required to set up the server (e.g. something like a Django `collectstatic`)

## `app.json` and `scripts.dokku`

Dokku accomplishes this by using an `app.json` file. The format in use is similar to format of Heroku's [app.json](https://devcenter.heroku.com/articles/app-json-schema).
However, Dokku currently only supports the nodes `scripts.dokku.predeploy` and `scripts.dokku.postdeploy`.
For buildpack apps, simply place an `app.json` file in the root of your repository.
For dockerfile apps, place `app.json` in the configured `WORKDIR` directory; otherwise Dokku defaults to the buildpack app behavior of looking in `/app`.

> NOTE: postdeploy changes are *NOT* committed to the app image.

### Example app.json

> NOTE: Only the `scripts.dokku.predeploy` and `scripts.dokku.postdeploy` tasks are supported by Dokku at this time. All other fields will be ignored and can be omitted.

```json
{
  "scripts": {
    "dokku": {
      "predeploy": "touch /app/predeploy.test",
      "postdeploy": "curl https://some.external.api.service.com/deployment?state=success"
    }
  }
}
```
