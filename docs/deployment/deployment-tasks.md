# Deployment Tasks

Sometimes you need to run a command on at deployment time, but before an app is completely deployed.
Common use cases include:

* Checking a database is initialised
* Running database migrations
* Any commands required to set up the server (e.g. something like a Django `collectstatic`)

## `app.json` and `scripts.dokku`

Dokku accomplishes this by using an `app.json` file. We (mostly) use the same format as Heroku's [app.json](https://devcenter.heroku.com/articles/app-json-schema).
However, dokku currently only supports the nodes `scripts.dokku.predeploy` and `scripts.dokku.postdeploy`.
Simply place an `app.json` file in the root of your repository.
NOTE: `app.json` is only supported in buildpack-deployed apps and postdeploy changes are not committed to the app image.

### Example app.json

```
{
  "name": "barebones nodejs",
  "description": "A barebones Node.js app using Express 4.",
  "keywords": [
    "nodejs",
    "express"
  ],
  "repository": "https://github.com/michaelshobbs/node-js-sample",
  "scripts": {
    "dokku": {
      "predeploy": "touch /app/predeploy.test",
      "postdeploy": "touch /app/postdeploy.test"
    }
  },
  "env": {
    "SECRET_TOKEN": {
      "description": "A secret key for verifying the integrity of signed cookies.",
      "value": "secret"
    },
    "WEB_CONCURRENCY": {
      "description": "The number of processes to run.",
      "generator": "echo 5"
    }
  },
  "image": "gliderlabs/herokuish",
  "addons": [
    "dokku-postgres",
    "dokku-redis"
  ],
  "buildpacks": [
    {
      "url": "https://github.com/heroku/heroku-buildpack-nodejs"
    }
  ]
}
```
