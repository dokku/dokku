# Installing on Linode

Due to how Linode installs custom kernels, Linode instances *require* a reboot before they can fully work with Docker/Dokku. The Official Dokku StackScript should take care of this process for you, and will email notify you when the instance is running and you can proceed with installation.

## Using StackScript

Deploy using the following (experimental) Official StackScript:

- https://www.linode.com/stackscripts/view/11348

## Without StackScript

* Build a Ubuntu 13.04 instance

* Follow these instructions: https://www.linode.com/docs/tools-reference/custom-kernels-distros/run-a-distributionsupplied-kernel-with-pvgrub#ubuntu-1304-raring

* If `apt-get update` no longer works:

    * Verify if apt-get is trying to use ipv6 instead of ipv4 (e.g. you read something like "[Connecting to us.archive.ubuntu.com (2001:67c:1562::14)]" and apt-get would not proceed). In that case, follow these instructions: http://unix.stackexchange.com/questions/9940/convince-apt-get-not-to-use-ipv6-method (append "precedence ::ffff:0:0/96  100" to /etc/gai.conf)

    * OR: change `/etc/apt/sources.list` to one mentioned in http://mirrors.ubuntu.com/mirrors.txt

* Run the following commands:

    ```shell
    apt-get update

    apt-get install lxc wget bsdtar linux-image-extra-$(uname -r)

    modprobe aufs
    ```
* After this, you can install dokku the default way:

    ```shell
    wget -qO- https://raw.github.com/progrium/dokku/master/bootstrap.sh | sudo bash
    ```
