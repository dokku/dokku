# Buildpack Deployment

Dokku normally defaults to using [heroku buildpacks](https://devcenter.heroku.com/articles/buildpacks) for deployment, though this may be overridden by committing a valid `Dockerfile` to the root of your repository and pushing the repository to your Dokku installation. To avoid this automatic `Dockerfile` deployment detection, you may do one of the following:

- Use `dokku config:set` to set the `BUILDPACK_URL` environment variable.
- Add `BUILDPACK_URL` to a committed `.env` file in the root of your repository.
  - See the [environment variable documentation](/docs/configuration/environment-variables.md) for more details.
- Create a `.buildpacks` file in the root of your repository.

## Switching from Dockerfile deployments

If an application was previously deployed via Dockerfile, the following commands should be run before a buildpack deploy will succeed:

```shell
dokku config:unset --no-restart node-js-app DOKKU_DOCKERFILE_CMD DOKKU_DOCKERFILE_ENTRYPOINT DOKKU_PROXY_PORT_MAP
```

## Specifying a custom buildpack

In certain cases you may want to specify a custom buildpack. While Dokku uses herokuish to support all the [official heroku buildpacks](https://github.com/gliderlabs/herokuish#buildpacks), it is possible that the buildpack detection does not work well for your application. As well, you may wish to use a custom buildpack to handle specific application logic.

To use a specific buildpack, you can run the following Dokku command:

```shell
# replace REPOSITORY_URL with your buildpack's url
dokku config:set node-js-app BUILDPACK_URL=REPOSITORY_URL

# example: using a specific ruby buildpack version
dokku config:set node-js-app BUILDPACK_URL=https://github.com/heroku/heroku-buildpack-ruby.git#v142
```

Please check the documentation for your particular buildpack as you may need to include configuration files (such as a Procfile) in your project root.

## Using multiple buildpacks

You can only set a single buildpack using the `BUILDPACK_URL`, though there may be times when you wish to use multiple buildpacks. To do so, simply create a `.buildpacks` file in the base of your repository. This file should list all the buildpacks, one-per-line. For instance, if you wish to use both the `nodejs` and `ruby` buildpacks, your `.buildpacks` file should contain the following:

```shell
https://github.com/heroku/heroku-buildpack-nodejs.git#v87
https://github.com/heroku/heroku-buildpack-ruby.git#v142
```

> Always remember to pin your buildpack versions when using the multi-buildpacks method, or you may find deploys changing your deployed environment.

You may also choose to set just a single buildpack in this file, though that is up to you.

Please check the documentation for your particular buildpack(s) as you may need to include configuration files (such as a Procfile) in your project root.

## Using a specific buildpack version

As Dokku pins all buildpacks via herokuish releases, there may be occasions where a local buildpack version is out of date. If you wish to use a more recent version of the buildpack, you may use any of the above methods to specify a buildpack **without** the git sha attached like so:

```shell
# using the latest nodejs buildpack
dokku config:set node-js-app BUILDPACK_URL=https://github.com/heroku/heroku-buildpack-nodejs
```

You may also wish to use a **specific** version of a buildpack, which is also simple

```shell
# using v87 of the nodejs buildpack
dokku config:set node-js-app BUILDPACK_URL=https://github.com/heroku/heroku-buildpack-nodejs#v87
```

## Specifying commands via Procfile

While many buildpacks have a default command that is run when a detected repository is pushed, it is possible to override this command via a Procfile. A Procfile can also be used to specify multiple commands, each of which is subject to process scaling. See the [process scaling documentation](/docs/deployment/process-management.md) for more details around scaling individual processes.

A Procfile is a file named `Procfile`. It should be named `Procfile` exactly, and not anything else. For example, `Procfile.txt` is not valid. The file should be a simple text file.

The file must be placed in the root directory of your application. It will not function if placed in a subdirectory.

If the file exists, it should not be empty, as doing so may result in a failed deploy.

The syntax for declaring a `Procfile` is as follows. Note that the format is one process type per line, with no duplicate process types.

```
<process type>: <command>
```

If, for example, you have multiple queue workers and wish to scale them separately, the following would be a valid way to work around the requirement of not duplicating process types:

```Procfile
worker:           env QUEUE=* bundle exec rake resque:work
importantworker:  env QUEUE=important bundle exec rake resque:work
```

The `web` process type holds some significance in that it is the only process type that is automatically scaled to `1` on the initial application deploy. See the [process scaling documentation](/docs/deployment/process-management.md) for more details around scaling individual processes.

## Curl Build Timeouts

Certain buildpacks may time out in retrieving dependencies via curl. This can happen when your network connection is poor or if there is significant network congestion. You may see a message similar to `gzip: stdin: unexpected end of file` after a curl command.

If you see output similar this when deploying , you may need to override the curl timeouts to increase the length of time allotted to those tasks. You can do so via the `config` plugin:

```shell
dokku config:set --global CURL_TIMEOUT=1200
dokku config:set --global CURL_CONNECT_TIMEOUT=180
```

## Clearing buildpack cache

See the [repository management documentation](/docs/advanced-usage/repository-management.md#clearing-app-cache).
