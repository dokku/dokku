# app.json

`app.json` is a manifest format for describing web apps. It declares cron tasks, healthchecks, and other information required to run an app on Dokku. This document describes the schema in detail.

> [!IMPORTANT]
> While the `app.json` format used by Dokku is based on the one [supported by Heroku](https://devcenter.heroku.com/articles/app-json-schema), not all Heroku functionality is supported by Dokku.

## Cron

```json
{
  "cron": [
    {
      "command": "echo 'hello'",
      "schedule": "0 1 * * *"
    }
  ]
}
```

(list, optional) A list of cron resources. Keys are the names of the process types. The values are an object containing one or more of the following properties:

- `command`: (string, required)
- `maintenance`: (boolean, optional)
- `schedule`: (string, required)
- `concurrency_policy`: (string, optional, default: `allow`, options: `allow`, `forbid`, `replace`)

## Env

```json
{
  "env": {
    "SIMPLE_VAR": "default_value",
    "SECRET_KEY": {
      "description": "A secret key for signing tokens",
      "generator": "secret"
    },
    "DATABASE_URL": {
      "description": "PostgreSQL connection string",
      "required": true
    },
    "OPTIONAL_VAR": {
      "description": "An optional configuration value",
      "value": "default",
      "required": false
    },
    "SYNC_VAR": {
      "description": "A variable that updates on every deploy",
      "value": "synced_value",
      "sync": true
    }
  }
}
```

(object, optional) A key-value object for environment variable configuration. Keys are the variable names. Values can be either a string (used as the default value) or an object with the following properties:

- `description`: (string, optional) Human-readable explanation of the variable's purpose
- `value`: (string, optional) Default value for the variable
- `required`: (boolean, optional, default: `true`) Whether the variable must have a value
- `generator`: (string, optional) Function to generate the value. Currently only `"secret"` is supported, which generates a 64-character cryptographically secure hex string
- `sync`: (boolean, optional, default: `false`) If `true`, the value will be set on every deploy, overwriting any existing value

### Behavior

Environment variables from `app.json` are processed during the first deploy, before the predeploy script runs. The behavior depends on the variable configuration:

1. **Variables with `value` or simple string**: The default value is set if the variable doesn't already exist
2. **Variables with `generator: "secret"`**: A random 64-character hex string is generated if the variable doesn't exist
3. **Required variables without a value or generator**: If a TTY is available, the user is prompted for a value. Otherwise, the deploy fails with an error
4. **Optional variables without a value**: Skipped silently if no TTY is available

On subsequent deploys:
- Variables are NOT re-set unless `sync: true` is specified
- Variables with `sync: true` are always set to their configured value, overwriting any manual changes
- Variables that already have values are not modified

### Examples

**Simple default value:**
```json
{
  "env": {
    "WEB_CONCURRENCY": "5"
  }
}
```

**Generated secret (recommended for API keys, tokens, etc.):**
```json
{
  "env": {
    "SECRET_KEY_BASE": {
      "description": "Base secret for session encryption",
      "generator": "secret"
    }
  }
}
```

**Required variable that must be provided:**
```json
{
  "env": {
    "DATABASE_URL": {
      "description": "PostgreSQL connection URL",
      "required": true
    }
  }
}
```

**Variable that stays in sync with app.json:**
```json
{
  "env": {
    "FEATURE_FLAGS": {
      "value": "new_ui,dark_mode",
      "sync": true
    }
  }
}
```

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
- `service`: (map of string to oject, optional) governs how non-web processes are exposed as services on the network

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

### Service

```json
{
  "formation": {  
    "internal-web": {
      "service": {
        "exposed": true
      }
    }
  }
}
```

(object, optional) A key-value object specifying how to expose non-web processes as services.

- `service`: (boolean, optional) Whether to expose a process as a network service. The `PORT` variable will be set to 5000.

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
