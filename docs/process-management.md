# Process/Container management

Dokku supports rudimentary process (really container) management via the `ps` plugin.

```
ps <app>                                        List processes running in app container(s)
ps:rebuildall                                   Rebuild all apps
ps:rebuild <app>                                Rebuild an app
ps:restartall                                   Restart all deployed app containers
ps:restart <app>                                Restart app container(s)
ps:scale <app> <proc>=<count> [<proc>=<count>]  Set how many processes of a given process to run
ps:start <app>                                  Start app container(s)
ps:stop <app>                                   Stop app container(s)
```

*NOTE*: As of v0.3.14, `dokku deploy:all` in now deprecated by `ps:restartall` and will be removed in a future version.


## Scaling

Dokku allows you to run multiple process types at different container counts. For example, if you had an app that contained 1 web app listener and 1 background job processor, dokku can, spin up 1 container for each process type defined in the Procfile. By default we will only start the web process. However, if you wanted 2 job processors running simultaneously, you can modify this behavior in a few ways.

### DOKKU_SCALE file

You can optionally create a `DOKKU_SCALE` file in the root of your repository. Dokku expects this file to contain one line for every process defined in your Procfile. Example:
```
web=1
worker=2
```

### `ps:scale` command

Dokku can also manage scaling itself via the `ps:scale` command. This command can be used to scale multiple process types at the same time.

```
dokku ps:scale app_name web=1 worker=2
```

*NOTE*: Dokku will always use the DOKKU_SCALE file that ships with the repo to override any local settings.


## The web proctype

Like Heroku, we handle the `web` proctype differently from others. The `web` proctype is the only proctype that will invoke custom checks as defined by a CHECKS file. It is also the only proctype that will be launched in a container that is either proxied via nginx or bound to an external port.

## Image tagging

The dokku tags plugin allows you to add docker image tags to the currently deployed app image for versioning and subsequent deployment.

```
tags <app>                                       List all app image tags
tags:create <app> <tag>                          Add tag to latest running app image
tags:deploy <app> <tag>                          Deploy tagged app image
tags:destroy <app> <tag>                         Remove app image tag
```

Example:
```
root@dokku:~# dokku tags node-js-app
=====> Image tags for dokku/node-js-app
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
dokku/node-js-app   latest              936a42f25901        About a minute ago   1.025 GB

root@dokku:~# dokku tags:create node-js-app v0.9.0
=====> Added v0.9.0 tag to dokku/node-js-app

root@dokku:~# dokku tags node-js-app
=====> Image tags for dokku/node-js-app
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
dokku/node-js-app   latest              936a42f25901        About a minute ago   1.025 GB
dokku/node-js-app   v0.9.0              936a42f25901        About a minute ago   1.025 GB

root@dokku:~# dokku tags:deploy node-js-app v0.9.0
-----> Releasing node-js-app (dokku/node-js-app:v0.9.0)...
-----> Deploying node-js-app (dokku/node-js-app:v0.9.0)...
-----> Running pre-flight checks
       For more efficient zero downtime deployments, create a file CHECKS.
       See http://progrium.viewdocs.io/dokku/checks-examples.md for examples
       CHECKS file not found in container: Running simple container check...
-----> Waiting for 10 seconds ...
-----> Default container check successful!
=====> node-js-app container output:
       Detected 512 MB available memory, 512 MB limit per process (WEB_MEMORY)
       Recommending WEB_CONCURRENCY=1
       > node-js-sample@0.1.0 start /app
       > node index.js
       Node app is running at localhost:5000
=====> end node-js-app container output
-----> Running post-deploy
-----> Configuring node-js-app.dokku.me...
-----> Creating http nginx.conf
-----> Running nginx-pre-reload
       Reloading nginx
-----> Shutting down old containers in 60 seconds
=====> 025eec3fa3b442fded90933d58d8ed8422901f0449f5ea0c23d00515af5d3137
=====> Application deployed:
       http://node-js-app.dokku.me

```
