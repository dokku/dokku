# Docker Installation Notes

Pull the dokku/dokku image:

```shell
docker pull dokku/dokku:0.22.5
```

Next, run the image.

```shell
docker run \
  --env DOKKU_HOSTNAME=dokku.me \
  --name dokku \
  --publish 3022:22 \
  --publish 8080:80 \
  --publish 8443:443 \
  --volume /var/lib/dokku:/mnt/dokku \
  --volume /var/run/docker.sock:/var/run/docker.sock \
  dokku/dokku:0.22.5
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

Application repositories, plugin config, and plugin data is persisted to disk within the specified host directory for `/var/lib/dokku`.

To install custom plugins, create a `plugin-list` file in the host's `/var/lib/dokku` directory. The plugins listed herein will be automatically installed by Dokku on container boot. This file should be the following format:

```yaml
plugin_name: repository_url
```

An example for installing the postgres and redis plugins follows:

```yaml
postgres: https://github.com/dokku/dokku-postgres.git
redis: https://github.com/dokku/dokku-redis.git
```

To initialize ssh-keys within the container, use `docker exec` to enter the container and run the appropriate ssh-keys commands.

```shell
docker exec -it dokku bash
```

Please see the [user management documentation](/docs/deployment/user-management.md) for more information.
