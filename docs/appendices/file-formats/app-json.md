# app.json

`app.json` is a manifest format for describing web apps. It declares cron tasks, healthchecks, and other information required to run an app on Dokku. This document describes the schema in detail.

> [!IMPORTANT]
> While the `app.json` format used by Dokku is based on the one [supported by Heroku](https://devcenter.heroku.com/articles/app-json-schema), not all Heroku functionality is supported by Dokku.

## Cron

```json
{
  "crons": [
    {
      "command": "echo 'hello'",
      "schedule": "0 1 * * *"
    }
  ]
}
```

(list, optional) A list of cron resources. Keys are the names of the process types. The values are an object containing one or more of the following properties:

- `command`: (string, required)
- `schedule`: (string, required)

## Formation

```json
{
  "formation": {
    "web": {
      "max_parallel": 1,
      "quantity": 1
    }
  }
}
```

(object, optional) A key-value object for process type configuration. Keys are the names of the process types. The values are an object containing one or more of the following properties:

- `autoscaling` (map of string to object, optional) autoscaling rules. See the autoscaling section for more details
- `max_parallel`: (int, optional) number of instances to deploy in parallel at a given time
- `quantity`: (int, optional) number of processes to maintain. Default 1 for web processes, 0 for all others.

### Autoscaling

```json
{
  "formation": {  
    "web": {
      "autoscaling": {
        "cooldown_period_seconds": 300,
        "max_quantity": 10,
        "min_quantity": 1,
        "polling_interval_seconds": 30,
        "triggers": {
          "http": {
            "metadata": {
              "url": "https://example.com/health"
            }
          }
        }
      }
    }
  }
}
```

(object, optional) A key-value object for autoscaling configuration. Keys are the names of the process types. The values are an object containing one or more of the following properties:

- `cooldown_period_seconds`: (int, optional)
- `max_quantity`: (int, optional)
- `min_quantity`: (int, optional)
- `polling_interval_seconds`: (int, optional)
- `triggers`: (object, optional)

An autoscaling trigger consists of the following properties:

- `name`: (string, optional)
- `type`: (string, optional)
- `metadata`: (object, optional)

## Healthchecks

```json
{
  "healthchecks": {
    "web": [
      {
        "type":        "startup",
        "name":        "web check",
        "description": "Checking if the app responds to the /health/ready endpoint",
        "path":        "/health/ready",
        "attempts": 3
      }
    ]
  }
}
```

(object, optional) A key-value object specifying healthchecks to run for a given process type.

- `attempts`: (int, optional)
- `command`: (list of strings, optional)
- `content`: (string, optional)
- `httpHeaders`: (list of header objects, optional)
- `initialDelay`: (int, optional)
- `listening`: (boolean, optional)
- `name`: (string, optional)
- `path`: (string, optional)
- `port`: (int, optional)
- `scheme`: (string, optional)
- `timeout`: (int, optional)
- `type`: (string, optional)
- `uptime`: (int, optional)
- `wait`: (int, optional)
- `warn`: (boolean, optional)
- `onFailure`: (object, optional)

## Scripts

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

(object, optional) A key-value object specifying scripts or shell commands to execute at different stages in the build/release process.

- `dokku.predeploy`: (string, optional)
    - When to use: This should be used if your app does not support arbitrary build commands and you need to make changes to the built image.
    - Are changes committed to the image at this phase: Yes
    - Example use-cases
        - Bundling assets in a slightly different way
        - Installing a custom package from source or copying a binary into place
- `dokku.postdeploy`: (string, optional)
    - When to use: This should be used in conjunction with external systems to signal the completion of your deploy.
    - Are changes committed to the image at this phase: No
    - Example use-cases
        - Notifying slack that your app is deployed
        - Coordinating traffic routing with a central load balancer
- `postdeploy`: (string, optional)
    - When to use: This should be used when you wish to run a command _once_, after the app is created and not on subsequent deploys to the app.
    - Are changes committed to the image at this phase: No
    - Example use-cases
        - Setting up OAuth clients and DNS
        - Loading seed/test data into the app’s test database
