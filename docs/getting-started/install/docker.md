# Docker Installation Notes

Pull the dokku/dokku image:

```shell
docker pull dokku/dokku:0.24.0
```

Next, run the image.

```shell
docker container run \
  --env DOKKU_HOSTNAME=dokku.me \
  --name dokku \
  --publish 3022:22 \
  --publish 8080:80 \
  --publish 8443:443 \
  --volume /var/lib/dokku:/mnt/dokku \
  --volume /var/run/docker.sock:/var/run/docker.sock \
  dokku/dokku:0.24.0
```

Dokku is run in the following configuration:

- The global hostname is set to `dokku.me` on boot.
- The container name is dokku.
- Container SSH port 22 is exposed on the host as 3022.
- Container HTTP port 80 is exposed on the host as 8080.
- Container HTTPS port 443 is exposed on the host as 8443.
- Data within the container is stored on the host within the `/var/lib/dokku` directory.
- The docker socket is mounted into container
- The "web installer" is not supported.

Application repositories, plugin config, as well as plugin data are persisted to disk within the specified host directory for `/var/lib/dokku`.

Other docker container options can also be used when running Dokku, though the specific outcome will depend upon the specified options. For example, the Dokku container's nginx port can be bound to a specific host ip by specifying `--publish $HOST_IP:8080:80`, where `$HOST_IP` is the IP desired. Please see the [docker container run documentation](https://docs.docker.com/engine/reference/commandline/run/) for further explanation for various docker arguments.

## Plugin Installation

To install custom plugins, create a `plugin-list` file in the host's `/var/lib/dokku` directory. The plugins listed herein will be automatically installed by Dokku on container boot. This file should be the following format:

```yaml
plugin_name: repository_url
```

An example for installing the postgres and redis plugins follows:

```yaml
postgres: https://github.com/dokku/dokku-postgres.git
redis: https://github.com/dokku/dokku-redis.git
```

## SSH Key Management

To initialize ssh-keys within the container, use `docker exec` to enter the container and run the appropriate ssh-keys commands.

```shell
docker exec -it dokku bash
```

Please see the [user management documentation](/docs/deployment/user-management.md) for more information.

## Pushing Applications

When exposing the Dokku container's SSH port (22) on 3022, something similar to the following will need to be setup within the user's `~/.ssh/config`:

```
Host dokku.docker
  HostName 127.0.0.1
  Port 3022
```

In the above example, the hostname `127.0.0.1` is being aliased to `dokku.docker`, while the port is being overriden to `3022`. All SSH commands - including git pushes - for the hostname `dokku.docker` will be transparently sent to `127.0.0.1:3022`.
