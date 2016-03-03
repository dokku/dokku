# Zero Downtime Deploys

By default, following a deploy, dokku will wait for `10` seconds before routing traffic to the newly created container. If the container is not running after this time, the deploy will be marked as failed and the old container will continue serving traffic.

You can modify the wait time globally or per application:

```shell
dokku config:set --global DOKKU_DEFAULT_CHECKS_WAIT=30
dokku config:set <app> DOKKU_DEFAULT_CHECKS_WAIT=30
```

## Custom Checks

You can also use dokku's `checks` functionality to more precisely determine when an application is ready to serve traffic, instead of waiting for a set time. This can be useful for speeding up your deploys, or if your application takes a variable period to boot.

To specify custom checks, add a `CHECKS` file to the root of your project directory. This is a text file with one line per check. Empty lines and lines starting with `#` are ignored.

A check is a relative URL and may be followed by expected content from the page, for example:

```
/about  Our Amazing Team
```

By default, dokku will wait `5` seconds before running the checks to give the container time to boot. This value can be overridden on a per-app basis in the `CHECKS` file by setting `WAIT=nn`. You may also override this for the global dokku installation:

```shell
dokku config:set --global DOKKU_CHECKS_WAIT=15
```

By default, dokku will retry the checks `5` times until the checks are successful. If all `5` checks fail, the deployment is considered failed. This can be overridden in the `CHECKS` file by setting `ATTEMPTS=nn`. This number is also configurable globally:

```shell
dokku config:set --global DOKKU_CHECKS_ATTEMPTS=2
```

If your application runs multiple processes (a background worker configured in a `Procfile`, for example) you may want to disable the default check wait time to avoid the `10` second default delay per process:

```shell
dokku config:set <app> DOKKU_DEFAULT_CHECKS_WAIT=0
```

## Retirement of Old Containers

By default, dokku will wait `60` seconds before stopping the old container so that existing connections are given a chance to complete. This value is also configurable globally or per application:

```shell
dokku config:set --global DOKKU_WAIT_TO_RETIRE=120
dokku config:set <app> DOKKU_WAIT_TO_RETIRE=120
```

> Note that during this time, multiple containers may be running on your server, which can be an issue for memory-hungry applications on memory-constrained servers.

## Custom Checks Examples

The CHECKS file may contain empty lines, comments (lines starting with #), settings (NAME=VALUE) and check instructions. The format of a check instruction is a path, optionally followed by the expected content. For example:

```
/                       My Amazing App
/stylesheets/index.css  .body
/scripts/index.js       $(function()
/images/logo.png
```

To check an application that supports multiple hostnames, use relative URLs that include the hostname, for example:

```
//admin.example.com            Admin Dashboard
//static.example.com/logo.png
```

You can also specify the protocol to explicitly check HTTPS requests.

```
https://admin.example.com            Admin Dashboard
https://static.example.com/logo.png
```

By default, dokku will wait `5` seconds before running the checks, retry the checks `5` times until the checks are successful, and timeout each check to `30` seconds. You can change these by setting `WAIT`, `ATTEMPTS`, and `TIMEOUT` to different values, for example:

```
WAIT=30
ATTEMPTS=10
TIMEOUT=60

/ My Amazing App
```

## Example: Successful Rails Deployment

In this example, a Rails application is successfully deployed to dokku. The initial round of checks fails while the server is starting, but it starts after 3 checks and the deployment is successful.

### CHECKS file

````
WAIT=10        # wait for 10 seconds since the app takes a while to boot
ATTEMPTS=2     # only attempt the checks twice
/checks    ok  # /checks returns a text response containing "ok"
````

### config/routes.rb

````
get '/checks', to: proc {[200, {}, ['ok']]}
````

### Deploy Output

> Note: The output has been trimmed for brevity

````
git push dokku master

-----> Cleaning up...
-----> Building myapp from herokuish...
-----> Adding BUILD_ENV to build environment...
-----> Ruby app detected
-----> Compiling Ruby/Rails
-----> Using Ruby version: ruby-2.0.0

.....

-----> Discovering process types
       Procfile declares types -> web
-----> Releasing myapp...
-----> Deploying myapp...
-----> Running pre-flight checks
-----> Attempt 1/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/checks => "ok"
 !
curl: (7) Failed to connect to 172.17.0.155 port 5000: Connection refused
 !    Check attempt 1/6 failed.
-----> Attempt 2/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/checks => "ok"
 !
curl: (7) Failed to connect to 172.17.0.155 port 5000: Connection refused
 !    Check attempt 2/6 failed.
-----> Attempt 3/6 Waiting for 10 seconds ...
       CHECKS expected result:
       http://localhost/checks => "ok"
-----> All checks successful!
=====> myapp container output:
       => Booting Thin
       => Rails 4.2.0 application starting in production on http://0.0.0.0:5000
       => Run `rails server -h` for more startup options
       => Ctrl-C to shutdown server
       Thin web server (v1.6.3 codename Protein Powder)
       Maximum connections set to 1024
       Listening on 0.0.0.0:5000, CTRL+C to stop
=====> end myapp container output
-----> Running post-deploy
-----> Configuring myapp.dokku.example.com...
-----> Creating http nginx.conf
-----> Running nginx-pre-reload
       Reloading nginx
-----> Shutting down old container in 60 seconds
=====> Application deployed:
       http://myapp.dokku.example.com
````

## Example: Failing Rails Deployment

In this example, a Rails application fails to deploy because the postgres database connection fails. The initial checks will fail while we wait for the server to start up, just like in the above example. However, once the server does start accepting connections, we will see an error 500 due to the postgres database connection failure.

Once the attempts have been exceeded, the deployment fails and we see the container output, which shows the Postgres connection errors.

### CHECKS file

````
WAIT=10
ATTEMPTS=6
/
````

> The check to the root url '/' would normally access the database.

### Deploy Output

> Note: The output has been trimmed for brevity

````
git push dokku master

-----> Cleaning up...
-----> Building myapp from herokuish...
-----> Adding BUILD_ENV to build environment...
-----> Ruby app detected
-----> Compiling Ruby/Rails
-----> Using Ruby version: ruby-2.0.0

.....

Discovering process types
Procfile declares types -> web
Releasing myapp...
Deploying myapp...
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
=====> myapp container output:
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
=====> end myapp container output
/usr/local/bin/dokku: line 49: 23409 Killed                  dokku deploy "$APP"
To dokku@dokku.example.com:myapp
 ! [remote rejected] dokku -> master (pre-receive hook declined)
error: failed to push some refs to 'dokku@dokku.example.com:myapp'
````
