# Debian Package Installation Notes

As of 0.3.18, dokku defaults to being installed via debian package. While certain hosts may require extra work to get running, you may optionally wish to automate the installation of dokku without the use of our `bootstrap.sh` bash script. The following are the steps run by said script:

```shell
curl --silent https://get.docker.io/gpg 2> /dev/null | apt-key add - 2>&1 >/dev/null
curl --silent https://packagecloud.io/gpg.key 2> /dev/null | apt-key add - 2>&1 >/dev/null

echo "deb http://get.docker.io/ubuntu docker main" > /etc/apt/sources.list.d/docker.list
echo "deb https://packagecloud.io/dokku/dokku/ubuntu/ trusty main" > /etc/apt/sources.list.d/dokku.list

sudo apt-get update > /dev/null
sudo apt-get install -qq -y linux-image-extra-`uname -r` apt-transport-https
sudo apt-get install -qq -y dokku
```

## Unattended installation

In case you want to perform an unattended installation of dokku, this is made possible through [debconf](https://en.wikipedia.org/wiki/Debconf_%28software_package%29), which allows you to configure a package before installing it.

You can set any of the below options through the `debconf-set-selections` command, for example to enable vhost-based deployments:

```bash
echo "dokku dokku/vhost_enable boolean true" | debconf-set-selections
```

After setting the desired options, proceed with the installation as described above.

### debconf options

| Name               | Type    | Default               | Description                                                              |
| ------------------ | ------- | --------------------- | ------------------------------------------------------------------------ |
| dokku/web_config   | boolean | true                  | Use web-based config for below options                                   |
| dokku/vhost_enable | boolean | false                 | Use vhost-based deployments (e.g. <app>.dokku.me)                        |
| dokku/hostname     | string  | dokku.me              | Hostname, used as vhost domain and for showing app URL after deploy      |
| dokku/key_file     | string  | /root/.ssh/id_rsa.pub | SSH key to add to the Dokku user (Will be ignored on `dpkg-reconfigure`) |
