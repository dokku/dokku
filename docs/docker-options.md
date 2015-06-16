docker-options
========================

Usage
-----

```bash
$ dokku help
...
    docker-options <app>                               Display apps docker options for all phases
    docker-options <app> <phase(s)>                    Display apps docker options for phase (comma seperated phase list)
    docker-options:add <app> <phase(s)> OPTION         Add docker option to app for phase (comma seperated phase list)
    docker-options:remove <app> <phase(s)> OPTION      Remove docker option from app for phase (comma seperated phase list)
...
````

Add some options

```bash
$ dokku docker-options:add myapp deploy "-v /host/path:/container/path"
$ dokku docker-options:add myapp run "-v /another/container/path"
$ dokku docker-options:add myapp "-link container_name:alias"
```

Check what we added

```bash
$ dokku docker-options myapp
-link container_name:alias
-v /host/path:/container/path
-v /another/container/path
```

Remove an option
```bash
$ dokku docker-options:remove myapp "-link container_name:alias"
```

Advanced Usage (avoid if possible)
------------

In your applications folder (/home/dokku/app_name) create a file called DOCKER_OPTIONS.

Inside this file list one docker option per line. For example:

```bash
-link container_name:alias
-v /host/path:/container/path
-v /another/container/path
```

The above example will result in the following options being passed to docker during deploy and docker run:

```bash
-link container_name:alias -v /host/path:/container/path -v /another/container/path
```

You may also include comments (lines beginning with a #) and blank lines in the DOCKER_OPTIONS file.

Move information on docker options can be found here: http://docs.docker.io/en/latest/reference/run/ .
