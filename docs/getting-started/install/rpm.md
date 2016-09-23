# RPM Package Installation Notes

Dokku defaults to being installed via RPM package on CentOS 7. While certain hosts may require extra work to get running, you may optionally wish to automate the installation of Dokku without the use of our `bootstrap.sh` bash script. The following are the steps run by said script:

```shell
# install docker
curl -fsSL https://get.docker.com/ | sh

# install epel for nginx packages to be available
sudo yum install -y epel-release

# install dokku
curl -s https://packagecloud.io/install/repositories/dokku/dokku/script.rpm.sh | sudo bash
sudo yum install -y herokuish dokku
sudo dokku plugin:install-dependencies --core
```
