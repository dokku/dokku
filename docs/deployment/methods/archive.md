# Deploying from a remote tarball or zip

## Initializing an app repository from an archive file

> [!IMPORTANT]
> New as of 0.24.0

A Dokku app repository can be initialized or updated from the contents of an archive file via the `git:from-archive` command. This is an excellent way of tracking changes when deploying pre-built binary archives, such as java jars or go binaries. This can also be useful when deploying directly from a GitHub repository at a specific commit.

```shell
dokku git:from-archive node-js-app https://github.com/dokku/smoke-test-app/releases/download/2.0.0/smoke-test-app.tar
```

In the above example, Dokku will build the app as if the repository contained the extracted contents of the specified archive file.

Triggering a build with the same archive file multiple times will result in Dokku exiting `0` early as there will be no changes detected.

The `git:from-archive` command can optionally take a git `user.name` and `user.email` argument (in that order) to customize the author. If the arguments are left empty, they will fallback to `Dokku` and `automated@dokku.sh`, respectively.

```shell
dokku git:from-archive node-js-app https://github.com/dokku/smoke-test-app/releases/download/2.0.0/smoke-test-app.tar "Camila" "camila@example.com"
```

The default archive type is always set to `.tar`. To use a different archive type, specify the `--archive-type` flag. Failure to do so will result in a failure to extract the archive.

```shell
dokku git:from-archive --archive-type zip node-js-app https://github.com/dokku/smoke-test-app/archive/2.0.0.zip "Camila" "camila@example.com"
```

Finally, if the archive url is specified as `--`, the archive will be fetched from stdin.

```shell
curl -sSL https://github.com/dokku/smoke-test-app/releases/download/2.0.0/smoke-test-app.tar | dokku git:from-archive node-js-app  --
```
