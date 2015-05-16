# Checks Examples

The CHECKS file may contain empty lines, comments (lines starting with #),
settings (NAME=VALUE) and check instructions.

The format of a check instruction is a path, optionally followed by the
expected content.  For example:

	  /                       My Amazing App
	  /stylesheets/index.css  .body
	  /scripts/index.js       $(function()
	  /images/logo.png

To check an application that supports multiple hostnames, use relative URLs
that include the hostname, for example:

	 //admin.example.com     Admin Dashboard
	 //static.example.com/logo.png

You can also specify the protocol to explicitly check HTTPS requests.

The default behavior is to wait for 5 seconds before running the first check,
and timeout each check to 30 seconds.

By default, checks will be attempted 5 times.  (Retried 4 times)

You can change these by setting WAIT, TIMEOUT and ATTEMPTS to different values, for
example:

	  WAIT=30     # Wait 1/2 minute
	  TIMEOUT=60  # Timeout after a minute
	  ATTEMPTS=10  # attempt checks 10 times

# Example: Successful Rails Deployment
In this example, a rails applicaiton is successfully deployed to dokku.  The initial round of checks fails while the server is starting, but once it starts they succeed and the deployment is successful.
ATTEMPTS is set to 6, but the third attempt succeeds.

## CHECKS file

````
WAIT=5
ATTEMPTS=6
/check.txt simple_check
````

> check.txt is a text file returning the string 'simple_check'

## Deploy Output

> Note: The output has been trimmed for brevity

````
git push dokku master

-----> Cleaning up...
-----> Building myapp from buildstep...
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
-----> Attempt 1/6 Waiting for 5 seconds ...
       CHECKS expected result:
       http://localhost/check.txt => "simple_check"
 !
curl: (7) Failed to connect to 172.17.0.155 port 5000: Connection refused
 !    Check attempt 1/6 failed.  
-----> Attempt 2/6 Waiting for 5 seconds ...
       CHECKS expected result:
       http://localhost/check.txt => "simple_check"
 !
curl: (7) Failed to connect to 172.17.0.155 port 5000: Connection refused
 !    Check attempt 2/6 failed.  
-----> Attempt 3/6 Waiting for 5 seconds ...
       CHECKS expected result:
       http://localhost/check.txt => "simple_check"
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

# Example: Failing Rails Deployment
In this example, a Rails application fails to deploy.  The reason for the failure is that the postgres database connection fails.  The initial checks will fail while we wait for the server to start up, just like in the above example.  However, once the server does start accepting connections, we will see an error 500 due to the postgres database connection failure.

Once the attempts have been exceeded, the deployment fails and we see the container output, which shows the Postgres connection errors.

## CHECKS file

````
WAIT=5
ATTEMPTS=6
/
````

> The check to the root url '/' would normally access the database.

## Deploy Output

> Note: The output has been trimmed for brevity

````
git push dokku master

-----> Cleaning up...
-----> Building myapp from buildstep...
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
-----> Attempt 1/6 Waiting for 5 seconds ...
       CHECKS expected result:
       http://localhost/ => ""
 !
curl: (7) Failed to connect to 172.17.0.188 port 5000: Connection refused
 !    Check attempt 1/6 failed.  
-----> Attempt 2/6 Waiting for 5 seconds ...
       CHECKS expected result:
       http://localhost/ => ""
 !    
curl: (7) Failed to connect to 172.17.0.188 port 5000: Connection refused
 !    Check attempt 2/6 failed.  
-----> Attempt 3/6 Waiting for 5 seconds ...
       CHECKS expected result:
       http://localhost/ => ""
 !    
curl: (22) The requested URL returned error: 500 Internal Server Error
 !    Check attempt 3/6 failed.  
-----> Attempt 4/6 Waiting for 5 seconds ...
       CHECKS expected result:
       http://localhost/ => ""
 !    
curl: (22) The requested URL returned error: 500 Internal Server Error
 !    Check attempt 4/6 failed.  
-----> Attempt 5/6 Waiting for 5 seconds ...
       CHECKS expected result:
       http://localhost/ => ""
 !    
curl: (22) The requested URL returned error: 500 Internal Server Error
 !    Check attempt 5/6 failed.  
-----> Attempt 6/6 Waiting for 5 seconds ...
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
