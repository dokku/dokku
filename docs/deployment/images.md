# Image Tagging

> New as of 0.4.0

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
       > node-js-app@0.1.0 start /app
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
