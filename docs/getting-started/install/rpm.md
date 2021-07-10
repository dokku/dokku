# RPM Package Installation Notes

> New as of 0.8.0 *(experimental)*

> **Warning:** Web installer is not available on CentOS. You will need to configure [SSH keys](/docs/deployment/user-management.md#adding-ssh-keys) and [virtual hosts](/docs/configuration/domains.md#customizing-hostnames) using dokku command line interface.

Dokku defaults to being installed via RPM package on CentOS 7. While certain hosts may require extra work to get running, you may optionally wish to automate the installation of Dokku without the use of our `bootstrap.sh` Bash script. The following are the steps run by said script:

```shell
# Install docker
curl -fsSL https://get.docker.com/ | sh

# Install epel for nginx packages to be available
sudo yum install -y epel-release

# Install dokku
curl -s https://packagecloud.io/install/repositories/dokku/dokku/script.rpm.sh | sudo bash
sudo yum install -y herokuish dokku
sudo dokku plugin:install-dependencies --core
systemctl start nginx

# Enable docker and nginx on system startup
systemctl enable docker
systemctl enable nginx
```

The [devicemapper](https://docs.docker.com/engine/userguide/storagedriver/device-mapper-driver/) is the default Docker storage driver on CentOS 7 and defaults to a configuration mode known as `loop-lvm`. This mode is designed to work out-of-the-box with no additional configuration. However, production deployments should not run under `loop-lvm` mode. The preferred configuration for production deployments is [`direct-lvm`](https://docs.docker.com/engine/userguide/storagedriver/device-mapper-driver/#/configure-direct-lvm-mode-for-production).
