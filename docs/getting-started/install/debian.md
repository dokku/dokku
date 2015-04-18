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
