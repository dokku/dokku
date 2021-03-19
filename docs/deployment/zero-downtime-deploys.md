# Zero Downtime Deploys
----

!!! tip "New as of 0.5.0"

```
checks:disable <app> [process-type(s)]   Disable zero-downtime deployment for all processes (or comma-separated process-type list) ***WARNING: this will cause downtime during deployments***
checks:enable <app> [process-type(s)]    Enable zero-downtime deployment for all processes (or comma-separated process-type list)
checks:report [<app>] [<flag>]           Displays a checks report for one or more apps
checks:run <app> [process-type(s)]       Runs zero-downtime checks for all processes (or comma-separated process-type list)
checks:skip <app> [process-type(s)]      Skip zero-downtime checks for all processes (or comma-separated process-type list)
```

By default, Dokku will wait `10` seconds after starting each container before assuming it is up and proceeding with the deploy. Once this has occurred for all containers started by an application, traffic will be switched to point to your new containers. Dokku will also wait a further `60` seconds *after* the deploy is complete before terminating old containers in order to give time for long running connections to terminate. In either case, you may have more than one container running for a given application.

You may both create user-defined checks for web processes using a `CHECKS` file, as well as customize any and all parts of this experience using the checks plugin.

!!! info
    Web checks are performed via `curl` on Dokku host. Some application code - such
    as the Django framework - checks for specific hostnames or header values, these
    checks will fail. To avoid this:

    - Remove such checks from your code: Modify your application to remove the hostname check completely.
    - Allow checks from all hostnames: Modify your application to accept a dynamically provided hostname.
    - Specify the domain within the check: See below for further documentation.

## Configuring check settings using the `config` plugin

There are certain settings that can be configured via environment variables:

- `DOKKU_DEFAULT_CHECKS_WAIT`: (default: `10`) If no user-defined checks are specified - or if the process being checked is not a `web` process - this is the period of time Dokku will wait before checking that a container is still running.
- `DOKKU_DOCKER_STOP_TIMEOUT`: (default: `10`) Configurable grace period given to the `docker stop` command. If a container has not stopped by this time, a `kill -9` signal or equivalent is sent in order to force-terminate the container. Both the `ps:stop` and `apps:destroy` commands *also* respect this value. If not specified, the Docker defaults for the [`docker stop` command](https://docs.docker.com/engine/reference/commandline/stop/) will be used.
- `DOKKU_WAIT_TO_RETIRE`: (default: `60`) After a successful deploy, the grace period given to old containers before they are stopped/terminated. This is useful for ensuring completion of long-running HTTP connections.

The following settings may also be specified in the `CHECKS` file, though are available as environment variables in order to ease application reuse.

- `DOKKU_CHECKS_WAIT`: (default: `5`) Wait this many seconds for the container to start before running checks.
- `DOKKU_CHECKS_TIMEOUT`: (default: `30`) Wait this many seconds for each response before marking it as a failure.
- `DOKKU_CHECKS_ATTEMPTS`: (default: `5`) Number of retries for to run for a specific check before marking it as a failure

## Skipping and Disabling Checks

!!! info
    Note that `checks:disable` will now (as of 0.6.0) cause downtime for that process-type during deployments. Previously, it acted as `checks:skip` currently does.

You can choose to skip checks completely on a per-application/per-process basis. Skipping checks will avoid the default 10 second waiting period entirely, as well as any other user-defined checks.

```shell
# process type specification is optional
dokku checks:skip node-js-app worker,web
```

```shell-session
-----> Skipping zero downtime for app's (node-js-app) proctypes (worker,web)
-----> Unsetting node-js-app
-----> Unsetting DOKKU_CHECKS_DISABLED
-----> Setting config vars
       DOKKU_CHECKS_SKIPPED: worker,web
```

Zero downtime checks can also be disabled completely. This will stop old containers *before* new ones start, which may result in broken connections and downtime if your application fails to boot properly.

```shell
dokku checks:disable node-js-app worker
```

```shell-session
-----> Disabling zero downtime for app's (node-js-app) proctypes (worker)
-----> Setting config vars
       DOKKU_CHECKS_DISABLED: worker
-----> Setting config vars
       DOKKU_CHECKS_SKIPPED: web
```

### Displaying checks reports for an app

!!! tip "New as of 0.8.1"

You can get a report about the app's checks status using the `checks:report` command:

```shell
dokku checks:report
```

```shell-session
=====> node-js-app checks information
       Checks disabled list: none
       Checks skipped list: none
=====> python-app checks information
       Checks disabled list: none
       Checks skipped list: none
=====> ruby-app checks information
       Checks disabled list: _all_
       Checks skipped list: none
```

You can run the command for a specific app also.

```shell
dokku checks:report node-js-app
```

```shell-session
=====> node-js-app checks information
       Checks disabled list: none
       Checks skipped list: none
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku checks:report node-js-app --checks-disabled-list
```

## Customizing checks

If your application needs a longer period to boot up - perhaps to load data into memory, or because of slow boot time - you may also use Dokku's `checks` functionality to more precisely check whether an application can serve traffic or not.

Checks are run against the detected `web` process from your application's `Procfile`. For non-web processes, Dokku will fallback to the aforementioned process uptime check.

To specify checks, add a `CHECKS` file to the root of your project directory. The `CHECKS` file should be plain text and may contain:

- check instructions
- settings (NAME=VALUE)
- comments (lines starting with #)
- empty lines

!!! info
    For Dockerfile and Docker Image based deploys, the file *must* be in the `WORKDIR` directory of the built image. `/app` is used by default as the root container directory for buildpack-based deploys.

### Check instructions

The format of a check instruction is a path or relative URL, optionally followed by the expected content:

```
/about  Our Amazing Team
```

The `CHECKS` file can contain multiple checks:

```
/                       My Amazing App
/stylesheets/index.css  .body
/scripts/index.js       $(function()
/images/logo.png
```

To check an application that supports multiple hostnames, use relative URLs that include the hostname:

```
//admin.dokku.me  Admin Dashboard
//static.dokku.me/logo.png
```

You can also specify the protocol to explicitly check HTTPS requests:

```
https://admin.dokku.me  Admin Dashboard
https://static.dokku.me/logo.png
```

While a full URL may be used in order to invoke checks, if you are using relative URLs, the port *must* be omitted.

!!! tip "Changed as of 0.22.5"

Please note that dollar sign bracket characters (`{` and `}`) must be escaped when used within a `CHECKS` file. Escaping follows golang template rules. The proper way to do this is via one of the following methods:

```
# escaping the `{` character
# using double-quotes
{{"{"}}

# using raw string constants
{{`{`}}

# escaping the `}` character
# using double-quotes
{{"}"}}

# using raw string constants
{{`}`}}
```

### Templating Checks Files

!!! tip "New as of 0.22.5"

An app's `CHECKS` file is sent through a single pass of the [`sigil`](https://github.com/gliderlabs/sigil/) templating tool. This enables usage of Golang templating within application `CHECKS` files. In addition to general templating access, access to app environment variables is also allowed via the `var` function:

```
{{ var "SOME_ENV_VAR" }}
```

This may be useful if certain zero-downtime checks require access to an app-specific value, such as a domain name.

### Check settings

The default behavior is to wait for `5` seconds before running the checks, to timeout the checks after `30` seconds, and to attempt the checks `5` times. If the checks fail `5` times, the deployment is considered failed and the old container will continue serving traffic.

You can change the default behavior by setting `WAIT`, `TIMEOUT`, and `ATTEMPTS` to different values in the `CHECKS` file:

```
WAIT=30     # Wait 1/2 minute
TIMEOUT=60  # Timeout after a minute
ATTEMPTS=10 # Attempt checks 10 times

/  My Amazing App
```

## Manually invoking checks

Checks can also be manually invoked via the `checks:run` command. This can be used to check the status of an application via cron to provide integration with external healthchecking software.

Checks are run against a specific application:

```shell
dokku checks:run APP
```

```shell-session
-----> Running pre-flight checks
-----> Running checks for app (APP.web.1)
       For more efficient zero downtime deployments, create a file CHECKS.
       See https://dokku.com/docs/deployment/zero-downtime-deploys/ for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
-----> Running checks for app (APP.web.2)
       For more efficient zero downtime deployments, create a file CHECKS.
       See https://dokku.com/docs/deployment/zero-downtime-deploys/ for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
-----> Running checks for app (APP.worker.1)
       For more efficient zero downtime deployments, create a file CHECKS.
       See https://dokku.com/docs/deployment/zero-downtime-deploys/ for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
```

Checks can be scoped to a particular process type:

```shell
dokku checks:run node-js-app worker
```

```shell-session
-----> Running pre-flight checks
-----> Running checks for app (APP.worker.1)
       For more efficient zero downtime deployments, create a file CHECKS.
       See https://dokku.com/docs/deployment/zero-downtime-deploys/ for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
```

An app process ID may also be specified:

```shell
dokku checks:run node-js-app web.2
```

```shell-session
-----> Running pre-flight checks
-----> Running checks for app (APP.web.2)
       For more efficient zero downtime deployments, create a file CHECKS.
       See https://dokku.com/docs/deployment/zero-downtime-deploys/ for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
```

Non-existent process types will result in an error:

```shell
dokku checks:run node-js-app non-existent
```

```shell-session
-----> Running pre-flight checks
Invalid process type specified (APP.non-existent)
```

Non-existent process IDs will *also* result in an error

```shell
dokku checks:run node-js-app web.3
```

```shell-session
-----> Running pre-flight checks
Invalid container id specified (APP.web.3)
```

## Example: Successful Rails deployment

In this example, a Rails application is successfully deployed to Dokku. The initial round of checks fails while the server is starting, but once it starts they succeed and the deployment is successful. `WAIT` is set to `10` because our application takes a while to boot up. `ATTEMPTS` is set to `6`, but the third attempt succeeds.

### `CHECKS` file

```
WAIT=10
ATTEMPTS=6
/check.txt  simple_check
```

For this check to work, we've added a line to `config/routes.rb` that simply returns a string:

```
get '/check.txt', to: proc {[200, {}, ['simple_check']]}
```

### Deploy output

!!! note "The output has been trimmed for brevity."

```shell
git push dokku master
```

```shell-session
-----> Cleaning up...
-----> Building node-js-app from herokuish...
-----> Adding BUILD_ENV to build environment...
-----> Ruby app detected
-----> Compiling Ruby/Rails
-----> Using Ruby version: ruby-2.0.0

.....

-----> Discovering process types
       Procfile declares types -> web
-----> Releasing node-js-app...
-----> Deploying node-js-app...
-----> Running pre-flight checks
-----> Attempt 1/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/check.txt => "simple_check"
 !
curl: (7) Failed to connect to 172.17.0.155 port 5000: Connection refused
 !    Check attempt 1/6 failed.
-----> Attempt 2/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/check.txt => "simple_check"
 !
curl: (7) Failed to connect to 172.17.0.155 port 5000: Connection refused
 !    Check attempt 2/6 failed.
-----> Attempt 3/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/check.txt => "simple_check"
-----> All checks successful!
=====> node-js-app container output:
       => Booting Thin
       => Rails 4.2.0 application starting in production on http://0.0.0.0:5000
       => Run `rails server -h` for more startup options
       => Ctrl-C to shutdown server
       Thin web server (v1.6.3 codename Protein Powder)
       Maximum connections set to 1024
       Listening on 0.0.0.0:5000, CTRL+C to stop
=====> end node-js-app container output
-----> Running post-deploy
-----> Configuring myapp.dokku.me...
-----> Creating http nginx.conf
       Reloading nginx
-----> Shutting down old container in 60 seconds
=====> Application deployed:
       http://myapp.dokku.me
```

## Example: Failing Rails deployment

In this example, a Rails application fails to deploy. The reason for the failure is that the PostgreSQL database connection fails. The initial checks will fail while we wait for the server to start up, just like in the above example. However, once the server does start accepting connections, we will see an error 500 due to the PostgreSQL database connection failure.

Once the attempts have been exceeded, the deployment fails and we see the container output, which shows the PostgreSQL connection errors.

### `CHECKS` file

```
WAIT=10
ATTEMPTS=6
/
```

!!! info "The check to the root url `/` would normally access the database."

### Deploy output

!!! note "The output has been trimmed for brevity."

```shell
git push dokku master
```

```shell-session
-----> Cleaning up...
-----> Building node-js-app from herokuish...
-----> Adding BUILD_ENV to build environment...
-----> Ruby app detected
-----> Compiling Ruby/Rails
-----> Using Ruby version: ruby-2.0.0

.....

Discovering process types
Procfile declares types -> web
Releasing node-js-app...
Deploying node-js-app...
Running pre-flight checks
-----> Attempt 1/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/ => ""
 !
curl: (7) Failed to connect to 172.17.0.188 port 5000: Connection refused
 !    Check attempt 1/6 failed.
-----> Attempt 2/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/ => ""
 !
curl: (7) Failed to connect to 172.17.0.188 port 5000: Connection refused
 !    Check attempt 2/6 failed.
-----> Attempt 3/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/ => ""
 !
curl: (22) The requested URL returned error: 500 Internal Server Error
 !    Check attempt 3/6 failed.
-----> Attempt 4/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/ => ""
 !
curl: (22) The requested URL returned error: 500 Internal Server Error
 !    Check attempt 4/6 failed.
-----> Attempt 5/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/ => ""
 !
curl: (22) The requested URL returned error: 500 Internal Server Error
 !    Check attempt 5/6 failed.
-----> Attempt 6/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/ => ""
 !
curl: (22) The requested URL returned error: 500 Internal Server Error
Could not start due to 1 failed checks.
 !    Check attempt 6/6 failed.
=====> node-js-app container output:
       => Booting Thin
       => Rails 4.2.0 application starting in production on http://0.0.0.0:5000
       => Run `rails server -h` for more startup options
       => Ctrl-C to shutdown server
       Thin web server (v1.6.3 codename Protein Powder)
       Maximum connections set to 1024
       Listening on 0.0.0.0:5000, CTRL+C to stop
       Started GET "/" for 172.17.42.1 at 2015-03-26 21:36:47 +0000
         Is the server running on host "172.17.42.1" and accepting
         TCP/IP connections on port 5431?
       PG::ConnectionBad (could not connect to server: Connection refused
         Is the server running on host "172.17.42.1" and accepting
         TCP/IP connections on port 5431?
       ):
         vendor/bundle/ruby/2.0.0/gems/activerecord-4.2.0/lib/active_record/connection_adapters/postgresql_adapter.rb:651:in `initialize'
         vendor/bundle/ruby/2.0.0/gems/activerecord-4.2.0/lib/active_record/connection_adapters/postgresql_adapter.rb:651:in `new'
         vendor/bundle/ruby/2.0.0/gems/activerecord-4.2.0/lib/active_record/connection_adapters/postgresql_adapter.rb:651:in `connect'
         vendor/bundle/ruby/2.0.0/gems/activerecord-4.2.0/lib/active_record/connection_adapters/postgresql_adapter.rb:242:in `initialize'
         vendor/bundle/ruby/2.0.0/gems/activerecord-4.2.0/lib/active_record/connection_adapters/postgresql_adapter.rb:44:in `new'
         vendor/bundle/ruby/2.0.0/gems/activerecord-4.2.0/lib/active_record/connection_adapters/postgresql_adapter.rb:44:in `postgresql_connection
=====> end node-js-app container output
/usr/bin/dokku: line 49: 23409 Killed                  dokku deploy "$APP"
To dokku@dokku.me:myapp
 ! [remote rejected] dokku -> master (pre-receive hook declined)
error: failed to push some refs to 'dokku@dokku.me:myapp'
```
