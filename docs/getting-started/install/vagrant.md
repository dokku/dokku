# Vagrant Installation Notes

1. Download and install [VirtualBox](https://www.virtualbox.org/wiki/Downloads).

2. Download and install [Vagrant](http://www.vagrantup.com/downloads.html).

3. Clone Dokku.

```shell
git clone https://github.com/dokku/dokku.git
```

4. Create VM.

```shell
# Optional ENV arguments:
# - `BOX_NAME`
# - `BOX_URI`
# - `BOX_MEMORY`
# - `DOKKU_DOMAIN`
# - `DOKKU_IP`
# - `FORWARDED_PORT`.
cd path/to/dokku

# for most users
vagrant up

# windows users must instead use the following in an elevated command prompt
vagrant up dokku-windows
```

5. Setup SSH Config in `~/.ssh/config`.

```ini
Host dokku.me
    Port 22
    RequestTTY yes
```

> For users that have customized the IP address of their VM - either in a custom `Vagrantfile` or via the `DOKKU_IP` environment variable - and are not using `10.0.0.2` for the Vagrant IP, you'll need to instead use the output of `vagrant ssh-config dokku` for your `~/.ssh/config` entry.

6. Register your SSH key via `cat ~/.ssh/id_rsa.pub | ssh root@dokku.me dokku ssh-keys:add main`.

7. Enable virtualhost naming by running `ssh root@dokku.me dokku domains:add-global dokku.me`.

Please note, the `dokku.me` domain is setup to point to `10.0.0.2` along with all subdomains (i.e. `yourapp.dokku.me`). If you change the `DOKKU_IP` in your Vagrant setup you'll need to update your `/etc/hosts` file to point your reconfigured IP address.

You are now ready to deploy an app or install plugins.
